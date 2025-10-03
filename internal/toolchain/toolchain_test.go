package toolchain

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cybellereaper/selenelang/internal/project"
	"github.com/cybellereaper/selenelang/internal/runtime"
)

func TestLastSegment(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"":                   "",
		"module":             "module",
		"github.com/example": "example",
		"nested/path/module": "module",
	}

	for input, expect := range cases {
		got := lastSegment(input)
		if got != expect {
			t.Fatalf("lastSegment(%q) = %q, want %q", input, got, expect)
		}
	}
}

func TestAttachModuleCreatesHierarchy(t *testing.T) {
	t.Parallel()

	env := runtime.NewEnvironment()
	leaf := runtime.NewModule("util", map[string]runtime.Value{
		"exported": runtime.NewNumber(42),
	})

	attachModule(env, "selene/util", leaf)

	if _, ok := env.Get("util"); !ok {
		t.Fatalf("expected leaf module to be addressable by final segment")
	}

	rootVal, ok := env.Get("selene")
	if !ok {
		t.Fatalf("expected root namespace to be installed")
	}
	rootModule, ok := rootVal.(*runtime.Module)
	if !ok {
		t.Fatalf("expected root to be a module, got %T", rootVal)
	}

	childVal, ok := rootModule.Exports["util"]
	if !ok {
		t.Fatalf("expected root module to expose child module")
	}
	if childVal != leaf {
		t.Fatalf("expected child module to match leaf module pointer")
	}
}

func TestAttachModuleMergesExistingTree(t *testing.T) {
	t.Parallel()

	env := runtime.NewEnvironment()
	existingLeaf := runtime.NewModule("util", map[string]runtime.Value{
		"existing": runtime.NewString("value"),
	})
	attachModule(env, "selene/util", existingLeaf)

	newLeaf := runtime.NewModule("runtime", map[string]runtime.Value{
		"fresh": runtime.NewBoolean(true),
	})
	attachModule(env, "selene/runtime", newLeaf)

	rootVal, ok := env.Get("selene")
	if !ok {
		t.Fatalf("expected root namespace to be installed")
	}
	rootModule, ok := rootVal.(*runtime.Module)
	if !ok {
		t.Fatalf("expected root to be a module, got %T", rootVal)
	}

	utilVal, ok := rootModule.Exports["util"]
	if !ok {
		t.Fatalf("expected util module to remain exported")
	}
	utilModule, ok := utilVal.(*runtime.Module)
	if !ok {
		t.Fatalf("expected util export to be a module, got %T", utilVal)
	}
	if _, ok := utilModule.Exports["existing"]; !ok {
		t.Fatalf("expected existing util export to be preserved")
	}

	runtimeVal, ok := rootModule.Exports["runtime"]
	if !ok {
		t.Fatalf("expected runtime module to be exported alongside util")
	}
	runtimeModule, ok := runtimeVal.(*runtime.Module)
	if !ok {
		t.Fatalf("expected runtime export to be a module, got %T", runtimeVal)
	}
	if _, ok := runtimeModule.Exports["fresh"]; !ok {
		t.Fatalf("expected runtime module to expose new bindings")
	}

	if direct, ok := env.Get("runtime"); !ok || direct != newLeaf {
		t.Fatalf("expected direct binding for final segment to reference new module")
	}
}

