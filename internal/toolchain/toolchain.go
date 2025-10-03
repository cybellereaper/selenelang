// Package toolchain provides helpers that replicate the Selene CLI pipeline.
package toolchain

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"selenelang/internal/ast"
	"selenelang/internal/lexer"
	"selenelang/internal/parser"
	"selenelang/internal/project"
	"selenelang/internal/runtime"
)

// ParseFile reads, lexes, and parses a Selene source file into an AST program.
// It mirrors the CLI's behaviour so that other packages (tests, example runners,
// and auxiliary tooling) can reuse the same entry point without duplicating the
// lexing/parsing pipeline.
func ParseFile(filename string) (*ast.Program, string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read %s: %w", filename, err)
	}
	source := string(content)
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		return nil, "", fmt.Errorf("parse error:\n%s", strings.Join(errs, "\n"))
	}
	return program, source, nil
}

// ExecuteFile parses and runs a Selene source file within the provided runtime.
// It is a light wrapper around ParseFile and runtime.Run that ensures consistent
// error formatting across tooling entry points.
func ExecuteFile(rt *runtime.Runtime, filename string) error {
	program, _, err := ParseFile(filename)
	if err != nil {
		return err
	}
	if _, err := rt.Run(program); err != nil {
		return fmt.Errorf("runtime error: %w", err)
	}
	return nil
}

// LoadDependencies wires vendored modules recorded in selene.toml/selene.lock
// into the provided runtime so that imports work when evaluating a standalone
// entry point. The logic mirrors the CLI implementation but is exposed as a
// reusable helper for tests and additional tooling commands.
func LoadDependencies(rt *runtime.Runtime, entry string) error {
	abs, err := filepath.Abs(entry)
	if err != nil {
		return err
	}
	root, err := project.FindRoot(filepath.Dir(abs))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	manifest, err := project.LoadManifest(root)
	if err != nil {
		return err
	}
	if len(manifest.Dependencies) == 0 {
		return nil
	}
	lockfile, err := project.LoadLockfile(root)
	if err != nil {
		return err
	}
	modules := project.SortedModules(manifest.Dependencies)
	for _, module := range modules {
		dep := manifest.Dependencies[module]
		locked, ok := lockfile.Lookup(module)
		if !ok {
			return fmt.Errorf("dependency %s is not recorded in selene.lock", module)
		}
		vendorPath := filepath.Join(root, locked.Vendor)
		if err := project.VerifyChecksum(vendorPath, locked.Checksum); err != nil {
			return fmt.Errorf("%s: %w", module, err)
		}
		if err := loadVendoredModule(rt, module, vendorPath); err != nil {
			return fmt.Errorf("%s@%s: %w", module, dep.Version, err)
		}
	}
	return nil
}

func loadVendoredModule(rt *runtime.Runtime, modulePath, vendorPath string) error {
	files, err := project.ListSeleneFiles(vendorPath)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no .selene files found in %s", vendorPath)
	}
	depRuntime := runtime.New()
	for _, file := range files {
		if err := ExecuteFile(depRuntime, file); err != nil {
			return err
		}
	}
	exports := depRuntime.Environment().Snapshot()
	for _, builtin := range []string{"print", "format", "spawn", "channel", "__package__"} {
		delete(exports, builtin)
	}
	moduleVal := runtime.NewModule(lastSegment(modulePath), exports)
	attachModule(rt.Environment(), modulePath, moduleVal)
	return nil
}

func attachModule(env *runtime.Environment, modulePath string, moduleVal *runtime.Module) {
	segments := strings.Split(modulePath, "/")
	if len(segments) == 0 {
		return
	}
	current := moduleVal
	for i := len(segments) - 2; i >= 0; i-- {
		parent := runtime.NewModule(segments[i], map[string]runtime.Value{segments[i+1]: current})
		current = parent
	}
	rootName := segments[0]
	if existing, ok := env.Get(rootName); ok {
		if existingModule, ok := existing.(*runtime.Module); ok {
			mergeModules(existingModule, current)
		} else {
			env.Set(rootName, current)
		}
	} else {
		env.Set(rootName, current)
	}
	env.Set(segments[len(segments)-1], moduleVal)
}

func mergeModules(target, source *runtime.Module) {
	if target.Exports == nil {
		target.Exports = make(map[string]runtime.Value)
	}
	keys := make([]string, 0, len(source.Exports))
	for name := range source.Exports {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		val := source.Exports[name]
		if existing, ok := target.Exports[name]; ok {
			tMod, tOK := existing.(*runtime.Module)
			sMod, sOK := val.(*runtime.Module)
			if tOK && sOK {
				mergeModules(tMod, sMod)
				continue
			}
		}
		target.Exports[name] = val
	}
}

func lastSegment(path string) string {
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}
