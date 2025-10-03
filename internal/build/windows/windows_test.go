package windows

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindModuleRootWalksParents(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/project"), 0o644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}
	nested := filepath.Join(root, "cmd", "app")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}
	got, err := findModuleRoot(nested)
	if err != nil {
		t.Fatalf("findModuleRoot returned error: %v", err)
	}
	if got != root {
		t.Fatalf("findModuleRoot = %s, want %s", got, root)
	}
}

func TestFindModuleRootErrorsWhenMissing(t *testing.T) {
	dir := t.TempDir()
	if _, err := findModuleRoot(dir); err == nil {
		t.Fatalf("expected error when go.mod is missing")
	}
}
