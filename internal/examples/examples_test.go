package examples_test

import (
	"errors"
	"fmt"
	"io"
	iofs "io/fs"
	"os"
	"strings"
	"testing"

	"selenelang/internal/examples"
	"selenelang/internal/project"
)

func TestExamplesRunAcrossBackends(t *testing.T) {
	scripts := discoverScripts(t)
	modes := []examples.Mode{examples.ModeInterpreter, examples.ModeVM, examples.ModeJIT}
	for _, mode := range modes {
		mode := mode
		for _, script := range scripts {
			script := script
			name := fmt.Sprintf("%s/%s", mode, strings.ReplaceAll(script.Relative, "/", "_"))
			t.Run(name, func(t *testing.T) {
				if err := examples.Run(script, mode, io.Discard); err != nil {
					t.Fatalf("%s failed: %v", script.Relative, err)
				}
			})
		}
	}
}

func discoverScripts(t *testing.T) []examples.Script {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to resolve working directory: %v", err)
	}
	root, err := project.FindRoot(wd)
	if err != nil {
		if errors.Is(err, iofs.ErrNotExist) {
			root = wd
		} else {
			t.Fatalf("failed to locate manifest: %v", err)
		}
	}
	roots, err := examples.ManifestRoots(root)
	if err != nil {
		t.Fatalf("failed to read example roots: %v", err)
	}
	scripts, err := examples.Discover(root, roots)
	if err != nil {
		t.Fatalf("failed to discover examples: %v", err)
	}
	if len(scripts) == 0 {
		t.Fatalf("no examples discovered under %v", roots)
	}
	return scripts
}
