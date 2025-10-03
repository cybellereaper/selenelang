<p align="left">
  <img src="https://github.com/cybellereaper/selenelang/blob/main/assets/image.png" width="98" alt="Selene logo" />
</p>

# Selene Language Toolkit

[![Go Tests](https://github.com/cybellereaper/selenelang/actions/workflows/go-tests.yml/badge.svg)](https://github.com/cybellereaper/selenelang/actions/workflows/go-tests.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/cybellereaper/selenelang.svg)](https://pkg.go.dev/github.com/cybellereaper/selenelang)

Selene is an experimental programming-language frontend written in Go. It packages the full toolchain needed to prototype new language ideas: a lexer, a Pratt parser that produces rich abstract syntax trees (ASTs), an interpreter-grade runtime, a bytecode compiler and virtual machine, an opinionated formatter, and an ergonomic CLI. Use Selene to iterate on language design, embed the runtime inside Go applications, or publish tooling such as language servers and transpilers.

## Overview of the toolchain

Selene is intentionally modular so that every layer can be reused on its own or composed into a full pipeline:

- **Tokens and lexer (`internal/token`, `internal/lexer`)** – Provides the keyword table, punctuators, literal types, and a lexer that emits tokens with precise `Pos`/`End` locations. Block (`/* ... */`) and line (`// ...`) comments are skipped while preserving offsets for diagnostics.
- **AST (`internal/ast`)** – Defines node structures for packages, declarations, statements, expressions, patterns, and contracts. Each node embeds positional metadata which powers error reporting, IDE integrations, and code generation.
- **Parser (`internal/parser`)** – Implements a Pratt parser with operator-precedence rules that understand Selene modules, imports, variable and function declarations, structs/classes/enums, pattern-matching, contracts, and expression-level features such as Elvis (`?:`), safe navigation (`?.`), and extension-method syntax.
- **Runtime interpreter (`internal/runtime`)** – Executes ASTs in a tree-walking interpreter with lexical scopes, immutable and mutable bindings, first-class functions, user-defined types, pointer semantics (`&`, `*`), structural interfaces (`is`/`!is`), resource-safe `using` blocks, structured error handling (`try`/`catch`/`finally`, `throw`), lightweight concurrency primitives (`spawn`, `channel`, `await`), and a standard library of built-ins (`print`, `format`, etc.). Function contracts are enforced at call sites.
- **Bytecode compiler and VM** – `runtime.Compile` turns parse trees into compact bytecode chunks that can be executed by the Selene virtual machine (`selene run --vm`). The VM reuses runtime values so you can experiment with evaluation order or ahead-of-time execution.
- **JIT runner and Windows bundler** – `selene run --jit` keeps cached bytecode for reduced dispatch overhead, and `selene build --windows-exe` packages programs together with the runtime to create standalone Windows executables.
- **Formatter and transpiler** – `selene fmt` produces canonical layouts for `.selene` sources (the same engine backs the language server), and `selene transpile --lang go` generates Go scaffolding that mirrors Selene packages for embedding or experimentation.
- **Language server and editor tooling** – `selene lsp` speaks the Language Server Protocol, enabling diagnostics, completion, formatting, and symbol indexing in editors. A ready-to-publish VS Code extension (`vscode-extension/`) wraps the server with syntax highlighting and packaging workflows.
- **Project metadata** – `selene.toml` and `selene.lock` record module metadata, documentation roots, example directories, and vendored dependencies with SHA-256 checksums so that builds remain reproducible.

## Getting started

### Prerequisites
- Go 1.21 or later (the module itself targets Go 1.24, but the CLI compiles with the current stable toolchain).
- A POSIX-compatible shell for running the examples.

### Install the module and CLI

```bash
go mod tidy
go build ./...
```

Install the `selene` binary into your `$GOBIN` to make the CLI globally available:

```bash
go install ./cmd/selene
```

### Run the first script

Create a simple program in `examples/hello.selene`:

```selene
// examples/hello.selene
let greeting: String = "Hello";

fn greet(name: String): String => greeting + ", " + name;

fn main() {
    print(greet("Selene"));
}
```

Execute it with the interpreter:

```bash
selene run examples/hello.selene
```

Output:

```
Hello, Selene
```

## CLI capabilities

The `cmd/selene` binary exposes the entire toolchain. Core subcommands include:

| Command | Purpose |
| --- | --- |
| `selene run <file>` | Interpret a script directly, or add `--vm` / `--jit` to execute via the bytecode VM or JIT runner. |
| `selene tokens <file>` | Print the token stream emitted by the lexer. |
| `selene fmt [-w] <pattern>` | Format Selene sources in-place or to STDOUT. |
| `selene build --out <file> <input>` | Compile a script to bytecode and write the chunk to disk. |
| `selene transpile --lang go --out <file> <input>` | Generate Go scaffolding for the given Selene module. |
| `selene test --mode all --verbose` | Execute curated examples through the interpreter, VM, or JIT pipelines. |
| `selene deps add/list/verify` | Manage vendored dependencies with cryptographic checksums recorded in `selene.toml` and `selene.lock`. |
| `selene init <module>` | Scaffold a new workspace with a manifest, documentation skeleton, and starter source file. |
| `selene lsp` | Launch the Language Server Protocol endpoint used by editors and the VS Code extension. |

Use `selene help <command>` for the full flag reference.

## Example gallery

The `examples/` folder is intentionally comprehensive so you can explore language semantics and runtime behavior. Highlights include:

- `language_tour.selene` – a single program that exercises modules, structs, enums, pattern matching, and advanced control flow.
- `collections.selene` – arrays, objects, optional property access, and Elvis defaults.
- `concurrency.selene` – task spawning, channel communication, and `await` coordination.
- `conditions.selene` – rule-based `condition { when ... }` dispatch.
- `errors.selene` – `using` blocks, structured error handling, and `throw` expressions.
- `interfaces.selene` – structural interface conformance with structs and extension methods.
- `pointers.selene` – address-of/dereference operators and reference semantics.

Run any example with `selene run path/to/example.selene` to see the runtime in action.

## Documentation

- The `docs/` directory is publishable via GitHub Pages. Configure the repository to serve `main` → `/docs` to host the language reference, quick-start guide, and embedding walkthroughs.
- The Go API reference for the module is available on [pkg.go.dev](https://pkg.go.dev/github.com/cybellereaper/selenelang) via the badge above.
- AST and runtime packages expose positional metadata (`Pos`, `End`) to power tooling such as linters, formatters, and IDE integrations.

## Embedding Selene inside Go

Execute Selene code from your own Go programs by wiring together the lexer, parser, and runtime:

```go
package main

import (
    "fmt"
    "os"

    "selenelang/internal/lexer"
    "selenelang/internal/parser"
    "selenelang/internal/runtime"
)

func main() {
    script, err := os.ReadFile("examples/hello.selene")
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

## Repository layout

```
cmd/selene/         Command-line entry point and subcommand wiring
internal/token/     Token definitions and keyword lookup tables
internal/lexer/     Scanner that produces tokens with precise positions
internal/parser/    Pratt parser that builds AST nodes from token streams
internal/ast/       AST structures with metadata for tooling
internal/runtime/   Interpreter, runtime values, VM compiler, and execution pipelines
examples/           Curated Selene programs showcasing language features
docs/               Publishable documentation set for GitHub Pages
vscode-extension/   VS Code extension that wraps the language server and syntax assets
assets/             Project artwork and supporting images
selene.toml         Project manifest describing modules, docs, and dependencies
selene.lock         Locked dependency checksums for reproducible vendors
vendor/             Vendored Selene packages with verified hashes
```

## Roadmap and contributions

Selene already implements a broad runtime surface area, but there is room for continued iteration: richer standard libraries, deeper tooling integrations, additional bytecode targets, and community-driven language features. Contributions, issues, and experiments are welcome—clone the repository, open pull requests, and share what you build with Selene.
