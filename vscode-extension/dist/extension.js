"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.activate = activate;
exports.deactivate = deactivate;
const vscode_1 = require("vscode");
const clientManager_1 = require("./clientManager");
const languageClientFactory_1 = require("./languageClientFactory");
let manager;
async function activate(context) {
    const output = vscode_1.window.createOutputChannel('Selene Language Server');
    const status = vscode_1.window.createStatusBarItem('seleneLanguageServer', vscode_1.StatusBarAlignment.Left, 1);
    status.text = 'Selene LSP: idle';
    status.tooltip = 'Selene language server is not running';
    status.command = 'selene.restartLanguageServer';
    status.show();
    manager = new clientManager_1.SeleneClientManager({ workspace: vscode_1.workspace, window: vscode_1.window }, output, status, languageClientFactory_1.createLanguageClient);
    await manager.activate();
    context.subscriptions.push(vscode_1.commands.registerCommand('selene.restartLanguageServer', async () => {
        if (!manager) {
            return;
        }
        await manager.restart({ announce: true });
    }));
    context.subscriptions.push({
        dispose: () => {
            void deactivate();
        },
    });
}
async function deactivate() {
    if (!manager) {
        return;
    }
    const current = manager;
    manager = undefined;
    await current.dispose();
}
//# sourceMappingURL=extension.js.map