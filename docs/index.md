---
layout: default
title: Selene Overview
---

# Selene Overview

Selene is an experimental programming language toolkit implemented in Go. The lexer, parser, runtime, bytecode compiler, and CLI work together to help you prototype new language ideas, embed Selene as a scripting engine, or analyze Selene source code programmatically.

<div align="center">
  <img src="https://media.giphy.com/media/v1.Y2lkPTc5MGI3NjExYzQza2N1OGV0dXZqYnZwcGdiMWl6MXYwODh6aWliNnMwOWhwZ2V0dSZlcD12MV9naWZzX3NlYXJjaCZjdD1n/OBhJzs7PaY4HS/giphy.gif" alt="Looping moon animation" width="220" />
  <p><em>Selene's moon mascot watches over your language experiments.</em></p>
</div>

## Quick links

- [Guided onboarding](guides/getting-started.md) – install the CLI, run scripts, and understand the project layout.
- [Language tour](guides/language-tour.md) – explore Selene syntax, expression forms, and runtime semantics.
- [Language reference](reference/) – browse every construct recognized by the lexer, parser, and runtime.
- [Embedding guide](integration/embedding.md) – execute Selene code from your own Go applications.
- [Example showcase](showcase/) – curated walkthroughs of the sample programs included in the repository.

## What you can build today

The current runtime covers expression evaluation, pattern matching, first-class functions, package headers, modules with Go-style imports, user-defined types, and imperative control flow with augmented assignment helpers. Recent updates add pointer semantics, structural interfaces, extension methods, rich string interpolation/formatting (including raw and triple-quoted literals), resource-safe `using` statements, `try`/`catch`/`finally` with `throw`, lightweight concurrency (`spawn`, `channel`, `await`), rule-driven `condition { when ... }` dispatch, secure dependency vendoring backed by `selene.toml`/`selene.lock` manifests, and editor integration powered by the bundled `selene lsp` language server. A ready-to-run VS Code extension under `vscode-extension/` wires that server into a polished editing experience complete with syntax highlighting and restart commands. A bytecode compiler/VM and Go transpiler unlock alternate execution flows, while the built-in formatter keeps Selene sources tidy both on the command line and inside supporting editors thanks to document/workspace symbol indexing and `textDocument/formatting` support.

Combine these building blocks—`if`/`for`/`while`, structs/classes/enums, pattern-matching, interfaces, concurrency primitives, and vendored packages—to sketch real-world scripts. The standard library remains intentionally tiny so you can experiment with your own runtime helpers or pull in external Selene modules with deterministic checksums.

## Publishing the docs

To publish this site with GitHub Pages:

1. Push the repository to GitHub.
2. Open **Settings → Pages**.
3. Choose the `main` branch and the `/docs` folder.
4. Save your changes and wait for GitHub to build the static site.

Once deployed you will have a hosted reference for the language, examples, and integration guidance.
