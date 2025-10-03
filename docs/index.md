---
layout: default
title: Selene Overview
---

# Selene Overview

Selene is an experimental programming language toolkit implemented in Go. The lexer, parser, and runtime work together to let you
prototype new language ideas, embed Selene as a scripting language, or analyze Selene source code programmatically.

## Quick links

- [Getting started](getting-started.md) – install the CLI, run scripts, and understand the project layout.
- [Language tour](language-tour.md) – explore Selene syntax, expression forms, and runtime semantics.
- [Embedding guide](embedding.md) – execute Selene code from your own Go applications.
- [Example scripts](examples.md) – curated walkthroughs of the sample programs included in the repository.

## What you can build today

The current runtime focuses on expression evaluation, first-class functions, arrays, and objects. It does not yet implement control
flow statements such as loops or `if`/`match`, making it best suited for DSL experiments, configuration, and data transformation
tasks driven by function calls and expression chaining.

## Publishing the docs

To publish this site with GitHub Pages:

1. Push the repository to GitHub.
2. Open **Settings → Pages**.
3. Choose the `main` branch and the `/docs` folder.
4. Save your changes and wait for GitHub to build the static site.

Once deployed you will have a hosted reference for the language, examples, and integration guidance.
