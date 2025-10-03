package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureVendorTreeAndCopy(t *testing.T) {
	root := t.TempDir()
	vendorPath, err := EnsureVendorTree(root)
	if err != nil {
		t.Fatalf("EnsureVendorTree returned error: %v", err)
	}
	if info, err := os.Stat(vendorPath); err != nil || !info.IsDir() {
		t.Fatalf("expected vendor directory to exist: %v", err)
	}

	src := filepath.Join(root, "src")
	if err := os.Mkdir(src, 0o755); err != nil {
		t.Fatalf("failed to create src: %v", err)
	}
	sourceFile := filepath.Join(src, "module.selene")
	if err := os.WriteFile(sourceFile, []byte("fn main() {}"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}
	dest := filepath.Join(root, "vendor_copy")
	if err := CopyIntoVendor(src, dest); err != nil {
		t.Fatalf("CopyIntoVendor returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "module.selene")); err != nil {
		t.Fatalf("copied file missing: %v", err)
	}
}

func TestHashDirectoryAndVerifyChecksum(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.selene"), []byte("let a = 1"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("data"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	digest1, err := HashDirectory(dir)
	if err != nil {
		t.Fatalf("HashDirectory returned error: %v", err)
	}
	digest2, err := HashDirectory(dir)
	if err != nil {
		t.Fatalf("HashDirectory second call returned error: %v", err)
	}
	if digest1 != digest2 {
		t.Fatalf("expected deterministic digest, got %s and %s", digest1, digest2)
	}
	if err := VerifyChecksum(dir, digest1); err != nil {
		t.Fatalf("VerifyChecksum reported error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.selene"), []byte("let a = 2"), 0o644); err != nil {
		t.Fatalf("failed to update file: %v", err)
	}
	if err := VerifyChecksum(dir, digest1); err == nil {
		t.Fatalf("expected checksum mismatch after modification")
	}
}

func TestListSeleneFiles(t *testing.T) {
	dir := t.TempDir()
	files := []string{
		filepath.Join(dir, "a.selene"),
		filepath.Join(dir, "nested", "c.selene"),
		filepath.Join(dir, "nested", "b.txt"),
	}
	if err := os.WriteFile(files[0], []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(files[1]), 0o755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}
	if err := os.WriteFile(files[1], []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.WriteFile(files[2], []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	listed, err := ListSeleneFiles(dir)
	if err != nil {
		t.Fatalf("ListSeleneFiles returned error: %v", err)
	}
	expected := []string{files[0], files[1]}
	if len(listed) != len(expected) {
		t.Fatalf("expected %d files, got %d", len(expected), len(listed))
	}
	for i, path := range expected {
		if listed[i] != path {
			t.Fatalf("expected %s at index %d, got %s", path, i, listed[i])
		}
	}
}
