# Selene Language Toolkit

Selene is an experimental programming language frontend implemented in Go. The current toolchain provides a lexer, Pratt-style parser, rich abstract syntax tree (AST) types, and a simple command-line interface (CLI) for working with Selene source files. Use it to experiment with language ideas, build static tooling, or embed Selene as a hostable scripting language.

## Features

- **Tokenization** – The lexer understands Selene keywords, operators, literals, and skips both line (`// ...`) and block (`/* ... */`) comments while tracking precise source positions.
- **AST definitions** – The `internal/ast` package models programs, declarations, statements, expressions, patterns, and contract syntax, all annotated with start and end positions for tooling use.
- **Pratt parser** – The parser supports modules, variable and function declarations, class/struct/enum definitions, contracts, match expressions, and rich expression precedence (assignment, Elvis, logical, comparison, arithmetic, calls, indexing, and safe navigation).
- **CLI utility** – The `cmd/selene` binary can parse Selene files and optionally dump the raw token stream to help debug sources or tooling integrations.

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
module hello {
    import std.io.console as console;

    let greeting: String = "Hello";

    fn greet(name: String): String => greeting + ", " + name;

    fn main() {
        console.print(greet("Selene"));
    }
}
```

The parser recognizes module declarations, imports with aliases, immutable `let` bindings (and mutable `var`), functions with optional type annotations, expression-bodied functions (`=>`), and string concatenation.

## Using the CLI

Parse a script to ensure it is syntactically valid:

```bash
selene examples/hello.selene
```

Expected output:

```
Parsed examples/hello.selene successfully (1 top-level items).
```

Inspect the token stream to debug lexer behavior:

```bash
selene --tokens examples/hello.selene
```

You will see one token per line, including keywords, identifiers, punctuation, and literals.

## Embedding Selene as a scripting language

The toolkit is intended for host applications that need an embeddable language. Use the lexer and parser packages to turn source into an AST, then evaluate or transform the tree however you like. The snippet below shows how to parse a script and iterate over the top-level declarations:

```go
package main

import (
    "fmt"
    "os"

    "selenelang/internal/lexer"
    "selenelang/internal/parser"
    "selenelang/internal/ast"
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

    for _, item := range program.Items {
        switch node := item.(type) {
        case *ast.ModuleDeclaration:
            fmt.Printf("module %s with %d statements\n", node.Name.Name, len(node.Body.Statements))
        case *ast.FunctionDeclaration:
            fmt.Printf("fn %s with %d params\n", node.Name.Name, len(node.Params))
        }
    }
}
```

From here you can:

- Build an interpreter that walks the AST and executes statements in your host environment.
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
examples/           Sample Selene sources for experimentation
```

## Next steps

Selene currently focuses on front-end tooling. To make it a full scripting language you might implement:

- A runtime environment with built-in functions and data types.
- A bytecode compiler and virtual machine.
- Source-to-source transpilers or code generators.
- Language services like formatting, symbol indexing, or IDE support.

Contributions and experiments are welcome—have fun exploring new language ideas with Selene!
