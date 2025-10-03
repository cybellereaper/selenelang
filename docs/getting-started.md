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

You should see usage information along with command-line options such as `--tokens`.

## Run your first script

Execute the bundled greeting example:

```bash
selene examples/hello.selene
```

You should see output similar to:

```
Hello, Selene
```

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
```

## Next steps

- Browse the [language tour](language-tour.md) to see the supported syntax and runtime behavior.
- Consult the [embedding guide](embedding.md) to run Selene inside your own Go projects.
- Experiment with the scripts in [examples](examples.md) and try extending them with your own functions.
