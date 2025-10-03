---
layout: default
title: Getting Started
---

# Getting started

This guide walks through installing the Selene toolchain, running scripts, and navigating the repository.

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

Use Go to compile the `selene` executable:

```bash
go build ./...
```

Install it into your `$GOBIN` so it is accessible anywhere:

```bash
go install ./cmd/selene
```

Verify the installation by checking the version banner:

```bash
selene --help
```

You should see usage information describing the `run`, `tokens`, `init`, and `deps` subcommands.

## Run your first script

Execute the bundled greeting example:

```bash
selene run examples/hello.selene
```

You should see output similar to:

```
Hello, Selene
```

You can also view the raw token stream without executing the script:

```bash
selene tokens examples/hello.selene
```

### Scaffold a new project

Generate a manifest, documentation folder, and starter source file in the current directory:

```bash
selene init github.com/you/stellar-selene --name "stellar-selene"
```

The command writes `selene.toml`, prepares a `docs/` directory, and places a `src/main.selene` entry point that prints a greeting.

### Vendor dependencies

To depend on another Selene package, vendor it into your project and lock its checksum:

```bash
selene deps add --path ../richmath --source https://github.com/selene-lang/richmath github.com/selene-lang/richmath v1.0.0
selene deps list
selene deps verify
```

Vendored code is copied into `vendor/` while `selene.lock` records the SHA-256 digest for reproducibility.

## Project layout

```
cmd/selene/         CLI entry point
internal/token/     Token definitions and keyword lookup
internal/lexer/     Scanner producing tokens with positions
internal/parser/    Pratt parser producing AST nodes
internal/ast/       AST node structures used by the parser
internal/runtime/   Tree-walking interpreter and runtime values
examples/           Sample Selene sources for experimentation
docs/               Documentation site published via GitHub Pages
vendor/             Vendored Selene packages managed by selene deps
selene.toml        Project manifest and dependency metadata
selene.lock        Locked dependency checksums
```

## Next steps

- Browse the [language tour](language-tour.md) to see the supported syntax and runtime behavior.
- Consult the [embedding guide](embedding.md) to run Selene inside your own Go projects.
- Experiment with the scripts in [examples](examples.md) and try extending them with your own functions.
