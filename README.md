# Selene Language Toolkit

Selene is an experimental programming language frontend implemented in Go. The toolchain now includes a lexer, Pratt-style parser, rich abstract syntax tree (AST) types, and a lightweight tree-walking interpreter. Use it to experiment with language ideas, embed Selene as a scripting language, or extend the runtime with new features.

## Features

- **Tokenization** – The lexer understands Selene keywords, operators, and literals while skipping both line (`// ...`) and block (`/* ... */`) comments with precise source positions.
- **AST definitions** – The `internal/ast` package models programs, declarations, statements, expressions, patterns, and contracts with positional metadata for diagnostics and tooling.
- **Pratt parser** – The parser supports modules, variable and function declarations, class/struct/enum definitions, contracts, match expressions, and rich expression precedence (assignment, Elvis, logical, comparison, arithmetic, calls, indexing, and safe navigation).
- **Runtime interpreter** – The `internal/runtime` package evaluates Selene programs with lexical scoping, immutable/mutable bindings, first-class functions, arrays, objects, arithmetic, comparisons, and a small standard library (`print`).
- **CLI utility** – The `cmd/selene` binary can execute Selene scripts or dump the raw token stream to help debug sources and tooling integrations.

## Installation

Selene is distributed as a standard Go module. With Go 1.21 or newer installed, fetch the dependencies and build the CLI:

```bash
go mod tidy
go build ./...
```

To install the `selene` CLI in your `$GOBIN`, run:

```bash
go install ./cmd/selene
```

## Writing your first Selene script

Create `examples/hello.selene` with a few declarations and expressions:

```selene
// examples/hello.selene
let greeting: String = "Hello";

fn greet(name: String): String => greeting + ", " + name;

fn main() {
    print(greet("Selene"));
}

main();
```

This script demonstrates immutable `let` bindings, optional type annotations, expression-bodied functions (`=>`), string concatenation, first-class functions, and calling the built-in `print` helper. The final `main();` call drives execution just like a typical scripting entry point.

## Using the CLI

Run a script to execute it:

```bash
selene examples/hello.selene
```

Expected output:

```
Hello, Selene
```

Inspect the token stream instead of running the program:

```bash
selene --tokens examples/hello.selene
```

You will see one token per line, including keywords, identifiers, punctuation, and literals.

## Embedding Selene as a scripting language

Use the lexer, parser, and runtime packages to execute Selene inside your Go application:

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

From here you can:

- Extend the runtime with additional built-in functions or host integrations.
- Compile Selene into another language or bytecode target.
- Write linters, formatters, or static analyzers that inspect the rich node metadata.

Because every node exposes both `Pos()` and `End()` locations, it is straightforward to produce precise diagnostics or IDE features.

## Project layout

```
cmd/selene/         CLI entry point
internal/token/     Token definitions and keyword lookup
internal/lexer/     Scanner producing tokens with positions
internal/parser/    Pratt parser producing AST nodes
internal/ast/       AST node structures used by the parser
internal/runtime/   Tree-walking interpreter and runtime values
examples/           Sample Selene sources for experimentation
```

## Next steps

Selene currently ships with a minimal interpreter focused on core language features. Future work could explore:

- A richer standard library and module system for reusable code.
- A bytecode compiler and virtual machine.
- Source-to-source transpilers or code generators.
- Language services like formatting, symbol indexing, or IDE support.

Contributions and experiments are welcome—have fun exploring new language ideas with Selene!
