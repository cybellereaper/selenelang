package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// PrepareDependency vendors the module from srcPath into the project and computes its checksum.
func PrepareDependency(root, module, version, source, srcPath string) (Dependency, LockedDependency, error) {
	if module == "" {
		return Dependency{}, LockedDependency{}, errors.New("module path is required")
	}
	if version == "" {
		return Dependency{}, LockedDependency{}, errors.New("version is required")
	}
	if srcPath == "" {
		return Dependency{}, LockedDependency{}, errors.New("source path is required for vendoring")
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
	dep := Dependency{Version: version, Source: source}
	lock := LockedDependency{Module: module, Version: version, Checksum: checksum, Vendor: vendorRel}
	return dep, lock, nil
}
