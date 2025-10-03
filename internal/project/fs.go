package project

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	hasher := sha256.New()
	for _, rel := range files {
		filePath := filepath.Join(root, rel)
		f, err := os.Open(filePath)
		if err != nil {
			return "", err
		}
		if _, err := hasher.Write([]byte(strings.ReplaceAll(rel, "\\", "/"))); err != nil {
			f.Close()
			return "", err
		}
		if _, err := hasher.Write([]byte{0}); err != nil {
			f.Close()
			return "", err
		}
		data, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			return "", err
		}
		normalized := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		normalized = bytes.ReplaceAll(normalized, []byte("\r"), []byte("\n"))
		if _, err := hasher.Write(normalized); err != nil {
			return "", err
		}
	}
	return "sha256-" + hex.EncodeToString(hasher.Sum(nil)), nil
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
