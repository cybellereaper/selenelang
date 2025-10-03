# Selene Language Toolkit

Selene is an experimental programming language frontend implemented in Go. The toolchain now includes a lexer, Pratt-style parser, rich abstract syntax tree (AST) types, and a lightweight tree-walking interpreter. Use it to experiment with language ideas, embed Selene as a scripting language, or extend the runtime with new features.

## Features

- **Tokenization** – The lexer understands Selene keywords, operators, and literals while skipping both line (`// ...`) and block (`/* ... */`) comments with precise source positions.
- **AST definitions** – The `internal/ast` package models programs, declarations, statements, expressions, patterns, and contracts with positional metadata for diagnostics and tooling.
- **Pratt parser** – The parser supports modules, variable and function declarations, class/struct/enum definitions, contracts, match statements with pattern destructuring, and rich expression precedence (assignment, Elvis, logical, comparison, arithmetic, calls, indexing, and safe navigation).
- **Runtime interpreter** – The `internal/runtime` package evaluates Selene programs with lexical scoping, modules and imports, immutable/mutable bindings, first-class functions, user-defined structs/classes/enums, arrays, objects, rich control flow (`if`/`for`/`while`, `return`, `break`, `continue`), Elvis and optional operators, pattern matching (including struct and enum cases), and a small standard library (`print`). Function contracts are enforced at call time.
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

## Example gallery

Explore the scripts in the `examples/` directory to get a feel for different Selene features:

- `examples/hello.selene` – introductory greetings, expression-bodied functions, and simple printing.
- `examples/math.selene` – arithmetic helpers, composing functions, and string interpolation via `+` concatenation.
- `examples/collections.selene` – arrays, objects, optional property access, and indexing into both arrays and strings.
- `examples/options.selene` – optional chaining, Elvis defaults, and non-null assertions in practice.
- `examples/patterns.selene` – destructuring objects with `match` statements and pattern bindings.
- `examples/recursion.selene` – recursive functions that use `match` clauses for control flow.
- `examples/control.selene` – `if`/`else` branches, `for`/`while` loops, and the `return`/`break`/`continue` statements.
- `examples/types.selene` – struct methods, simple classes, enums, and pattern matching on enum cases.
- `examples/modules.selene` – modules, imports, and reusing exported helpers.
- `examples/contracts.selene` – standalone contract declarations and inline postconditions that validate function results.

Run any script with `selene path/to/script.selene` to see the output directly in your terminal.

## Documentation site

The repository ships with a `docs/` folder that is ready to publish with [GitHub Pages](https://docs.github.com/en/pages). To
serve the reference documentation:

1. Open the repository settings on GitHub.
2. Under **Pages**, choose the `main` branch and `/docs` folder as the publishing source.
3. Save the settings—within a minute GitHub will build and host the static site.

The generated site contains:

- A project overview and quick-start guide.
- Detailed walkthroughs of Selene syntax and runtime capabilities.
- A comprehensive [language reference](reference.md) cataloging every construct supported by the lexer and parser.
- Embedding examples showing how to execute Selene from your own Go programs.

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

- A richer standard library, asynchronous execution primitives, or packaging support for larger projects.
- A bytecode compiler and virtual machine.
- Source-to-source transpilers or code generators.
- Language services like formatting, symbol indexing, or IDE support.

Contributions and experiments are welcome—have fun exploring new language ideas with Selene!
