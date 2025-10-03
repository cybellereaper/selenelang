package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"selenelang/internal/ast"
	buildwindows "selenelang/internal/build/windows"
	"selenelang/internal/format"
	"selenelang/internal/jit"
	"selenelang/internal/lexer"
	"selenelang/internal/lsp"
	"selenelang/internal/parser"
	"selenelang/internal/project"
	"selenelang/internal/runtime"
	"selenelang/internal/token"
	"selenelang/internal/transpile"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		if err := runCommand(os.Args[2:]); err != nil {
			exitWithError(err)
		}
	case "tokens":
		if err := tokensCommand(os.Args[2:]); err != nil {
			exitWithError(err)
		}
	case "init":
		if err := initCommand(os.Args[2:]); err != nil {
			exitWithError(err)
		}
	case "deps":
		if err := depsCommand(os.Args[2:]); err != nil {
			exitWithError(err)
		}
	case "lsp":
		if err := lspCommand(os.Args[2:]); err != nil {
			exitWithError(err)
		}
	case "fmt":
		if err := fmtCommand(os.Args[2:]); err != nil {
			exitWithError(err)
		}
	case "build":
		if err := buildCommand(os.Args[2:]); err != nil {
			exitWithError(err)
		}
	case "transpile":
		if err := transpileCommand(os.Args[2:]); err != nil {
			exitWithError(err)
		}
	default:
		if err := runCommand(os.Args[1:]); err != nil {
			exitWithError(err)
		}
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: selene <command> [options]")
	fmt.Fprintln(os.Stderr, "commands:")
	fmt.Fprintln(os.Stderr, "  run [--tokens|--vm|--jit] <file> execute a Selene source file")
	fmt.Fprintln(os.Stderr, "  tokens <file>           dump the token stream for a file")
	fmt.Fprintln(os.Stderr, "  init <module> [--name]  create a new Selene project")
	fmt.Fprintln(os.Stderr, "  deps <subcommand>       manage project dependencies (add, list, verify)")
	fmt.Fprintln(os.Stderr, "  lsp                    start the Selene language server on stdio")
	fmt.Fprintln(os.Stderr, "  fmt [flags] <files>    format Selene source files")
	fmt.Fprintln(os.Stderr, "  build [--out|--windows-exe] <file>   compile Selene bytecode, emit listings, or build Windows executables")
	fmt.Fprintln(os.Stderr, "  transpile [flags] <file>  convert Selene sources to another language")
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func runCommand(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	tokensFlag := fs.Bool("tokens", false, "print the token stream")
	vmFlag := fs.Bool("vm", false, "execute using the Selene virtual machine")
	jitFlag := fs.Bool("jit", false, "execute using the Selene JIT engine")
	disFlag := fs.Bool("disassemble", false, "dump bytecode before executing with --vm")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return errors.New("run requires a source file")
	}
	filename := fs.Arg(0)
	if *tokensFlag {
		return dumpTokens(filename)
	}
	rt := runtime.New()
	if err := loadDependencies(rt, filename); err != nil {
		return err
	}
	if *jitFlag {
		program, _, err := parseFile(filename)
		if err != nil {
			return err
		}
		compiled, err := jit.Compile(program)
		if err != nil {
			return err
		}
		if _, err := compiled.Run(rt); err != nil {
			return fmt.Errorf("jit error: %w", err)
		}
		return nil
	}
	if *vmFlag {
		program, _, err := parseFile(filename)
		if err != nil {
			return err
		}
		chunk, err := rt.Compile(program)
		if err != nil {
			return err
		}
		if *disFlag {
			fmt.Println(chunk.Disassemble())
		}
		if _, err := rt.RunChunk(chunk); err != nil {
			return fmt.Errorf("vm error: %w", err)
		}
		return nil
	}
	return executeFile(rt, filename)
}

func tokensCommand(args []string) error {
	fs := flag.NewFlagSet("tokens", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return errors.New("tokens requires a source file")
	}
	return dumpTokens(fs.Arg(0))
}