func TestLoadDependenciesInstallsModules(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "selene.toml"), `
[project]
module = "example.com/app"

[dependencies]
"github.com/example/math" = { version = "v1.0.0", source = "https://example.com/math.git" }
`)

	vendorPath := filepath.Join(root, "vendor", "github.com", "example", "math@v1.0.0")
	writeFile(t, filepath.Join(vendorPath, "lib.selene"), `
let exported: Number = 42;

fn add(a: Number, b: Number): Number => a + b;
`)

	checksum, err := project.HashDirectory(vendorPath)
	if err != nil {
		t.Fatalf("failed to hash vendor directory: %v", err)
	}

	lock := fmt.Sprintf(`[[dependency]]
module = "github.com/example/math"
version = "v1.0.0"
checksum = "%s"
vendor = "vendor/github.com/example/math@v1.0.0"

`, checksum)
	writeFile(t, filepath.Join(root, "selene.lock"), lock)

	entry := filepath.Join(root, "app.selene")
	writeFile(t, entry, "// entry point placeholder\n")

	rt := runtime.New()
	if err := LoadDependencies(rt, entry); err != nil {
		t.Fatalf("LoadDependencies returned error: %v", err)
	}

	mathVal, ok := rt.Environment().Get("math")
	if !ok {
		t.Fatalf("expected dependency module to be bound by final segment")
	}
	mathModule, ok := mathVal.(*runtime.Module)
	if !ok {
		t.Fatalf("expected math binding to be a module, got %T", mathVal)
	}
	if _, ok := mathModule.Exports["exported"]; !ok {
		t.Fatalf("expected exported value in module exports")
	}
	if _, ok := mathModule.Exports["add"]; !ok {
		t.Fatalf("expected function export in module exports")
	}
	for _, builtin := range []string{"print", "format", "spawn", "channel", "__package__"} {
		if _, ok := mathModule.Exports[builtin]; ok {
			t.Fatalf("unexpected builtin %q leaked into exports", builtin)
		}
	}

	githubVal, ok := rt.Environment().Get("github.com")
	if !ok {
		t.Fatalf("expected fully-qualified module tree to be installed")
	}
	githubModule, ok := githubVal.(*runtime.Module)
	if !ok {
		t.Fatalf("expected github.com binding to be a module, got %T", githubVal)
	}
	exampleVal, ok := githubModule.Exports["example"]
	if !ok {
		t.Fatalf("expected github.com module to contain example segment")
	}
	exampleModule, ok := exampleVal.(*runtime.Module)
	if !ok {
		t.Fatalf("expected example segment to be module, got %T", exampleVal)
	}
	nestedVal, ok := exampleModule.Exports["math"]
	if !ok {
		t.Fatalf("expected nested math module in hierarchy")
	}
	if nestedVal != mathModule {
		t.Fatalf("expected nested math module to reference leaf module")
	}
}

func TestLoadDependenciesRequiresLockedEntries(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "selene.toml"), `
[project]
module = "example.com/app"

[dependencies]
"github.com/example/math" = { version = "v1.0.0", source = "https://example.com/math.git" }
`)
	writeFile(t, filepath.Join(root, "selene.lock"), "")
	entry := filepath.Join(root, "app.selene")
	writeFile(t, entry, "// entry point placeholder\n")

	rt := runtime.New()
	err := LoadDependencies(rt, entry)
	if err == nil {
		t.Fatalf("expected error when lockfile is missing dependency entry")
	}
	if got, want := err.Error(), "dependency github.com/example/math is not recorded in selene.lock"; got != want {
		t.Fatalf("unexpected error: got %q, want %q", got, want)
	}
}

func TestLoadDependenciesRequiresSeleneSources(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "selene.toml"), `
[project]
module = "example.com/app"

[dependencies]
"github.com/example/math" = { version = "v1.0.0", source = "https://example.com/math.git" }
`)

	vendorPath := filepath.Join(root, "vendor", "github.com", "example", "math@v1.0.0")
	if err := os.MkdirAll(vendorPath, 0o755); err != nil {
		t.Fatalf("failed to create vendor dir: %v", err)
	}
	checksum, err := project.HashDirectory(vendorPath)
	if err != nil {
		t.Fatalf("failed to hash vendor directory: %v", err)
	}
	lock := fmt.Sprintf(`[[dependency]]
module = "github.com/example/math"
version = "v1.0.0"
checksum = "%s"
vendor = "vendor/github.com/example/math@v1.0.0"

`, checksum)
	writeFile(t, filepath.Join(root, "selene.lock"), lock)
	entry := filepath.Join(root, "app.selene")
	writeFile(t, entry, "// entry point placeholder\n")

	rt := runtime.New()
	err = LoadDependencies(rt, entry)
	if err == nil {
		t.Fatalf("expected error when vendored module is empty")
	}
	if got, want := err.Error(), fmt.Sprintf("github.com/example/math@v1.0.0: no .selene files found in %s", vendorPath); got != want {
		t.Fatalf("unexpected error: got %q, want %q", got, want)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create directory %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}
