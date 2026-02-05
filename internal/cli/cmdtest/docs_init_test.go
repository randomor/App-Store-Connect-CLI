package cmdtest

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocsInitCreatesReferenceAndLinks(t *testing.T) {
	runInitCreatesReferenceAndLinks(t, []string{"docs", "init"})
}

func TestInitCreatesReferenceAndLinks(t *testing.T) {
	runInitCreatesReferenceAndLinks(t, []string{"init"})
}

func runInitCreatesReferenceAndLinks(t *testing.T, args []string) {
	t.Helper()
	root := RootCommand("1.2.3")

	tempDir := t.TempDir()
	repoRoot := filepath.Join(tempDir, "repo")
	subDir := filepath.Join(repoRoot, "subdir")

	if err := os.MkdirAll(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("create repo root error: %v", err)
	}
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("create subdir error: %v", err)
	}

	agentsPath := filepath.Join(repoRoot, "AGENTS.md")
	claudePath := filepath.Join(repoRoot, "CLAUDE.md")
	if err := os.WriteFile(agentsPath, []byte("# AGENTS.md\n"), 0o644); err != nil {
		t.Fatalf("write AGENTS.md error: %v", err)
	}
	if err := os.WriteFile(claudePath, []byte("@Agents.md\n"), 0o644); err != nil {
		t.Fatalf("write CLAUDE.md error: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working dir error: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWD)
	}()
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("chdir error: %v", err)
	}

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse(args); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := root.Run(context.Background()); err != nil {
			t.Fatalf("run error: %v", err)
		}
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	var payload struct {
		Path        string   `json:"path"`
		Created     bool     `json:"created"`
		Overwritten bool     `json:"overwritten"`
		Linked      []string `json:"linked"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}

	expectedPath := resolvePath(t, filepath.Join(repoRoot, "ASC.md"))
	actualPath := resolvePath(t, payload.Path)
	if actualPath != expectedPath {
		t.Fatalf("expected path %q, got %q", expectedPath, actualPath)
	}
	if !payload.Created {
		t.Fatal("expected created to be true")
	}
	if payload.Overwritten {
		t.Fatal("expected overwritten to be false")
	}

	linked := map[string]bool{}
	for _, path := range payload.Linked {
		linked[resolvePath(t, path)] = true
	}
	agentsResolved := resolvePath(t, agentsPath)
	claudeResolved := resolvePath(t, claudePath)
	if !linked[agentsResolved] || !linked[claudeResolved] {
		t.Fatalf("expected linked files to include %q and %q, got %v", agentsResolved, claudeResolved, payload.Linked)
	}

	ascData, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("read ASC.md error: %v", err)
	}
	if !strings.Contains(string(ascData), "# ASC CLI Reference") {
		t.Fatalf("expected ASC.md to contain header, got %q", string(ascData))
	}

	agentsData, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md error: %v", err)
	}
	if !strings.Contains(string(agentsData), "ASC.md") {
		t.Fatalf("expected AGENTS.md to include ASC.md reference, got %q", string(agentsData))
	}

	claudeData, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("read CLAUDE.md error: %v", err)
	}
	if !strings.Contains(string(claudeData), "@ASC.md") {
		t.Fatalf("expected CLAUDE.md to include @ASC.md, got %q", string(claudeData))
	}
}

func TestDocsInitRequiresForceToOverwrite(t *testing.T) {
	runInitRequiresForceToOverwrite(t, []string{"docs", "init"})
}

func TestInitRequiresForceToOverwrite(t *testing.T) {
	runInitRequiresForceToOverwrite(t, []string{"init"})
}

func runInitRequiresForceToOverwrite(t *testing.T, args []string) {
	t.Helper()
	root := RootCommand("1.2.3")

	tempDir := t.TempDir()
	repoRoot := filepath.Join(tempDir, "repo")
	if err := os.MkdirAll(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("create repo root error: %v", err)
	}

	ascPath := filepath.Join(repoRoot, "ASC.md")
	if err := os.WriteFile(ascPath, []byte("# Existing\n"), 0o644); err != nil {
		t.Fatalf("write ASC.md error: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working dir error: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWD)
	}()
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("chdir error: %v", err)
	}

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse(args); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		err := root.Run(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Fatalf("expected overwrite error, got %v", err)
		}
	})

	if stdout != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func resolvePath(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	return resolved
}
