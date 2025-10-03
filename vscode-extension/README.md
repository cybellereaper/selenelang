---

⚠️ **UNSTABLE FEATURE – WIP**  
This section is under active development. APIs and behavior may change without notice.  
Do not rely on this in production.

---

# Selene Language Support for VS Code

This Visual Studio Code extension provides an integrated experience for the Selene programming language. It bundles syntax highlighting, editor configuration, and a bridge to the `selene lsp` language server so scripts enjoy rich diagnostics and completions.

## Features

- Syntax highlighting for `.selene` sources powered by a TextMate grammar tuned to Selene keywords, string forms, and operators.
- Automatic bracket/quote pairing and standard comment toggles.
- Language Server Protocol (LSP) integration that launches the official Selene CLI to deliver diagnostics, completions, document formatting, semantic tokens, and symbol indexing.
- A persistent status bar item that surfaces the language server lifecycle (starting, ready, failed) along with a **Selene: Restart Language Server** command to quickly reload the backend after changing toolchain binaries or configuration.

## Requirements

- Install the Selene CLI (`go install ./cmd/selene`) so the extension can spawn `selene lsp`.
- Ensure the CLI is discoverable on your system `PATH`, or update the `Selene › Language Server Path` setting to reference an absolute path.

## Extension Settings

The extension exposes a few settings under **Selene**:

- `selene.languageServerPath`: Command used to start the language server (`selene` by default).
- `selene.languageServerArgs`: Arguments passed to the command (defaults to `["lsp"]`).
- `selene.languageServerEnv`: Additional environment variables merged into the server process.

## Packaging

Run the packaging scripts to build a distributable `.vsix` archive:

```bash
npm install
npm run package
```

The command produces `dist/selene-lang-support.vsix`, which you can install locally with `code --install-extension dist/selene-lang-support.vsix` or attach to a GitHub Release.

### Automated releases

The repository ships with a GitHub Actions workflow (`.github/workflows/extension-release.yml`) that runs the same packaging command on demand or whenever you push an `extension-v*` tag. It uploads the `.vsix` as both a workflow artifact and a release asset so the extension is always available from the **Releases** page.

## Development

1. Run `npm install` inside the `vscode-extension/` directory to fetch dependencies.
2. Open the folder in VS Code and run the **Launch Extension** configuration to debug.
3. The extension source lives under `src/` and is authored in TypeScript. The compiled output in `dist/` is the entrypoint that VS Code executes.
4. Run `npm run build` to emit the compiled extension, `npm run test` to execute the unit suite, and `npm run lint` to enforce code quality before committing changes.

The development workflow intentionally mirrors production: every command runs through the same TypeScript compiler and Mocha test harness used in CI, so extension behavior stays consistent whether you are iterating locally or packaging a release.
