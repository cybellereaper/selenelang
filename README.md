<table><img src="https://github.com/cybellereaper/selenelang/blob/main/assets/image.png" width="98" align="left"></table>

# Selene Language Toolkit

[![Go Tests](https://github.com/cybellereaper/selenelang/actions/workflows/go-tests.yml/badge.svg)](https://github.com/cybellereaper/selenelang/actions/workflows/go-tests.yml)

Selene is an experimental programming language frontend implemented in Go. The toolchain now includes a lexer, Pratt-style parser, rich abstract syntax tree (AST) types, and a lightweight tree-walking interpreter. Use it to experiment with language ideas, embed Selene as a scripting language, or extend the runtime with new features.

## Features

- **Tokenization** – The lexer understands Selene keywords, operators, and literals while skipping both line (`// ...`) and block (`/* ... */`) comments with precise source positions.
- **AST definitions** – The `internal/ast` package models programs, declarations, statements, expressions, patterns, and contracts with positional metadata for diagnostics and tooling.
- **Pratt parser** – The parser supports modules, variable and function declarations, class/struct/enum definitions, contracts, match statements with pattern destructuring, and rich expression precedence (assignment, Elvis, logical, comparison, arithmetic, calls, indexing, and safe navigation).
- **Runtime interpreter** – The `internal/runtime` package evaluates Selene programs with lexical scoping, package declarations, modules and imports (including Go-style string imports and aliases), immutable/mutable bindings, first-class functions, user-defined structs/classes/enums, arrays, objects, and rich control flow (`if`/`for`/`while`, `return`, `break`, `continue`). The runtime now supports pointer semantics (`&` and `*`), extension functions declared with `ext fn`, structural interfaces with `is`/`!is`, expressive string literals (interpolation with `${}`, format specifiers via `f"..."`, raw/triple-quoted strings), resource-safe `using` blocks, structured `try`/`catch`/`finally` plus `throw`, lightweight concurrency primitives (`spawn`, `channel`, `await`), rule-based `condition { when ... }` dispatch, and a growing standard library (`print`, `format`, `spawn`, `channel`). Function contracts are enforced at call time.
- **Bytecode compiler + VM** – Parse trees can be compiled into compact bytecode chunks via `runtime.Compile`, inspected with `selene build`, and executed through a lightweight virtual machine (`selene run --vm`) that reuses the runtime environment. This makes it easy to reason about evaluation order or embed Selene in ahead-of-time workflows.
- **JIT runner and Windows bundler** – `selene run --jit` executes programs through a cached JIT pipeline that avoids repeated interpreter dispatch, and `selene build --windows-exe` emits Windows executables that embed Selene sources alongside the new JIT runtime for easy distribution.
- **Formatter and transpiler** – `selene fmt` rewrites sources into a canonical layout (also exposed through the language server), and `selene transpile --lang go` emits Go scaffolding that mirrors Selene packages. These helpers simplify editor integration, CI formatting checks, and experimentation with cross-language code generation.
- **Project metadata** – The repository includes a `selene.toml` manifest and companion `selene.lock` file that describe the project module path, entry point, documentation roots, example directories, and vendored dependencies with reproducible SHA-256 checksums.
- **CLI utility** – The `cmd/selene` binary can execute Selene scripts, dump the raw token stream, scaffold new workspaces, manage vendored dependencies with secure checksum verification, compile/disassemble bytecode, run the VM, format sources, transpile to Go, and expose an editor-ready language server via `selene lsp`.
- **VS Code extension** – A ready-to-publish extension (`vscode-extension/`) packages syntax highlighting, editor configuration, and a client that proxies VS Code to the Selene language server for diagnostics, completions, formatting, and symbol indexing.

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
```

```

This script demonstrates immutable `let` bindings, optional type annotations, expression-bodied functions (`=>`), string concatenation, first-class functions, and calling the built-in `print` helper. Selene automatically invokes `main` for you, mirroring the Go toolchain's entry-point semantics.

## Using the CLI

Run a script to execute it:

```bash
selene run examples/hello.selene
```

Expected output:

```
Hello, Selene
```

Inspect the token stream instead of running the program:

```bash
selene tokens examples/hello.selene
```

You will see one token per line, including keywords, identifiers, punctuation, and literals.

Execute the same script through the bytecode virtual machine to validate the VM pipeline:

```bash
selene run --vm examples/hello.selene
```

Produce and inspect a bytecode listing without executing the program:

```bash
selene build --out hello.bc examples/hello.selene
cat hello.bc
```

Reformat Selene sources in place (or print the formatted output to STDOUT):

```bash
selene fmt -w examples/*.selene
```

Run every curated example through the interpreter, VM, or JIT pipeline:

```bash
selene test --mode all --verbose
```

Transpile Selene to Go scaffolding for experimentation or embedding:

```bash
selene transpile --lang go --out hello.go examples/hello.selene
```

Initialize a new project with a manifest, documentation, and starter source file:

```bash
selene init github.com/you/awesome-selene --name "awesome-selene"
```

Manage dependencies by vendoring local or remote sources with cryptographic checksums:

```bash
# Vendor a dependency from a local checkout (or a freshly cloned repository)
selene deps add --path ../richmath --source https://github.com/selene-lang/richmath github.com/selene-lang/richmath v1.0.0

# Inspect the recorded requirements
selene deps list

# Recalculate hashes to ensure the vendor tree matches selene.lock
selene deps verify
```

The dependency commands update both `selene.toml` and `selene.lock`, ensuring repeatable builds and tamper-proof vendored content.

## Editor integration

Launch the built-in language server to power diagnostics, completions, and formatting-friendly metadata inside your favourite editor:

```bash
selene lsp
```

The server speaks the [Language Server Protocol](https://microsoft.github.io/language-server-protocol/specification) over STDIN/STDOUT. Configure your editor to spawn `selene lsp` in your project root and Selene will stream syntax errors, lexer issues, and keyword/builtin completions as you type. Because dependency resolution happens before execution, diagnostics also reflect imported modules recorded in `selene.toml`/`selene.lock`.

Many editors support custom language servers out of the box:

- **VS Code** – Open the `vscode-extension/` folder in VS Code, run `npm install`, and launch the provided extension (`F5` → **Launch Extension**) to load Selene syntax highlighting and automatically start `selene lsp`. Ship ready-to-install builds with `npm run package`, which produces `dist/selene-lang-support.vsix` that you can sideload or upload to a GitHub Release.
- **Neovim** – Use `:LspStart` with a custom client definition via `lspconfig`, e.g. `cmd = { "selene", "lsp" }`.
- **Helix / Zed / Sublime** – Add a language server command referencing `selene lsp` and map it to `.selene` sources.

The server currently publishes:

- Syntax diagnostics sourced from both the lexer (illegal tokens) and parser (structural errors with precise locations).
- Keyword and builtin completions to accelerate script authoring.
- Document formatting edits powered by the same formatter behind `selene fmt`.
- Document/workspace symbol indexes so fuzzy symbol search in editors stays in sync with Selene sources.
- Document lifecycle awareness (`didOpen`, `didChange`, `didSave`, `didClose`) so diagnostics clear as files are edited.

Hover information, go-to-definition, and other advanced tooling can build on the same server foundation.

## VS Code extension releases

Automated packaging keeps the extension easy to distribute. The repository includes a GitHub Actions workflow that

- installs the extension dependencies,
- runs `npm run package` to emit `dist/selene-lang-support.vsix`, and
- uploads the result both as a build artifact and as a GitHub Release asset when you push an `extension-v*` tag.

Trigger the workflow manually from the **Actions** tab or by tagging a release:

```bash
git tag extension-v0.1.0
git push origin extension-v0.1.0
```

Once the workflow completes, download the generated `.vsix` from the release page or install it locally with `code --install-extension dist/selene-lang-support.vsix`.

## Example gallery

Explore the scripts in the `examples/` directory to get a feel for different Selene features:

- `examples/language_tour.selene` – a single program that threads together modules, structs, enums, condition blocks, and loops.
- `examples/hello.selene` – immutable bindings, expression-bodied functions, and string formatting helpers.
- `examples/math.selene` – arithmetic helpers exported from a module and consumed via imports.
- `examples/collections.selene` – arrays, objects, optional property access, and Elvis defaults.
- `examples/options.selene` – navigating nested optionals, non-null assertions, and graceful fallbacks.
- `examples/patterns.selene` – destructuring objects with `match` statements and pattern bindings.
- `examples/recursion.selene` – recursive factorial and Fibonacci implementations.
- `examples/control.selene` – loops, conditionals, and early exits (`return`/`break`/`continue`).
- `examples/types.selene` – struct methods, classes, enums, and matching on enum cases.
- `examples/modules.selene` – declaring modules and importing their exports by name.
- `examples/packages.selene` – package headers and augmented assignment operators.
- `examples/contracts.selene` – reusable contract declarations and inline postconditions.
- `examples/strings.selene` – interpolation, raw strings, format specifiers, and extensions.
- `examples/extensions.selene` – extension methods on built-in types such as `String` and `Int`.
- `examples/pointers.selene` – address-of/dereference operators and swapping by reference.
- `examples/interfaces.selene` – structural interface conformance with structs and extension methods.
- `examples/concurrency.selene` – spawning tasks, communicating over channels, and awaiting asynchronous work.
- `examples/vm.selene` – bytecode-friendly loops and accumulation helpers for VM experimentation.
- `examples/errors.selene` – `using` blocks, `try`/`catch`/`finally`, and explicit `throw` expressions.
- `examples/conditions.selene` – rule-driven `condition { when ... }` dispatch.
- `examples/dependency.selene` – importing a vendored module from `github.com/selene-lang/richmath`.

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
vendor/             Vendored Selene packages with checksums
selene.toml        Project manifest and dependency metadata
selene.lock        Locked dependency checksums for verification
```

## Next steps

Selene currently ships with a minimal interpreter focused on core language features. Future work could explore:

- A richer standard library, asynchronous execution primitives, or packaging support for larger projects.
- A bytecode compiler and virtual machine.
- Source-to-source transpilers or code generators.
- Language services like formatting, symbol indexing, or IDE support.

Contributions and experiments are welcome—have fun exploring new language ideas with Selene!
