package project

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PrepareDependency vendors the module from srcPath into the project and computes its checksum.
func PrepareDependency(root, module, version, source, srcPath string) (Dependency, LockedDependency, error) {
	if module == "" {
		return Dependency{}, LockedDependency{}, errors.New("module path is required")
	}
	if version == "" {
		return Dependency{}, LockedDependency{}, errors.New("version is required")
	}
	var cleanup func()
	if srcPath == "" {
		fetched, closer, err := fetchDependencySource(module, version, source)
		if err != nil {
			return Dependency{}, LockedDependency{}, err
		}
		srcPath = fetched
		cleanup = closer
	}
	if cleanup != nil {
		defer cleanup()
	}
	absSrc, err := filepath.Abs(srcPath)
	if err != nil {
		return Dependency{}, LockedDependency{}, err
	}
	if info, err := os.Stat(absSrc); err != nil {
		return Dependency{}, LockedDependency{}, err
	} else if !info.IsDir() {
		return Dependency{}, LockedDependency{}, fmt.Errorf("%s is not a directory", absSrc)
	}
	if _, err := EnsureVendorTree(root); err != nil {
		return Dependency{}, LockedDependency{}, err
	}
	vendorRel := VendorPath(module, version)
	vendorDest := filepath.Join(root, vendorRel)
	if err := CopyIntoVendor(absSrc, vendorDest); err != nil {
		return Dependency{}, LockedDependency{}, err
	}
	checksum, err := HashDirectory(vendorDest)
	if err != nil {
		return Dependency{}, LockedDependency{}, err
	}
	if source == "" {
		source = module
	}
	dep := Dependency{Version: version, Source: source}
	lock := LockedDependency{Module: module, Version: version, Checksum: checksum, Vendor: vendorRel}
	return dep, lock, nil
}

func fetchDependencySource(module, version, source string) (string, func(), error) {
	repo := source
	if repo == "" {
		repo = module
	}
	tmpDir, err := os.MkdirTemp("", "selene-dep-*")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}
	dest := filepath.Join(tmpDir, "repo")
	cloneArgs := []string{"clone", "--depth", "1"}
	if version != "" {
		cloneArgs = append(cloneArgs, "--branch", version)
	}
	cloneArgs = append(cloneArgs, repo, dest)
	if err := runGit(cloneArgs...); err != nil {
		if version == "" {
			cleanup()
			return "", nil, err
		}
		_ = os.RemoveAll(dest)
		// Retry with a full clone followed by a checkout so tags and commit hashes work.
		if err := runGit("clone", repo, dest); err != nil {
			cleanup()
			return "", nil, err
		}
		if err := runGit("-C", dest, "checkout", version); err != nil {
			cleanup()
			return "", nil, err
		}
	}
	return dest, cleanup, nil
}

func runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
		}
		return fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return nil
}
