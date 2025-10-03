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
	vendorPath := filepath.Join(root, VendorDirectory)
	if err := os.MkdirAll(vendorPath, 0o755); err != nil {
		return "", err
	}
	return vendorPath, nil
}

// CopyIntoVendor copies the contents of src into the destination directory.
func CopyIntoVendor(src, dest string) error {
	if err := os.RemoveAll(dest); err != nil {
		return err
	}
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			if rel == "." {
				return os.MkdirAll(dest, 0o755)
			}
			return os.MkdirAll(target, 0o755)
		}
		if !d.Type().IsRegular() {
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	})
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
