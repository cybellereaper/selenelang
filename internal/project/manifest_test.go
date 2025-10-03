package project

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFindRootLocatesManifest(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ManifestName), []byte("[project]\nname = \"demo\"\n"), 0o644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}
	nested := filepath.Join(root, "examples", "app")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}
	start := filepath.Join(nested, "main.selene")
	if err := os.WriteFile(start, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}
	got, err := FindRoot(start)
	if err != nil {
		t.Fatalf("FindRoot returned error: %v", err)
	}
	if got != root {
		t.Fatalf("FindRoot = %s, want %s", got, root)
	}
}

func TestLoadManifestParsesSections(t *testing.T) {
	dir := t.TempDir()
	manifest := `[project]
name = "demo"
version = "1.0.0"
module = "example.com/demo"
entry = "main.selene"

[docs]
paths = ["docs"]

[examples]
roots = ["samples"]

[dependencies]
"lib/math" = { version = "0.1.0", source = "https://modules" }
`
	if err := os.WriteFile(filepath.Join(dir, ManifestName), []byte(manifest), 0o644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}
	loaded, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("LoadManifest returned error: %v", err)
	}
	if loaded.Project.Name != "demo" || loaded.Project.Version != "1.0.0" {
		t.Fatalf("unexpected project metadata: %+v", loaded.Project)
	}
	if len(loaded.Docs.Paths) != 1 || loaded.Docs.Paths[0] != "docs" {
		t.Fatalf("expected docs paths to include 'docs'")
	}
	dep, ok := loaded.Dependencies["lib/math"]
	if !ok {
		t.Fatalf("dependency lib/math not parsed")
	}
	if dep.Version != "0.1.0" || dep.Source != "https://modules" {
		t.Fatalf("unexpected dependency metadata: %+v", dep)
	}
}

func TestSaveManifestOrdersDependencies(t *testing.T) {
	dir := t.TempDir()
	manifest := &Manifest{
		Dependencies: map[string]Dependency{
			"zeta/lib":   {Version: "1.0.0"},
			"alpha/tool": {Version: "2.0.0", Source: "git"},
		},
	}
	manifest.Project.Name = "demo"
	manifest.Project.Version = "1.2.3"
	manifest.Project.Entry = "main.selene"
	if err := SaveManifest(dir, manifest); err != nil {
		t.Fatalf("SaveManifest returned error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ManifestName))
	if err != nil {
		t.Fatalf("failed to read manifest: %v", err)
	}
	contents := string(data)
	idxAlpha := strings.Index(contents, "\"alpha/tool\"")
	idxZeta := strings.Index(contents, "\"zeta/lib\"")
	if idxAlpha == -1 || idxZeta == -1 || idxAlpha > idxZeta {
		t.Fatalf("dependencies not sorted lexically:\n%s", contents)
	}
}

func TestLockfileSetAndLookup(t *testing.T) {
	lock := &Lockfile{}
	lock.Set(LockedDependency{Module: "lib/math", Version: "1.0.0"})
	lock.Set(LockedDependency{Module: "lib/math", Version: "1.1.0"})
	lock.Set(LockedDependency{Module: "core/json", Version: "0.5.0"})
	if len(lock.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(lock.Dependencies))
	}
	dep, ok := lock.Lookup("lib/math")
	if !ok || dep.Version != "1.1.0" {
		t.Fatalf("Lookup returned %+v, %v", dep, ok)
	}
}

func TestLoadLockfileParsesEntries(t *testing.T) {
	dir := t.TempDir()
	lockfile := `[[dependency]]
module = "lib/math"
version = "1.0.0"
checksum = "sha256-deadbeef"
vendor = "vendor/lib/math"

[[dependency]]
module = "core/json"
version = "0.2.0"
checksum = "sha256-feedface"
vendor = "vendor/core/json"
`
	if err := os.WriteFile(filepath.Join(dir, LockName), []byte(lockfile), 0o644); err != nil {
		t.Fatalf("failed to write lockfile: %v", err)
	}
	loaded, err := LoadLockfile(dir)
	if err != nil {
		t.Fatalf("LoadLockfile returned error: %v", err)
	}
	if len(loaded.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(loaded.Dependencies))
	}
	if dep, ok := loaded.Lookup("core/json"); !ok || dep.Version != "0.2.0" {
		t.Fatalf("Lookup returned %+v, %v", dep, ok)
	}
	emptyDir := t.TempDir()
	empty, err := LoadLockfile(emptyDir)
	if err != nil {
		t.Fatalf("LoadLockfile on missing file returned error: %v", err)
	}
	if len(empty.Dependencies) != 0 {
		t.Fatalf("expected empty lockfile, got %d entries", len(empty.Dependencies))
	}
}

func TestVendorPathNormalizesVersion(t *testing.T) {
	path := VendorPath("github.com/demo/lib", "v1.0.0-beta")
	prefix := VendorDirectory + string(filepath.Separator)
	if !strings.HasPrefix(path, prefix) {
		t.Fatalf("vendor path does not reside under %s: %s", prefix, path)
	}
	if runtime.GOOS == "windows" {
		if !strings.Contains(path, "\\") {
			t.Fatalf("expected windows path to contain backslashes: %s", path)
		}
	}
	if !strings.Contains(path, "github.com") {
		t.Fatalf("vendor path missing module name: %s", path)
	}
}
