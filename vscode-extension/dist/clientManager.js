"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.SeleneClientManager = void 0;
const configuration_1 = require("./configuration");
class SeleneClientManager {
    constructor(environment, output, status, clientFactory) {
        this.environment = environment;
        this.output = output;
        this.status = status;
        this.clientFactory = clientFactory;
        this.disposables = [];
        this.restarting = false;
    }
    async activate() {
        this.status.text = 'Selene LSP: starting…';
        this.status.tooltip = 'Starting the Selene language server';
        this.status.command = 'selene.restartLanguageServer';
        this.status.show();
        this.watcher = this.environment.workspace.createFileSystemWatcher('**/*.selene');
        this.disposables.push(this.watcher);
        this.disposables.push(this.environment.workspace.onDidChangeConfiguration(async (event) => {
            if (this.isLanguageServerConfigurationChange(event)) {
                await this.restart();
            }
        }));
        this.disposables.push(this.environment.workspace.onDidChangeWorkspaceFolders(async () => {
            await this.restart();
        }));
        await this.startClient();
    }
    async restart(options = {}) {
        if (this.restarting) {
            return;
        }
        this.restarting = true;
        try {
            await this.stopClient();
            await this.startClient();
            if (options.announce) {
                await this.environment.window.showInformationMessage('Selene language server restarted');
            }
        }
        finally {
            this.restarting = false;
        }
    }
    async dispose() {
        await this.stopClient();
        for (const disposable of this.disposables.splice(0, this.disposables.length)) {
            try {
                disposable.dispose();
            }
            catch (error) {
                this.output.appendLine(`[Selene] Failed to dispose resource: ${String(error)}`);
            }
        }
        this.status.text = 'Selene LSP: stopped';
        this.status.tooltip = 'Selene language server is not running';
    }
    async startClient() {
        const launch = this.resolveLaunchConfiguration();
        const serverOptions = {
            command: launch.command,
            args: launch.args,
            options: {
                env: launch.env,
                cwd: launch.cwd,
            },
        };
        const clientOptions = {
            documentSelector: [
                { scheme: 'file', language: 'selene' },
                { scheme: 'untitled', language: 'selene' },
            ],
            synchronize: this.watcher
                ? {
                    fileEvents: this.watcher,
                }
                : undefined,
            outputChannel: this.output,
        };
        const client = this.clientFactory({
            id: 'seleneLanguageServer',
            name: 'Selene Language Server',
            serverOptions,
            clientOptions,
        });
        this.client?.stateSubscription?.dispose();
        const stateSubscription = client.onDidChangeState((event) => {
            this.handleStateChange(client, event);
        });
        this.client = { instance: client, ready: false, stateSubscription };
        this.status.text = 'Selene LSP: starting…';
        this.status.tooltip = 'Starting the Selene language server';
        try {
            const startup = this.waitForRunning(client);
            await client.start();
            await startup;
            if (!this.client || this.client.instance !== client) {
                stateSubscription.dispose();
                return;
            }
            this.client.ready = true;
            this.status.text = 'Selene LSP: ready';
            this.status.tooltip = 'Selene language server is running';
            this.output.appendLine('[Selene] Language server ready');
        }
        catch (error) {
            if (this.client && this.client.instance === client) {
                this.client.ready = false;
            }
            const message = error instanceof Error ? error.message : String(error);
            this.status.text = 'Selene LSP: failed';
            this.status.tooltip = 'Failed to start the Selene language server';
            this.output.appendLine(`[Selene] Failed to start language server: ${message}`);
            await this.environment.window.showErrorMessage(`Failed to start Selene language server: ${message}`);
            stateSubscription.dispose();
            await this.stopClient({ updateStatus: false });
        }
    }
    async stopClient(options = {}) {
        if (!this.client) {
            return;
        }
        const { instance, stateSubscription } = this.client;
        this.client = undefined;
        stateSubscription?.dispose();
        try {
            await instance.stop();
        }
        catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            this.output.appendLine(`[Selene] Failed to stop language server: ${message}`);
        }
        if (options.updateStatus !== false) {
            this.status.text = 'Selene LSP: stopped';
            this.status.tooltip = 'Selene language server is not running';
        }
    }
    isLanguageServerConfigurationChange(event) {
        return (event.affectsConfiguration('selene.languageServerPath') ||
            event.affectsConfiguration('selene.languageServerArgs') ||
            event.affectsConfiguration('selene.languageServerEnv'));
    }
    resolveLaunchConfiguration() {
        const folders = this.environment.workspace.workspaceFolders ?? [];
        const primaryFolder = folders.length > 0 ? folders[0] : undefined;
        const configuration = this.environment.workspace.getConfiguration('selene', primaryFolder ?? null);
        return (0, configuration_1.resolveLaunchConfiguration)(configuration, primaryFolder);
    }
    handleStateChange(client, event) {
        if (!this.client || this.client.instance !== client) {
            return;
        }
        switch (event.newState) {
            case 2 /* ClientState.Running */:
                this.client.ready = true;
                this.status.text = 'Selene LSP: ready';
                this.status.tooltip = 'Selene language server is running';
                break;
            case 3 /* ClientState.Starting */:
                this.client.ready = false;
                this.status.text = 'Selene LSP: starting…';
                this.status.tooltip = 'Starting the Selene language server';
                break;
            case 1 /* ClientState.Stopped */:
                this.client.ready = false;
                this.status.text = 'Selene LSP: stopped';
                this.status.tooltip = 'Selene language server is not running';
                break;
            default:
                break;
        }
    }
    waitForRunning(client) {
        if (client.isRunning()) {
            return Promise.resolve();
        }
        return new Promise((resolve, reject) => {
            const disposable = client.onDidChangeState((event) => {
                if (event.newState === 2 /* ClientState.Running */) {
                    disposable.dispose();
                    resolve();
                }
                else if (event.newState === 1 /* ClientState.Stopped */) {
                    disposable.dispose();
                    reject(new Error('Language server stopped during startup'));
                }
            });
        });
    }
}
exports.SeleneClientManager = SeleneClientManager;
//# sourceMappingURL=clientManager.js.map