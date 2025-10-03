package project

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// ManifestName is the filename for Selene manifests.
	ManifestName = "selene.toml"
	// LockName is the filename for Selene dependency locks.
	LockName = "selene.lock"
	// VendorDirectory is the root directory for vendored dependencies.
	VendorDirectory = "vendor"
)

// Manifest represents the contents of a selene.toml file.
type Manifest struct {
	Project struct {
		Name    string
		Version string
		Module  string
		Entry   string
	}
	Docs struct {
		Paths []string
	}
	Examples struct {
		Roots []string
	}
	Dependencies map[string]Dependency
}

// Dependency describes a module requirement recorded in the manifest.
type Dependency struct {
	Version string
	Source  string
}

// LockedDependency captures an entry in selene.lock.
type LockedDependency struct {
	Module   string
	Version  string
	Checksum string
	Vendor   string
}

// Lockfile holds the resolved dependency set.
type Lockfile struct {
	Dependencies []LockedDependency
}

// FindRoot walks up from start to locate the nearest selene.toml manifest.
func FindRoot(start string) (string, error) {
	dir := start
	info, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		dir = filepath.Dir(dir)
	}
	for {
		candidate := filepath.Join(dir, ManifestName)
		if _, err := os.Stat(candidate); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fs.ErrNotExist
		}
		dir = parent
	}
}

// LoadManifest reads and decodes the selene.toml located at root.
func LoadManifest(root string) (*Manifest, error) {
	path := filepath.Join(root, ManifestName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	manifest := &Manifest{Dependencies: make(map[string]Dependency)}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	section := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.Trim(line, "[]")
			continue
		}
		switch section {
		case "project":
			if err := parseProjectLine(&manifest.Project, line); err != nil {
				return nil, err
			}
		case "docs":
			if err := parseArrayLine(&manifest.Docs.Paths, line); err != nil {
				return nil, err
			}
		case "examples":
			if err := parseArrayLine(&manifest.Examples.Roots, line); err != nil {
				return nil, err
			}
		case "dependencies":
			if err := parseDependencyLine(manifest.Dependencies, line); err != nil {
				return nil, err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return manifest, nil
}

func parseProjectLine(project *struct {
	Name    string
	Version string
	Module  string
	Entry   string
}, line string) error {
	key, value, ok := splitKeyValue(line)
	if !ok {
		return nil
	}
	parsed, err := parseString(value)
	if err != nil {
		return err
	}
	switch key {
	case "name":
		project.Name = parsed
	case "version":
		project.Version = parsed
	case "module":
		project.Module = parsed
	case "entry":
		project.Entry = parsed
	}
	return nil
}

func parseArrayLine(target *[]string, line string) error {
	key, value, ok := splitKeyValue(line)
	if !ok {
		return nil
	}
	if key != "paths" && key != "roots" {
		return nil
	}
	arr, err := parseStringArray(value)
	if err != nil {
		return err
	}
	*target = arr
	return nil
}

func parseDependencyLine(deps map[string]Dependency, line string) error {
	key, value, ok := splitKeyValue(line)
	if !ok {
		return nil
	}
	name, err := parseString(key)
	if err != nil {
		return fmt.Errorf("dependency keys must be quoted: %w", err)
	}
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "{") || !strings.HasSuffix(value, "}") {
		return fmt.Errorf("dependency %s must use inline table syntax", name)
	}
	inner := strings.TrimSpace(value[1 : len(value)-1])
	dep := Dependency{}
	fields := strings.Split(inner, ",")
	for _, field := range fields {
		fKey, fValue, ok := splitKeyValue(field)
		if !ok {
			continue
		}
		parsed, err := parseString(fValue)
		if err != nil {
			return err
		}
		switch strings.TrimSpace(fKey) {
		case "version":
			dep.Version = parsed
		case "source":
			dep.Source = parsed
		}
	}
	deps[name] = dep
	return nil
}

func splitKeyValue(line string) (string, string, bool) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	return key, value, true
}

func parseString(value string) (string, error) {
	value = strings.TrimSpace(value)
	if len(value) < 2 || value[0] != '"' || value[len(value)-1] != '"' {
		return "", fmt.Errorf("expected quoted string, got %q", value)
	}
	return value[1 : len(value)-1], nil
}

func parseStringArray(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]") {
		return nil, fmt.Errorf("expected array, got %q", value)
	}
	inner := strings.TrimSpace(value[1 : len(value)-1])
	if inner == "" {
		return []string{}, nil
	}
	parts := strings.Split(inner, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		s, err := parseString(part)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}