func fmtCommand(args []string) error {
	fs := flag.NewFlagSet("fmt", flag.ContinueOnError)
	write := fs.Bool("w", false, "write result to file instead of stdout")
	list := fs.Bool("l", false, "list files whose formatting differs")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return errors.New("fmt requires at least one file")
	}
	for i := 0; i < fs.NArg(); i++ {
		filename := fs.Arg(i)
		data, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		formatted, err := format.Source(string(data))
		if err != nil {
			return fmt.Errorf("%s: %w", filename, err)
		}
		if *write {
			if string(data) != formatted {
				if err := os.WriteFile(filename, []byte(formatted), 0o644); err != nil {
					return err
				}
			}
			continue
		}
		if *list {
			if string(data) != formatted {
				fmt.Fprintln(os.Stdout, filename)
			}
			continue
		}
		if fs.NArg() > 1 {
			if i > 0 {
				fmt.Fprintln(os.Stdout)
			}
			fmt.Fprintf(os.Stdout, "// %s\n", filename)
		}
		fmt.Fprint(os.Stdout, formatted)
	}
	return nil
}

func buildCommand(args []string) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	out := fs.String("out", "", "write bytecode listing to the provided file")
	windowsExe := fs.String("windows-exe", "", "produce a Windows executable that runs via the JIT engine")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return errors.New("build requires a source file")
	}
	filename := fs.Arg(0)
	rt := runtime.New()
	program, source, err := parseFile(filename)
	if err != nil {
		return err
	}
	chunk, err := rt.Compile(program)
	if err != nil {
		return err
	}
	if *windowsExe != "" {
		abs, err := filepath.Abs(filename)
		if err != nil {
			return err
		}
		startDir := filepath.Dir(abs)
		if err := buildwindows.BuildExecutable(startDir, filepath.Base(filename), source, *windowsExe); err != nil {
			return err
		}
	}
	listing := chunk.Disassemble()
	if *out != "" {
		return os.WriteFile(*out, []byte(listing), 0o644)
	}
	fmt.Print(listing)
	return nil
}

func transpileCommand(args []string) error {
	fs := flag.NewFlagSet("transpile", flag.ContinueOnError)
	lang := fs.String("lang", "go", "target language for transpilation")
	out := fs.String("out", "", "write transpiled source to file")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return errors.New("transpile requires a source file")
	}
	filename := fs.Arg(0)
	program, _, err := parseFile(filename)
	if err != nil {
		return err
	}
	var output string
	switch strings.ToLower(*lang) {
	case "go":
		output, err = transpile.ToGo(program)
	default:
		return fmt.Errorf("unsupported target language %q", *lang)
	}
	if err != nil {
		return err
	}
	if *out != "" {
		return os.WriteFile(*out, []byte(output), 0o644)
	}
	fmt.Print(output)
	return nil
}

func initCommand(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	nameFlag := fs.String("name", "", "project name to record in the manifest")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return errors.New("init requires a module path (e.g. github.com/user/project)")
	}
	modulePath := fs.Arg(0)
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(cwd, project.ManifestName)); err == nil {
		return fmt.Errorf("selene.toml already exists in %s", cwd)
	}
	projectName := *nameFlag
	if projectName == "" {
		parts := strings.Split(modulePath, "/")
		projectName = parts[len(parts)-1]
	}
	manifest := &project.Manifest{}
	manifest.Project.Name = projectName
	manifest.Project.Version = "0.1.0"
	manifest.Project.Module = modulePath
	manifest.Project.Entry = filepath.ToSlash(filepath.Join("src", "main.selene"))
	manifest.Docs.Paths = []string{"docs", "README.md"}
	manifest.Examples.Roots = []string{"examples"}
	manifest.Dependencies = make(map[string]project.Dependency)
	if err := project.SaveManifest(cwd, manifest); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(cwd, "src"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(cwd, "examples"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(cwd, "docs"), 0o755); err != nil {
		return err
	}
	mainSource := "package main;\n\nfn main() {\n    print(\"Hello from " + projectName + "!\");\n}\n\nmain();\n"
	entryPath := filepath.Join(cwd, "src", "main.selene")
	if err := os.WriteFile(entryPath, []byte(mainSource), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "initialized Selene module %s\n", modulePath)
	return nil
}

