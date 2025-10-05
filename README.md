<p align="center">
  <img src="https://github.com/cybellereaper/selenelang/blob/main/assets/logo-design-1.png" width="512" alt="Moon orbit animation" />
</p> 

<!-- <h1 align="center">Selene Language Toolkit</h1> -->

<p align="center">
  <a href="https://github.com/cybellereaper/selenelang/actions/workflows/go-tests.yml"><img src="https://github.com/cybellereaper/selenelang/actions/workflows/go-tests.yml/badge.svg" alt="Go Tests"></a>
  <a href="https://pkg.go.dev/github.com/cybellereaper/selenelang"><img src="https://pkg.go.dev/badge/github.com/cybellereaper/selenelang.svg" alt="Go Reference"></a>
</p>

## Table of contents

1. [Orbit overview](#orbit-overview)
2. [Quick launch](#quick-launch)
3. [CLI star chart](#cli-star-chart)
4. [Example nebula](#example-nebula)
5. [Documentation constellations](#documentation-constellations)
6. [Embedding rocket fuel](#embedding-rocket-fuel)
7. [Repository map](#repository-map)
8. [Contributing meteors](#contributing-meteors)

---

## Orbit overview

Selene is an experimental programming-language frontend written in Go. It packages the full toolchain needed to prototype new language ideas: a lexer, a Pratt parser that produces rich abstract syntax trees (ASTs), an interpreter-grade runtime, a bytecode compiler and virtual machine, an opinionated formatter, and an ergonomic CLI. Use Selene to iterate on language design, embed the runtime inside Go applications, or publish tooling such as language servers and transpilers.

<!-- <div align="center">
  <img src="https://media.giphy.com/media/v1.Y2lkPTc5MGI3NjExazllM2JwOW5jM2lmMzc5a2FnaGcxZHBua2p5cmNwYm9sOWp2NmZuYSZlcD12MV9naWZzX3NlYXJjaCZjdD1n/3oEjI6SIIHBdRxXI40/giphy.gif" width="320" alt="Stars swirling animation" />
</div> -->

## Quick launch

> 🚀 **Launch checklist:** Go 1.25.1+, a POSIX shell, and curiosity.

### Download a release build

Prebuilt archives live on the [GitHub Releases](https://github.com/cybellereaper/selenelang/releases) page for Linux, macOS, and Windows.
Grab the archive that matches your platform, unpack it, and place the `selene` (or `selene.exe`) binary somewhere on your `PATH`:

```bash
# Linux / macOS
tar -xzf selene-linux-amd64.tar.gz
sudo install selene-linux-amd64/selene /usr/local/bin/selene

# Windows (PowerShell)
Expand-Archive selene-windows-amd64.zip
Move-Item selene-windows-amd64\selene.exe "C:\\Program Files\\Selene\\selene.exe"
```

### Build from source

```bash
git clone https://github.com/cybellereaper/selenelang.git
cd selenelang
go mod tidy
go build ./...
go install ./cmd/selene
```

Fire up the CLI to verify either install path:

```bash
selene --help
```

### First flight

Create a simple program in `examples/fundamentals/hello.selene`:

```selene
// examples/fundamentals/hello.selene
let greeting: String = "Hello";

fn greet(name: String): String => greeting + ", " + name;

fn main() {
    print(greet("Selene"));
}
```

Run it through each propulsion system:

```bash
selene run examples/fundamentals/hello.selene
selene run --vm examples/fundamentals/hello.selene
selene run --jit examples/fundamentals/hello.selene
```

### Formatter + tests

Polish every script and exercise the curated gallery:

```bash
selene fmt -w examples
selene test --mode all --verbose
```

## CLI star chart

| Command | Purpose |
| --- | --- |
| `selene run <file>` | Interpret a script directly, or add `--vm` / `--jit` for alternate backends. |
| `selene tokens <file>` | Print the token stream emitted by the lexer. |
| `selene fmt [-w] <path>` | Format Selene sources in place or to STDOUT. |
| `selene build --out <file> <input>` | Compile a script to bytecode and write the chunk to disk. |
| `selene transpile --lang go --out <file> <input>` | Generate Go scaffolding for the given Selene module. |
| `selene test --mode all --verbose` | Execute curated examples through the interpreter, VM, and JIT pipelines. |
| `selene deps add/list/verify` | Manage vendored dependencies with cryptographic checksums. |
| `selene init <module>` | Scaffold a new workspace with a manifest, documentation skeleton, and starter source file. |
| `selene lsp` | Launch the Language Server Protocol endpoint used by editors and the VS Code extension. |

### Dependency management upgrades

The `deps add` subcommand can now source code directly from a Git repository, making it much easier to vendor Selene packages without crafting a staging directory first. Provide a module path and version, and (optionally) a `--source` URL if it differs from the module identifier:

```bash
selene deps add github.com/selene-lang/richmath v1.0.0 --source https://github.com/selene-lang/richmath.git
```

Selene will clone the tagged release into `vendor/`, compute the checksum, and update both `selene.toml` and `selene.lock`. You can still point `--path` at local sources when working offline—the flag remains available for advanced workflows.

## Example nebula

The `examples/` directory is now organized by theme so you can warp directly to the scenario you need:

```
examples/
  fundamentals/     Language basics (hello world, math, strings, flow control)
  modularity/       Modules, packages, and dependency management
  runtime/          Concurrency, error handling, and condition dispatch
  showcase/         Guided language tour that touches the whole surface area
  tooling/          VM, recursion, and extension-method experiments
  types-patterns/   Structs, enums, interfaces, contracts, and pattern matching
```

Run any script with `selene run path/to/example.selene` or stress-test them all via `selene test --mode all --verbose`.

## Documentation constellations

- The `docs/` folder powers a GitHub Pages site and mirrors the new structure: `guides/`, `integration/`, `reference/`, and `showcase/`.
- Start with [Guides → Getting started](docs/guides/getting-started.md) or [Guides → Language tour](docs/guides/language-tour.md).
- Dive into the [Reference](docs/reference/) for a formal catalog of syntax and semantics.
- Launch the [Example showcase](docs/showcase/) to view animated callouts for each script category.

## Embedding rocket fuel

Execute Selene code from your own Go programs by wiring together the lexer, parser, and runtime:

```go
package main

import (
    "fmt"
    "os"

    "github.com/cybellereaper/selenelang/internal/lexer"
    "github.com/cybellereaper/selenelang/internal/parser"
    "github.com/cybellereaper/selenelang/internal/runtime"
)

func main() {
    script, err := os.ReadFile("examples/fundamentals/hello.selene")
    if err != nil {
        panic(err)
    }

    l := lexer.New(string(script))
    p := parser.New(l)
    program := p.ParseProgram()
    if errs := p.Errors(); len(errs) > 0 {
        panic(errs)
    }

    rt := runtime.New()
    if _, err := rt.Run(program); err != nil {
        panic(err)
    }

    fmt.Println("script executed successfully")
}
```

From here you can extend the runtime with new built-ins, feed compiled bytecode into the VM, or export diagnostics into your own editor integrations.

## Development

Selene targets the Go 1.25.1 toolchain and includes a convenience `Makefile` to keep routine workflows fast:

```bash
# see the available tasks
make help

# format, vet, and test the codebase (tests run with the race detector and shuffling enabled)
make fmt
make vet
make test

# keep module metadata tidy and produce coverage artifacts
make tidy
make coverage

# check for known vulnerabilities (requires govulncheck)
make vulncheck
```

These commands wrap the standard Go tooling so contributors get consistent formatting, vetting, and testing locally and in CI.

## Repository map

```
assets/             Project artwork and supporting images
cmd/selene/         Command-line entry point and subcommand wiring
docs/
  guides/           Tutorials and onboarding material
  integration/      Embedding walkthroughs for Go applications
  reference/        Language reference for syntax and semantics
  showcase/         Animated example gallery
examples/
  fundamentals/     Core language building blocks
  modularity/       Module/package patterns
  runtime/          Advanced flow control & concurrency
  showcase/         The all-in-one language tour
  tooling/          VM + developer tooling playgrounds
  types-patterns/   Type system + pattern matching demonstrations
internal/           Lexer, parser, runtime, JIT, VM, and supporting packages
selene.toml         Project manifest describing modules, docs, and dependencies
selene.lock         Locked dependency checksums for reproducible vendors
vendor/             Vendored Selene packages with verified hashes
vscode-extension/   VS Code extension wrapping the language server and syntax assets
```

## Contributing meteors

Selene already implements a broad runtime surface area, but there is room for continued iteration: richer standard libraries, deeper tooling integrations, additional bytecode targets, and community-driven language features. Contributions, issues, and experiments are welcome—clone the repository, open pull requests, and share what you build with Selene.

> 💡 Pro tip: run `selene fmt -w .` and `selene test --mode all --verbose` before submitting changes to keep the galaxy shining.
