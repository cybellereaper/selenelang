import {
  commands,
  workspace,
  window,
  StatusBarAlignment,
  type ExtensionContext,
} from 'vscode';

import { SeleneClientManager } from './clientManager';
import { createLanguageClient } from './languageClientFactory';

let manager: SeleneClientManager | undefined;

export async function activate(context: ExtensionContext): Promise<void> {
  const output = window.createOutputChannel('Selene Language Server');
  const status = window.createStatusBarItem('seleneLanguageServer', StatusBarAlignment.Left, 1);
  status.text = 'Selene LSP: idle';
  status.tooltip = 'Selene language server is not running';
  status.command = 'selene.restartLanguageServer';
  status.show();

  manager = new SeleneClientManager(
    { workspace, window },
    output,
    status,
    createLanguageClient,
  );

  await manager.activate();

  context.subscriptions.push(
    commands.registerCommand('selene.restartLanguageServer', async () => {
      if (!manager) {
        return;
      }
      await manager.restart({ announce: true });
    }),
  );

  context.subscriptions.push({
    dispose: () => {
      void deactivate();
    },
  });
}

export async function deactivate(): Promise<void> {
  if (!manager) {
    return;
  }
  const current = manager;
  manager = undefined;
  await current.dispose();
}
