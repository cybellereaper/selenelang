const { workspace, window, commands } = require('vscode');
const { LanguageClient, TransportKind } = require('vscode-languageclient/node');

let client;

function resolveServerCommand() {
    const config = workspace.getConfiguration('selene');
    const command = config.get('languageServerPath', 'selene');
    const args = config.get('languageServerArgs', ['lsp']);
    const env = config.get('languageServerEnv', {});
    return { command, args, env };
}

async function startClient(context) {
    if (client) {
        await client.stop();
        client = undefined;
    }

    const { command, args, env } = resolveServerCommand();

    const serverOptions = {
        command,
        args,
        options: {
            env: { ...process.env, ...env },
            cwd: workspace.workspaceFolders && workspace.workspaceFolders.length > 0
                ? workspace.workspaceFolders[0].uri.fsPath
                : undefined,
        },
        transport: TransportKind.stdio,
    };

    const clientOptions = {
        documentSelector: [{ scheme: 'file', language: 'selene' }, { scheme: 'untitled', language: 'selene' }],
        synchronize: {
            fileEvents: workspace.createFileSystemWatcher('**/*.selene'),
        },
    };

    client = new LanguageClient('seleneLanguageServer', 'Selene Language Server', serverOptions, clientOptions);

    try {
        await client.start();
    } catch (err) {
        const message = err && err.message ? err.message : String(err);
        window.showErrorMessage(`Failed to start Selene language server: ${message}`);
        client = undefined;
    }
}

async function restartClient(context) {
    await startClient(context);
    window.showInformationMessage('Selene language server restarted');
}

function activate(context) {
    context.subscriptions.push(commands.registerCommand('selene.restartLanguageServer', async () => {
        await restartClient(context);
    }));

    context.subscriptions.push(workspace.onDidChangeConfiguration(async (event) => {
        if (event.affectsConfiguration('selene.languageServerPath') ||
            event.affectsConfiguration('selene.languageServerArgs') ||
            event.affectsConfiguration('selene.languageServerEnv')) {
            await restartClient(context);
        }
    }));

    startClient(context);
}

async function deactivate() {
    if (!client) {
        return;
    }
    const current = client;
    client = undefined;
    await current.stop();
}

module.exports = {
    activate,
    deactivate,
};
