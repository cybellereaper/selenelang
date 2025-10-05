package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareDependencyFromPath(t *testing.T) {
	root := t.TempDir()
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "module.selene"), []byte("// test artifact\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	dep, lock, err := PrepareDependency(root, "github.com/example/dep", "v0.1.0", "https://example.com/dep", src)
	if err != nil {
		t.Fatalf("PrepareDependency returned error: %v", err)
	}

	if dep.Version != "v0.1.0" {
		t.Fatalf("unexpected dep version: %s", dep.Version)
	}
	if dep.Source != "https://example.com/dep" {
		t.Fatalf("unexpected dep source: %s", dep.Source)
	}
	if lock.Checksum == "" {
		t.Fatalf("expected checksum to be populated")
	}

	expected := filepath.Join(root, VendorPath("github.com/example/dep", "v0.1.0"))
	if _, err := os.Stat(expected); err != nil {
		t.Fatalf("vendored directory missing: %v", err)
	}
}

func TestPrepareDependencyFromGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	repoDir := t.TempDir()
	runGitCmd(t, repoDir, "init")
	runGitCmd(t, repoDir, "config", "user.email", "ci@example.com")
	runGitCmd(t, repoDir, "config", "user.name", "CI")
	filePath := filepath.Join(repoDir, "lib.selene")
	if err := os.WriteFile(filePath, []byte("// dependency fixture\n"), 0o644); err != nil {
		t.Fatalf("write repo file: %v", err)
	}
	runGitCmd(t, repoDir, "add", ".")
	runGitCmd(t, repoDir, "commit", "-m", "initial commit")
	runGitCmd(t, repoDir, "tag", "v1.0.0")

	root := t.TempDir()
	dep, lock, err := PrepareDependency(root, "github.com/example/fixture", "v1.0.0", repoDir, "")
	if err != nil {
		t.Fatalf("PrepareDependency returned error: %v", err)
	}

	if dep.Source != repoDir {
		t.Fatalf("expected source to be repo path, got %s", dep.Source)
	}
	if lock.Vendor == "" {
		t.Fatalf("expected lock vendor path to be set")
	}
	vendoredFile := filepath.Join(root, lock.Vendor, "lib.selene")
	data, err := os.ReadFile(vendoredFile)
	if err != nil {
		t.Fatalf("reading vendored file: %v", err)
	}
	if !strings.Contains(string(data), "dependency fixture") {
		t.Fatalf("vendored file missing contents: %q", data)
	}
}

func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, out)
	}
}
