---
layout: default
title: Getting Started
---

# Getting started

Welcome to the Selene launchpad! This guide walks through installing the toolchain, running your first script, and exploring the freshly organized repository layout.

## Install prerequisites

- [Go 1.21+](https://go.dev/dl/) for building the CLI and embedding Selene.
- A terminal with access to standard developer tools.

## Fetch the source

Clone the repository and download dependencies:

```bash
git clone https://github.com/your-user/selenelang.git
cd selenelang
go mod tidy
```

## Build and install the CLI

Compile the `selene` executable and drop it onto your `$PATH`:

```bash
go build ./...
go install ./cmd/selene
```

Verify the installation:

```bash
selene --help
```

You should see usage information describing the `run`, `test`, `tokens`, `fmt`, `build`, `transpile`, `init`, `deps`, and `lsp` subcommands.

## Run your first script

Execute the bundled greeting example from the reorganized **fundamentals** collection:

```bash
selene run examples/fundamentals/hello.selene
```

Expected output:

```
Hello, Selene
```

Peek at the raw token stream without executing the script:

```bash
selene tokens examples/fundamentals/hello.selene
```

Run the same program through each execution backend:

```bash
selene run --vm examples/fundamentals/hello.selene
selene run --jit examples/fundamentals/hello.selene
```

Emit bytecode or package the script into a Windows executable:

```bash
selene build --out hello.bc examples/fundamentals/hello.selene
selene build --windows-exe hello.exe examples/fundamentals/hello.selene
```

Format every script in the example library:

```bash
selene fmt -w examples
```

Stress-test the full gallery via the interpreter, VM, and JIT backends:

```bash
selene test --mode all
```

Generate Go scaffolding from Selene code:

```bash
selene transpile --lang go --out hello.go examples/fundamentals/hello.selene
```

### Scaffold a new project

Prepare a manifest, documentation skeleton, and starter source file in the current directory:

```bash
selene init github.com/you/stellar-selene --name "stellar-selene"
```

The command writes `selene.toml`, prepares a `docs/` directory, and places a `src/main.selene` entry point that prints a greeting.

### Vendor dependencies

Vendor another Selene package and lock its checksum:

```bash
selene deps add --path ../richmath --source https://github.com/selene-lang/richmath github.com/selene-lang/richmath v1.0.0
selene deps list
selene deps verify
```

Vendored code is copied into `vendor/` while `selene.lock` records the SHA-256 digest for reproducibility.

## Enable editor support

The Selene CLI embeds a Language Server Protocol (LSP) implementation so editors can surface diagnostics and completions as you type. Launch it from your project root:

```bash
selene lsp
```

Point your editor's LSP client at the command above (for example, `cmd = { "selene", "lsp" }` in Neovim `lspconfig`). The server reports lexer/parser errors, clears diagnostics on save, formats documents, indexes document/workspace symbols, and offers keyword/builtin completions out of the box.

For Visual Studio Code users, the repository ships with a dedicated extension under `vscode-extension/`. Open that folder in VS Code, run `npm install`, and start the **Launch Extension** debug configuration to load Selene syntax highlighting and a preconfigured language server client. When you're ready to share it, generate a distributable archive with `npm run package`—the script produces `dist/selene-lang-support.vsix`, which you can install locally with `code --install-extension` or upload as a GitHub Release asset.

## Project layout

```
cmd/selene/         CLI entry point
internal/token/     Token definitions and keyword lookup
internal/lexer/     Scanner producing tokens with positions
internal/parser/    Pratt parser producing AST nodes
internal/ast/       AST node structures used by the parser
internal/runtime/   Tree-walking interpreter, VM, and JIT helpers
examples/
  fundamentals/     Language basics (hello world, math, strings, flow control)
  modularity/       Modules, packages, and dependency management
  runtime/          Concurrency, error handling, and condition dispatch
  showcase/         Guided language tour that touches the whole surface area
  tooling/          VM, recursion, and extension-method experiments
  types-patterns/   Structs, enums, interfaces, contracts, and pattern matching
docs/
  guides/           Tutorials and onboarding material
  integration/      Embedding walkthroughs for Go applications
  reference/        Formal language reference for syntax and semantics
  showcase/         Example gallery with runnable scripts
vendor/             Vendored Selene packages managed by `selene deps`
selene.toml        Project manifest and dependency metadata
selene.lock        Locked dependency checksums
```

## Next steps

- Browse the [language tour](language-tour.md) to see the supported syntax and runtime behavior.
- Consult the [embedding guide](../integration/embedding.md) to run Selene inside your own Go projects.
- Experiment with the scripts in the [example showcase](../showcase/) and extend them with your own functions.
