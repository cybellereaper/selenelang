// Package examples discovers and executes Selene example scripts.
package examples

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/cybellereaper/selenelang/internal/jit"
	"github.com/cybellereaper/selenelang/internal/project"
	"github.com/cybellereaper/selenelang/internal/runtime"
	"github.com/cybellereaper/selenelang/internal/toolchain"
)

// Mode identifies which execution backend should be used when running an example.
// It intentionally mirrors the CLI surface so the command and tests stay in sync.
type Mode string

const (
	ModeInterpreter Mode = "interp"
	ModeVM          Mode = "vm"
	ModeJIT         Mode = "jit"
)

// Script represents a runnable example discovered on disk.
type Script struct {
	Path     string
	Relative string
}

// Discover walks the provided example roots (relative to the repository root)
// and returns a stable, deduplicated list of runnable scripts.
func Discover(root string, roots []string) ([]Script, error) {
	if len(roots) == 0 {
		roots = []string{"examples"}
	}
	seen := make(map[string]struct{})
	scripts := make([]Script, 0)
	for _, entry := range roots {
		base := entry
		if !filepath.IsAbs(base) {
			base = filepath.Join(root, entry)
		}
		if err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					return nil
				}
				return err
			}
			if d.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".selene" {
				return nil
			}
			if _, ok := seen[path]; ok {
				return nil
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				rel = path
			}
			seen[path] = struct{}{}
			scripts = append(scripts, Script{Path: path, Relative: filepath.ToSlash(rel)})
			return nil
		}); err != nil {
			return nil, err
		}
	}
	slices.SortFunc(scripts, func(a, b Script) int {
		return cmp.Compare(a.Relative, b.Relative)
	})
	return scripts, nil
}

// Run executes a script using the selected mode. Output produced through the
// builtin `print` function is redirected to the provided writer when non-nil.
func Run(script Script, mode Mode, stdout io.Writer) error {
	program, _, err := toolchain.ParseFile(script.Path)
	if err != nil {
		return err
	}
	rt := runtime.New()
	if stdout != nil {
		rt.Environment().Set("print", runtime.NewBuiltin("print", func(args []runtime.Value) (runtime.Value, error) {
			for i, arg := range args {
				if i > 0 {
					if _, err := io.WriteString(stdout, " "); err != nil {
						return nil, err
					}
				}
				if _, err := io.WriteString(stdout, arg.Inspect()); err != nil {
					return nil, err
				}
			}
			if _, err := io.WriteString(stdout, "\n"); err != nil {
				return nil, err
			}
			return runtime.NullValue, nil
		}))
	}
	if err := toolchain.LoadDependencies(rt, script.Path); err != nil {
		return err
	}
	switch mode {
	case ModeInterpreter:
		_, err = rt.Run(program)
	case ModeVM:
		chunk, cerr := rt.Compile(program)
		if cerr != nil {
			return cerr
		}
		_, err = rt.RunChunk(chunk)
	case ModeJIT:
		compiled, cerr := jit.Compile(program)
		if cerr != nil {
			return cerr
		}
		_, err = compiled.Run(rt)
	default:
		err = fmt.Errorf("unknown execution mode %q", mode)
	}
	return err
}

// RunAll executes each script using every requested mode. It returns a slice of
// accumulated errors (rather than failing fast) so tooling can report the full
// set of failing examples.
func RunAll(scripts []Script, modes []Mode) []error {
	if len(modes) == 0 {
		modes = []Mode{ModeInterpreter}
	}
	errs := make([]error, 0, len(scripts)*len(modes))
	for _, script := range scripts {
		for _, mode := range modes {
			if err := Run(script, mode, io.Discard); err != nil {
				errs = append(errs, fmt.Errorf("%s [%s]: %w", script.Relative, mode, err))
			}
		}
	}
	return errs
}

// ManifestRoots returns the configured example roots (falling back to a default
// of `examples/` when the manifest omits the section).
func ManifestRoots(root string) ([]string, error) {
	manifest, err := project.LoadManifest(root)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) || errors.Is(err, os.ErrNotExist) {
			return []string{"examples"}, nil
		}
		return nil, err
	}
	if len(manifest.Examples.Roots) == 0 {
		return []string{"examples"}, nil
	}
	return slices.Clone(manifest.Examples.Roots), nil
}

// Capture executes the script using the interpreter and returns everything the
// program printed. It is a convenience helper for documentation tooling.
func Capture(script Script) (string, error) {
	buf := bytes.NewBuffer(nil)
	if err := Run(script, ModeInterpreter, buf); err != nil {
		return "", err
	}
	return strings.TrimRight(buf.String(), "\n"), nil
}
