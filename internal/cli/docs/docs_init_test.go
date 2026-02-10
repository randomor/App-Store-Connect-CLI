package docs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveOutputPath_RejectsNonASCMarkdownFile(t *testing.T) {
	target := filepath.Join(t.TempDir(), "README.md")
	if err := os.WriteFile(target, []byte("# Readme\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, _, err := resolveOutputPath(target)
	if !errors.Is(err, ErrInvalidASCReferencePath) {
		t.Fatalf("expected ErrInvalidASCReferencePath, got %v", err)
	}
}

func TestResolveOutputPath_RejectsFileLikeNonMarkdownPath(t *testing.T) {
	target := filepath.Join(t.TempDir(), "notes.txt")

	_, _, err := resolveOutputPath(target)
	if !errors.Is(err, ErrInvalidASCReferencePath) {
		t.Fatalf("expected ErrInvalidASCReferencePath, got %v", err)
	}
}

func TestResolveOutputPath_DirectoryPathResolvesASCFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "docs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create directory: %v", err)
	}

	path, linkRoot, err := resolveOutputPath(dir)
	if err != nil {
		t.Fatalf("resolveOutputPath error: %v", err)
	}

	expectedPath := filepath.Join(dir, ascReferenceFile)
	if path != expectedPath {
		t.Fatalf("expected path %q, got %q", expectedPath, path)
	}
	if linkRoot != dir {
		t.Fatalf("expected link root %q, got %q", dir, linkRoot)
	}
}

func TestInitReference_ReturnsTypedErrorWhenASCExists(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("create .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, ascReferenceFile), []byte("# Existing\n"), 0o644); err != nil {
		t.Fatalf("write ASC.md: %v", err)
	}

	_, err := InitReference(InitOptions{Path: repo, Force: false, Link: false})
	if !errors.Is(err, ErrASCReferenceExists) {
		t.Fatalf("expected ErrASCReferenceExists, got %v", err)
	}
}
