<p align="center">
  <img src="https://media.giphy.com/media/v1.Y2lkPTc5MGI3NjExbXA5cXA2ZXB3NmR6NjY0cHh6NnFrYWFoMnRxM2VpdTNmMGk2ZGszbiZlcD12MV9naWZzX3NlYXJjaCZjdD1n/1BdIP9eT0C8kc/giphy.gif" width="200" alt="VS Code neon animation" />
</p>

# Selene Language Support for VS Code

> ⚠️ **UNSTABLE FEATURE – WIP** – APIs and behavior may change without notice. Do not rely on this in production.

The Selene VS Code extension ships alongside the reorganized toolkit so you can code against the moonlit runtime with style. It bundles syntax highlighting, language server wiring, and packaging helpers to make editing `.selene` files feel otherworldly.

## Feature stardust

- **🎨 Syntax highlighting** powered by a TextMate grammar tuned to Selene keywords, string forms, and operators.
- **🧠 Smart language server** integration that launches `selene lsp` for diagnostics, completions, formatting, semantic tokens, and symbol indexing.
- **🔁 One-click restarts** via a persistent status bar item and the **Selene: Restart Language Server** command.
- **🌌 Cozy defaults** for bracket/quote pairing, comment toggles, and formatting so your editing orbit stays smooth.

## Prerequisites

- Install the Selene CLI (`go install ./cmd/selene`) so the extension can spawn `selene lsp`.
- Ensure the CLI lives on your system `PATH`, or update the **Selene › Language Server Path** setting to reference an absolute location.

## Extension settings

All settings live under the **Selene** namespace:

| Setting | Purpose |
| --- | --- |
| `selene.languageServerPath` | Command used to start the language server (`selene` by default). |
| `selene.languageServerArgs` | Arguments passed to the command (defaults to `["lsp"]`). |
| `selene.languageServerEnv` | Additional environment variables merged into the server process. |

## Packaging & release flow

```bash
npm install
npm run package
```

The command produces `dist/selene-lang-support.vsix`. Install it locally with:

```bash
code --install-extension dist/selene-lang-support.vsix
```

A GitHub Actions workflow (`.github/workflows/extension-release.yml`) mirrors the same packaging command whenever you push an `extension-v*` tag. It uploads the `.vsix` as both a workflow artifact and a release asset so the extension is always available from the **Releases** page.

## Development loop

1. Run `npm install` inside `vscode-extension/` to fetch dependencies.
2. Open the folder in VS Code and start the **Launch Extension** configuration to debug in a dedicated Extension Development Host.
3. Source lives under `src/` in TypeScript. `npm run build` emits the compiled output into `dist/`.
4. Run `npm run lint` and `npm run test` before committing to keep the neon glow consistent.

> 🛰️ Bonus: the main repository now exposes the curated examples under `examples/` → `fundamentals`, `runtime`, `types-patterns`, and more. Point the extension at those folders to explore real code while testing new features.
