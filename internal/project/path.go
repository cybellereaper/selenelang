package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveUnderRoot joins the provided path elements under root and ensures
// the resulting path does not escape the root directory. The resolved path
// is returned in absolute form to avoid ambiguity when performing safety
// checks before accessing the filesystem.
func ResolveUnderRoot(root string, elements ...string) (string, error) {
	if root == "" {
		return "", fmt.Errorf("root directory must not be empty")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	joined := filepath.Join(append([]string{absoluteRoot}, elements...)...)
	resolved, err := filepath.Abs(joined)
	if err != nil {
		return "", err
	}
	if resolved != absoluteRoot {
		prefix := absoluteRoot + string(os.PathSeparator)
		if !strings.HasPrefix(resolved, prefix) {
			return "", fmt.Errorf("path %q escapes root %q", resolved, absoluteRoot)
		}
	}
	return resolved, nil
}