// SaveManifest writes the manifest back to selene.toml in root.
func SaveManifest(root string, manifest *Manifest) error {
	if manifest.Dependencies == nil {
		manifest.Dependencies = make(map[string]Dependency)
	}
	var buf bytes.Buffer
	buf.WriteString("[project]\n")
	fmt.Fprintf(&buf, "name = \"%s\"\n", manifest.Project.Name)
	fmt.Fprintf(&buf, "version = \"%s\"\n", manifest.Project.Version)
	if manifest.Project.Module != "" {
		fmt.Fprintf(&buf, "module = \"%s\"\n", manifest.Project.Module)
	}
	fmt.Fprintf(&buf, "entry = \"%s\"\n\n", manifest.Project.Entry)

	buf.WriteString("[docs]\n")
	writeStringArray(&buf, "paths", manifest.Docs.Paths)
	buf.WriteString("\n")

	buf.WriteString("[examples]\n")
	writeStringArray(&buf, "roots", manifest.Examples.Roots)
	buf.WriteString("\n")

	if len(manifest.Dependencies) > 0 {
		buf.WriteString("[dependencies]\n")
		modules := SortedModules(manifest.Dependencies)
		for _, module := range modules {
			dep := manifest.Dependencies[module]
			fmt.Fprintf(&buf, "\"%s\" = { version = \"%s\"", module, dep.Version)
			if dep.Source != "" {
				fmt.Fprintf(&buf, ", source = \"%s\"", dep.Source)
			}
			buf.WriteString(" }\n")
		}
	}

	path := filepath.Join(root, ManifestName)
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func writeStringArray(buf *bytes.Buffer, key string, values []string) {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = fmt.Sprintf("\"%s\"", v)
	}
	buf.WriteString(fmt.Sprintf("%s = [%s]\n", key, strings.Join(quoted, ", ")))
}

// LoadLockfile decodes selene.lock if present. Missing files are treated as empty lockfiles.
func LoadLockfile(root string) (*Lockfile, error) {
	path := filepath.Join(root, LockName)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Lockfile{}, nil
		}
		return nil, err
	}
	lock := &Lockfile{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	current := LockedDependency{}
	reading := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if line == "[[dependency]]" {
			if reading {
				lock.Dependencies = append(lock.Dependencies, current)
			}
			current = LockedDependency{}
			reading = true
			continue
		}
		key, value, ok := splitKeyValue(line)
		if !ok {
			continue
		}
		parsed, err := parseString(value)
		if err != nil {
			return nil, err
		}
		switch key {
		case "module":
			current.Module = parsed
		case "version":
			current.Version = parsed
		case "checksum":
			current.Checksum = parsed
		case "vendor":
			current.Vendor = parsed
		}
	}
	if reading {
		lock.Dependencies = append(lock.Dependencies, current)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lock, nil
}

// SaveLockfile writes selene.lock to disk.
func SaveLockfile(root string, lock *Lockfile) error {
	var buf bytes.Buffer
	for _, dep := range lock.Dependencies {
		buf.WriteString("[[dependency]]\n")
		fmt.Fprintf(&buf, "module = \"%s\"\n", dep.Module)
		fmt.Fprintf(&buf, "version = \"%s\"\n", dep.Version)
		fmt.Fprintf(&buf, "checksum = \"%s\"\n", dep.Checksum)
		fmt.Fprintf(&buf, "vendor = \"%s\"\n\n", dep.Vendor)
	}
	path := filepath.Join(root, LockName)
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// Set ensures a dependency entry is updated (or appended) in the lockfile.
func (l *Lockfile) Set(dep LockedDependency) {
	for i, existing := range l.Dependencies {
		if existing.Module == dep.Module {
			l.Dependencies[i] = dep
			return
		}
	}
	l.Dependencies = append(l.Dependencies, dep)
}

// Lookup finds a locked dependency by module path.
func (l *Lockfile) Lookup(module string) (LockedDependency, bool) {
	for _, dep := range l.Dependencies {
		if dep.Module == module {
			return dep, true
		}
	}
	return LockedDependency{}, false
}

// VendorPath returns the canonical vendor directory for a module@version pair.
func VendorPath(module, version string) string {
	sanitized := strings.ReplaceAll(version, string(filepath.Separator), "_")
	modulePath := filepath.FromSlash(module + "@" + sanitized)
	return filepath.Join(VendorDirectory, modulePath)
}

// SortedModules returns the manifest dependency keys in lexical order.
func SortedModules(deps map[string]Dependency) []string {
	modules := make([]string, 0, len(deps))
	for module := range deps {
		modules = append(modules, module)
	}
	sort.Strings(modules)
	return modules
}