func depsCommand(args []string) error {
	if len(args) == 0 {
		return errors.New("deps requires a subcommand: add, list, verify")
	}
	switch args[0] {
	case "add":
		return depsAdd(args[1:])
	case "list":
		return depsList(args[1:])
	case "verify":
		return depsVerify(args[1:])
	default:
		return fmt.Errorf("unknown deps subcommand %q", args[0])
	}
}

func lspCommand(args []string) error {
	if len(args) > 0 {
		return errors.New("lsp does not accept positional arguments")
	}
	server := lsp.NewServer(os.Stdin, os.Stdout)
	return server.Run()
}

func depsAdd(args []string) error {
	fs := flag.NewFlagSet("deps add", flag.ContinueOnError)
	srcPath := fs.String("path", "", "path to the dependency sources to vendor")
	sourceURL := fs.String("source", "", "canonical repository URL for auditing")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 2 {
		return errors.New("deps add requires a module path and version")
	}
	module := fs.Arg(0)
	version := fs.Arg(1)
	if *sourceURL == "" {
		*sourceURL = module
	}
	root, err := project.FindRoot(mustGetwd())
	if err != nil {
		return fmt.Errorf("cannot locate selene.toml: %w", err)
	}
	manifest, err := project.LoadManifest(root)
	if err != nil {
		return err
	}
	dep, lockEntry, err := project.PrepareDependency(root, module, version, *sourceURL, *srcPath)
	if err != nil {
		return err
	}
	manifest.Dependencies[module] = dep
	if err := project.SaveManifest(root, manifest); err != nil {
		return err
	}
	lockfile, err := project.LoadLockfile(root)
	if err != nil {
		return err
	}
	lockfile.Set(lockEntry)
	if err := project.SaveLockfile(root, lockfile); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "added %s %s (checksum %s)\n", module, version, lockEntry.Checksum)
	return nil
}

func depsList(args []string) error {
	if len(args) != 0 {
		return errors.New("deps list does not take additional arguments")
	}
	root, err := project.FindRoot(mustGetwd())
	if err != nil {
		return fmt.Errorf("cannot locate selene.toml: %w", err)
	}
	manifest, err := project.LoadManifest(root)
	if err != nil {
		return err
	}
	lockfile, err := project.LoadLockfile(root)
	if err != nil {
		return err
	}
	modules := project.SortedModules(manifest.Dependencies)
	if len(modules) == 0 {
		fmt.Fprintln(os.Stdout, "(no dependencies)")
		return nil
	}
	fmt.Fprintf(os.Stdout, "MODULE\tVERSION\tSOURCE\tCHECKSUM\n")
	for _, module := range modules {
		dep := manifest.Dependencies[module]
		locked, _ := lockfile.Lookup(module)
		fmt.Fprintf(os.Stdout, "%s\t%s\t%s\t%s\n", module, dep.Version, dep.Source, locked.Checksum)
	}
	return nil
}

func depsVerify(args []string) error {
	if len(args) != 0 {
		return errors.New("deps verify does not take additional arguments")
	}
	root, err := project.FindRoot(mustGetwd())
	if err != nil {
		return fmt.Errorf("cannot locate selene.toml: %w", err)
	}
	manifest, err := project.LoadManifest(root)
	if err != nil {
		return err
	}
	lockfile, err := project.LoadLockfile(root)
	if err != nil {
		return err
	}
	modules := project.SortedModules(manifest.Dependencies)
	for _, module := range modules {
		locked, ok := lockfile.Lookup(module)
		if !ok {
			return fmt.Errorf("dependency %s is missing from selene.lock", module)
		}
		vendorPath := filepath.Join(root, locked.Vendor)
		if err := project.VerifyChecksum(vendorPath, locked.Checksum); err != nil {
			return fmt.Errorf("%s: %w", module, err)
		}
	}
	fmt.Fprintln(os.Stdout, "all dependency checksums verified")
	return nil
}

