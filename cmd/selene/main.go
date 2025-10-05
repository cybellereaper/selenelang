package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"

	buildwindows "github.com/cybellereaper/selenelang/internal/build/windows"
	"github.com/cybellereaper/selenelang/internal/examples"
	"github.com/cybellereaper/selenelang/internal/format"
	"github.com/cybellereaper/selenelang/internal/jit"
	"github.com/cybellereaper/selenelang/internal/lexer"
	"github.com/cybellereaper/selenelang/internal/lsp"
	"github.com/cybellereaper/selenelang/internal/project"
	"github.com/cybellereaper/selenelang/internal/runtime"
	"github.com/cybellereaper/selenelang/internal/token"
	"github.com/cybellereaper/selenelang/internal/toolchain"
	"github.com/cybellereaper/selenelang/internal/transpile"
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
	case "test":
		if err := testCommand(os.Args[2:]); err != nil {
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
	fmt.Fprintln(os.Stderr, "  test [flags]            execute all example scripts and report pass/fail status")
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
	if err := toolchain.LoadDependencies(rt, filename); err != nil {
		return err
	}
	if *jitFlag {
		program, _, err := toolchain.ParseFile(filename)
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
		program, _, err := toolchain.ParseFile(filename)
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
	return toolchain.ExecuteFile(rt, filename)
}

func testCommand(args []string) error {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	modeFlag := fs.String("mode", "all", "execution mode: interp, vm, jit, comma-separated list, or all")
	filter := fs.String("filter", "", "substring filter applied to example relative paths")
	list := fs.Bool("list", false, "list examples without executing them")
	verbose := fs.Bool("v", false, "print script output for each example")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}

	wd := mustGetwd()
	root, err := project.FindRoot(wd)
	if err != nil {
		if errors.Is(err, iofs.ErrNotExist) {
			root = wd
		} else {
			return err
		}
	}
	exampleRoots, err := examples.ManifestRoots(root)
	if err != nil {
		return err
	}
	scripts, err := examples.Discover(root, exampleRoots)
	if err != nil {
		return err
	}
	if *filter != "" {
		filtered := make([]examples.Script, 0, len(scripts))
		for _, script := range scripts {
			if strings.Contains(script.Relative, *filter) {
				filtered = append(filtered, script)
			}
		}
		scripts = filtered
	}
	if len(scripts) == 0 {
		if *filter != "" {
			return fmt.Errorf("no examples match filter %q", *filter)
		}
		return errors.New("no examples found")
	}
	if *list {
		for _, script := range scripts {
			fmt.Fprintln(os.Stdout, script.Relative)
		}
		return nil
	}
	modes, err := parseModes(*modeFlag)
	if err != nil {
		return err
	}
	var failures int
	for _, script := range scripts {
		for _, mode := range modes {
			var writer io.Writer
			var buf *bytes.Buffer
			if *verbose {
				buf = bytes.NewBuffer(nil)
				writer = buf
			} else {
				writer = io.Discard
			}
			if err := examples.Run(script, mode, writer); err != nil {
				failures++
				fmt.Fprintf(os.Stderr, "[FAIL] %s (%s): %v\n", script.Relative, mode, err)
				continue
			}
			fmt.Fprintf(os.Stdout, "[OK] %s (%s)\n", script.Relative, mode)
			if *verbose && buf.Len() > 0 {
				lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
				for _, line := range lines {
					if line == "" {
						continue
					}
					fmt.Fprintf(os.Stdout, "    %s\n", line)
				}
			}
		}
	}
	if failures > 0 {
		return fmt.Errorf("%d example(s) failed", failures)
	}
	return nil
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

func parseModes(input string) ([]examples.Mode, error) {
	if input == "" {
		return []examples.Mode{examples.ModeInterpreter}, nil
	}
	parts := strings.Split(input, ",")
	modes := make([]examples.Mode, 0, len(parts))
	seen := make(map[examples.Mode]struct{})
	add := func(mode examples.Mode) {
		if _, ok := seen[mode]; !ok {
			seen[mode] = struct{}{}
			modes = append(modes, mode)
		}
	}
	for _, part := range parts {
		switch strings.TrimSpace(strings.ToLower(part)) {
		case "", "interp", "interpreter":
			add(examples.ModeInterpreter)
		case "vm":
			add(examples.ModeVM)
		case "jit":
			add(examples.ModeJIT)
		case "all":
			add(examples.ModeInterpreter)
			add(examples.ModeVM)
			add(examples.ModeJIT)
		default:
			return nil, fmt.Errorf("unknown mode %q", part)
		}
	}
	if len(modes) == 0 {
		add(examples.ModeInterpreter)
	}
	return modes, nil
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
	root, err := projectRootOrWD()
	if err != nil {
		return err
	}
	for i := 0; i < fs.NArg(); i++ {
		filename := fs.Arg(i)
		resolved, err := resolvePathWithinRoot(root, filename)
		if err != nil {
			return err
		}
		data, err := readFileSecure(resolved)
		if err != nil {
			return err
		}
		formatted, err := format.Source(string(data))
		if err != nil {
			return fmt.Errorf("%s: %w", resolved, err)
		}
		if *write {
			if string(data) != formatted {
				if err := writeFileSecure(resolved, []byte(formatted)); err != nil {
					return err
				}
			}
			continue
		}
		if *list {
			if string(data) != formatted {
				fmt.Fprintln(os.Stdout, resolved)
			}
			continue
		}
		if fs.NArg() > 1 {
			if i > 0 {
				fmt.Fprintln(os.Stdout)
			}
			fmt.Fprintf(os.Stdout, "// %s\n", resolved)
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
	root, err := projectRootOrWD()
	if err != nil {
		return err
	}
	filename := fs.Arg(0)
	sourcePath, err := resolvePathWithinRoot(root, filename)
	if err != nil {
		return err
	}
	rt := runtime.New()
	program, source, err := toolchain.ParseFile(sourcePath)
	if err != nil {
		return err
	}
	chunk, err := rt.Compile(program)
	if err != nil {
		return err
	}
	if *windowsExe != "" {
		exePath, err := resolvePathWithinRoot(root, *windowsExe)
		if err != nil {
			return err
		}
		startDir := filepath.Dir(sourcePath)
		if err := buildwindows.BuildExecutable(startDir, filepath.Base(sourcePath), source, exePath); err != nil {
			return err
		}
	}
	listing := chunk.Disassemble()
	if *out != "" {
		outPath, err := resolvePathWithinRoot(root, *out)
		if err != nil {
			return err
		}
		return writeFileSecure(outPath, []byte(listing))
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
	root, err := projectRootOrWD()
	if err != nil {
		return err
	}
	filename := fs.Arg(0)
	resolved, err := resolvePathWithinRoot(root, filename)
	if err != nil {
		return err
	}
	program, _, err := toolchain.ParseFile(resolved)
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
		outPath, err := resolvePathWithinRoot(root, *out)
		if err != nil {
			return err
		}
		return writeFileSecure(outPath, []byte(output))
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
	srcDir, err := project.ResolveUnderRoot(cwd, "src")
	if err != nil {
		return err
	}
	if err := mkdirAllSecure(srcDir); err != nil {
		return err
	}
	examplesDir, err := project.ResolveUnderRoot(cwd, "examples")
	if err != nil {
		return err
	}
	if err := mkdirAllSecure(examplesDir); err != nil {
		return err
	}
	docsDir, err := project.ResolveUnderRoot(cwd, "docs")
	if err != nil {
		return err
	}
	if err := mkdirAllSecure(docsDir); err != nil {
		return err
	}
	mainSource := "package main;\n\nfn main() {\n    print(\"Hello from " + projectName + "!\");\n}\n"
	entryPath, err := project.ResolveUnderRoot(cwd, filepath.Join("src", "main.selene"))
	if err != nil {
		return err
	}
	if err := writeFileSecure(entryPath, []byte(mainSource)); err != nil {
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
	if err := validateLSPArgs(args); err != nil {
		return err
	}
	server := lsp.NewServer(os.Stdin, os.Stdout)
	return server.Run()
}

func validateLSPArgs(args []string) error {
	fs := flag.NewFlagSet("lsp", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	useStdio := fs.Bool("stdio", true, "communicate with the language client over stdio")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return errors.New("lsp does not accept positional arguments")
	}
	if !*useStdio {
		return errors.New("selene language server requires stdio transport")
	}
	return nil
}

func depsAdd(args []string) error {
	fs := flag.NewFlagSet("deps add", flag.ContinueOnError)
	srcPath := fs.String("path", "", "path to dependency sources (optional when using --source)")
	sourceURL := fs.String("source", "", "repository URL or module mirror used to fetch sources")
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
	fmt.Fprintf(os.Stdout, "added %s %s (checksum %s, vendor %s)\n", module, version, lockEntry.Checksum, lockEntry.Vendor)
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
		vendorPath, err := project.ResolveUnderRoot(root, locked.Vendor)
		if err != nil {
			return err
		}
		if err := project.VerifyChecksum(vendorPath, locked.Checksum); err != nil {
			return fmt.Errorf("%s: %w", module, err)
		}
	}
	fmt.Fprintln(os.Stdout, "all dependency checksums verified")
	return nil
}

func dumpTokens(filename string) error {
	root, err := projectRootOrWD()
	if err != nil {
		return err
	}
	resolved, err := resolvePathWithinRoot(root, filename)
	if err != nil {
		return err
	}
	content, err := readFileSecure(resolved)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", resolved, err)
	}
	l := lexer.New(string(content))
	for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
		fmt.Printf("%s\t%q\n", tok.Type, tok.Literal)
	}
	return nil
}

func projectRootOrWD() (string, error) {
	wd := mustGetwd()
	root, err := project.FindRoot(wd)
	if err != nil {
		if errors.Is(err, iofs.ErrNotExist) {
			return wd, nil
		}
		return "", err
	}
	return root, nil
}

func resolvePathWithinRoot(root, candidate string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absCandidate, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absRoot, absCandidate)
	if err != nil {
		return "", err
	}
	rel = filepath.Clean(rel)
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path %s escapes project root %s", absCandidate, absRoot)
	}
	return project.ResolveUnderRoot(absRoot, rel)
}

func readFileSecure(path string) ([]byte, error) {
	// #nosec G304 -- path is produced by resolvePathWithinRoot which constrains access to the project root.
	return os.ReadFile(path)
}

func writeFileSecure(path string, data []byte) error {
	return os.WriteFile(path, data, 0o600)
}

func mkdirAllSecure(path string) error {
	return os.MkdirAll(path, 0o750)
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}
