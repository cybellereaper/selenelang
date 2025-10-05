// Package project provides helpers for working with Selene project metadata.
package project

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/mod/sumdb/dirhash"
)

// EnsureVendorTree prepares the vendor directory under the root.
func EnsureVendorTree(root string) (string, error) {
	vendorPath, err := ResolveUnderRoot(root, VendorDirectory)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(vendorPath, 0o750); err != nil {
		return "", err
	}
	return vendorPath, nil
}

// CopyIntoVendor copies the contents of src into the destination directory.
func CopyIntoVendor(src, dest string) error {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	absDest, err := filepath.Abs(dest)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(absDest); err != nil {
		return err
	}
	return filepath.WalkDir(absSrc, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		resolvedPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(absSrc, resolvedPath)
		if err != nil {
			return err
		}
		target, err := ResolveUnderRoot(absDest, rel)
		if err != nil {
			return err
		}
		if d.IsDir() {
			if rel == "." {
				return os.MkdirAll(absDest, 0o750)
			}
			return os.MkdirAll(target, 0o750)
		}
		if !d.Type().IsRegular() {
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
			return err
		}
		return copyFile(resolvedPath, target)
	})
}

func copyFile(src, dest string) (err error) {
	// #nosec G304 -- src and dest are derived from paths constrained by ResolveUnderRoot and absolute walk roots.
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	// #nosec G304 -- dest path is sanitized via ResolveUnderRoot prior to invocation.
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

// HashDirectory produces a deterministic sha256 digest for the directory contents.
func HashDirectory(root string) (string, error) {
	sum, err := dirhash.HashDir(root, "", dirhash.Hash1)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(sum, "h1:") {
		return "", fmt.Errorf("unexpected hash prefix %q", sum)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(sum, "h1:"))
	if err != nil {
		return "", err
	}
	return "sha256-" + hex.EncodeToString(decoded), nil
}

// VerifyChecksum recomputes the directory digest and compares it to the expected checksum.
func VerifyChecksum(path, checksum string) error {
	digest, err := HashDirectory(path)
	if err != nil {
		return err
	}
	if digest != checksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", checksum, digest)
	}
	return nil
}

// ListSeleneFiles returns all .selene sources within root, sorted lexicographically.
func ListSeleneFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".selene" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}