func dumpTokens(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filename, err)
	}
	l := lexer.New(string(content))
	for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
		fmt.Printf("%s\t%q\n", tok.Type, tok.Literal)
	}
	return nil
}

func parseFile(filename string) (*ast.Program, string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read %s: %w", filename, err)
	}
	source := string(content)
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		return nil, "", fmt.Errorf("parse error:\n%s", strings.Join(errs, "\n"))
	}
	return program, source, nil
}

func executeFile(rt *runtime.Runtime, filename string) error {
	program, _, err := parseFile(filename)
	if err != nil {
		return err
	}
	if _, err := rt.Run(program); err != nil {
		return fmt.Errorf("runtime error: %w", err)
	}
	return nil
}

func loadDependencies(rt *runtime.Runtime, entry string) error {
	abs, err := filepath.Abs(entry)
	if err != nil {
		return err
	}
	root, err := project.FindRoot(filepath.Dir(abs))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	manifest, err := project.LoadManifest(root)
	if err != nil {
		return err
	}
	if len(manifest.Dependencies) == 0 {
		return nil
	}
	lockfile, err := project.LoadLockfile(root)
	if err != nil {
		return err
	}
	modules := project.SortedModules(manifest.Dependencies)
	for _, module := range modules {
		dep := manifest.Dependencies[module]
		locked, ok := lockfile.Lookup(module)
		if !ok {
			return fmt.Errorf("dependency %s is not recorded in selene.lock", module)
		}
		vendorPath := filepath.Join(root, locked.Vendor)
		if err := project.VerifyChecksum(vendorPath, locked.Checksum); err != nil {
			return fmt.Errorf("%s: %w", module, err)
		}
		if err := loadVendoredModule(rt, module, vendorPath); err != nil {
			return fmt.Errorf("%s@%s: %w", module, dep.Version, err)
		}
	}
	return nil
}

func loadVendoredModule(rt *runtime.Runtime, modulePath, vendorPath string) error {
	files, err := project.ListSeleneFiles(vendorPath)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no .selene files found in %s", vendorPath)
	}
	depRuntime := runtime.New()
	for _, file := range files {
		if err := executeFile(depRuntime, file); err != nil {
			return err
		}
	}
	exports := depRuntime.Environment().Snapshot()
	for _, builtin := range []string{"print", "format", "spawn", "channel", "__package__"} {
		delete(exports, builtin)
	}
	moduleVal := runtime.NewModule(lastSegment(modulePath), exports)
	attachModule(rt.Environment(), modulePath, moduleVal)
	return nil
}

func attachModule(env *runtime.Environment, modulePath string, moduleVal *runtime.Module) {
	segments := strings.Split(modulePath, "/")
	if len(segments) == 0 {
		return
	}
	current := moduleVal
	for i := len(segments) - 2; i >= 0; i-- {
		parent := runtime.NewModule(segments[i], map[string]runtime.Value{segments[i+1]: current})
		current = parent
	}
	rootName := segments[0]
	if existing, ok := env.Get(rootName); ok {
		if existingModule, ok := existing.(*runtime.Module); ok {
			mergeModules(existingModule, current)
		} else {
			env.Set(rootName, current)
		}
	} else {
		env.Set(rootName, current)
	}
	env.Set(segments[len(segments)-1], moduleVal)
}

func mergeModules(target, source *runtime.Module) {
	if target.Exports == nil {
		target.Exports = make(map[string]runtime.Value)
	}
	keys := make([]string, 0, len(source.Exports))
	for name := range source.Exports {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		val := source.Exports[name]
		if existing, ok := target.Exports[name]; ok {
			tMod, tOK := existing.(*runtime.Module)
			sMod, sOK := val.(*runtime.Module)
			if tOK && sOK {
				mergeModules(tMod, sMod)
				continue
			}
		}
		target.Exports[name] = val
	}
}

func lastSegment(path string) string {
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}
