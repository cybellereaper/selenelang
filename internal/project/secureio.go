package project

import "os"

// ReadFile resolves the provided path elements under root and returns the file contents.
// Paths are constrained by ResolveUnderRoot so that gosec understands the safety of the
// subsequent os.ReadFile call.
func ReadFile(root string, elements ...string) ([]byte, error) {
	resolved, err := ResolveUnderRoot(root, elements...)
	if err != nil {
		return nil, err
	}
	// #nosec G304 -- resolved path is guaranteed to stay within the provided root.
	return os.ReadFile(resolved)
}
