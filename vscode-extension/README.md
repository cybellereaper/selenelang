# Selene Language Support for VS Code

This Visual Studio Code extension provides an integrated experience for the Selene programming language. It bundles syntax highlighting, editor configuration, and a bridge to the `selene lsp` language server so scripts enjoy rich diagnostics and completions.

## Features

- Syntax highlighting for `.selene` sources powered by a TextMate grammar tuned to Selene keywords, string forms, and operators.
- Automatic bracket/quote pairing and standard comment toggles.
- Language Server Protocol (LSP) integration that launches the official Selene CLI to deliver diagnostics, completions, document formatting, and symbol indexing.
- A **Selene: Restart Language Server** command to quickly reload the backend after changing toolchain binaries or configuration.

## Requirements

- Install the Selene CLI (`go install ./cmd/selene`) so the extension can spawn `selene lsp`.
- Ensure the CLI is discoverable on your system `PATH`, or update the `Selene › Language Server Path` setting to reference an absolute path.

## Extension Settings

The extension exposes a few settings under **Selene**:

- `selene.languageServerPath`: Command used to start the language server (`selene` by default).
- `selene.languageServerArgs`: Arguments passed to the command (defaults to `["lsp"]`).
- `selene.languageServerEnv`: Additional environment variables merged into the server process.

## Development

1. Run `npm install` inside the `vscode-extension/` directory to fetch dependencies.
2. Open the folder in VS Code and run the **Launch Extension** configuration to debug.
3. The extension entrypoint is `extension.js`, which wires the VS Code client to the Selene language server.

The extension is authored in plain JavaScript to minimize build tooling—no bundlers or transpilers are required.
