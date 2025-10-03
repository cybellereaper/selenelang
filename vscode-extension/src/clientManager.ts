import type {
  ConfigurationChangeEvent,
  Disposable,
  FileSystemWatcher,
  OutputChannel,
  StatusBarItem,
  WorkspaceConfiguration,
  WorkspaceFolder,
  WorkspaceFoldersChangeEvent,
} from 'vscode';

import type { LaunchConfiguration, ConfigurationAccessor } from './configuration';
import { resolveLaunchConfiguration } from './configuration';
import type { CreateClientOptions, LanguageClientLike } from './languageClientFactory';

interface DisposableLike {
  dispose(): void;
}

interface StateChangeEventLike {
  oldState: number;
  newState: number;
}

const enum ClientState {
  Stopped = 1,
  Running = 2,
  Starting = 3,
}

export interface WorkspaceLike {
  workspaceFolders: readonly WorkspaceFolder[] | undefined;
  getConfiguration(section: string, scope?: WorkspaceFolder | null): WorkspaceConfiguration;
  onDidChangeConfiguration(listener: (event: ConfigurationChangeEvent) => void): Disposable;
  onDidChangeWorkspaceFolders(listener: (event: WorkspaceFoldersChangeEvent) => void): Disposable;
  createFileSystemWatcher(globPattern: string): FileSystemWatcher;
}

export interface WindowLike {
  showErrorMessage(message: string): Thenable<void>;
  showInformationMessage(message: string): Thenable<void>;
}

export interface ManagerEnvironment {
  workspace: WorkspaceLike;
  window: WindowLike;
}

export interface RestartOptions {
  readonly announce?: boolean;
}

export interface ClientFactory {
  (options: CreateClientOptions): LanguageClientLike;
}

interface ManagedClient {
  instance: LanguageClientLike;
  ready: boolean;
  stateSubscription?: DisposableLike;
}

export class SeleneClientManager {
  private client: ManagedClient | undefined;
  private watcher: FileSystemWatcher | undefined;
  private disposables: Disposable[] = [];
  private restarting = false;

  constructor(
    private readonly environment: ManagerEnvironment,
    private readonly output: OutputChannel,
    private readonly status: StatusBarItem,
    private readonly clientFactory: ClientFactory,
  ) {}

  async activate(): Promise<void> {
    this.status.text = 'Selene LSP: starting…';
    this.status.tooltip = 'Starting the Selene language server';
    this.status.command = 'selene.restartLanguageServer';
    this.status.show();

    this.watcher = this.environment.workspace.createFileSystemWatcher('**/*.selene');
    this.disposables.push(this.watcher);

    this.disposables.push(
      this.environment.workspace.onDidChangeConfiguration(async (event) => {
        if (this.isLanguageServerConfigurationChange(event)) {
          await this.restart();
        }
      }),
    );

    this.disposables.push(
      this.environment.workspace.onDidChangeWorkspaceFolders(async () => {
        await this.restart();
      }),
    );

    await this.startClient();
  }

  async restart(options: RestartOptions = {}): Promise<void> {
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
    } finally {
      this.restarting = false;
    }
  }

  async dispose(): Promise<void> {
    await this.stopClient();
    for (const disposable of this.disposables.splice(0, this.disposables.length)) {
      try {
        disposable.dispose();
      } catch (error) {
        this.output.appendLine(`[Selene] Failed to dispose resource: ${String(error)}`);
      }
    }
    this.status.text = 'Selene LSP: stopped';
    this.status.tooltip = 'Selene language server is not running';
  }

  private async startClient(): Promise<void> {
    const launch = this.resolveLaunchConfiguration();
    const serverOptions: CreateClientOptions['serverOptions'] = {
      command: launch.command,
      args: launch.args,
      options: {
        env: launch.env,
        cwd: launch.cwd,
      },
    };

    const clientOptions: CreateClientOptions['clientOptions'] = {
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
    } catch (error) {
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

  private async stopClient(options: { updateStatus?: boolean } = {}): Promise<void> {
    if (!this.client) {
      return;
    }
    const { instance, stateSubscription } = this.client;
    this.client = undefined;
    stateSubscription?.dispose();
    try {
      await instance.stop();
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      this.output.appendLine(`[Selene] Failed to stop language server: ${message}`);
    }
    if (options.updateStatus !== false) {
      this.status.text = 'Selene LSP: stopped';
      this.status.tooltip = 'Selene language server is not running';
    }
  }

  private isLanguageServerConfigurationChange(event: ConfigurationChangeEvent): boolean {
    return (
      event.affectsConfiguration('selene.languageServerPath') ||
      event.affectsConfiguration('selene.languageServerArgs') ||
      event.affectsConfiguration('selene.languageServerEnv')
    );
  }

  private resolveLaunchConfiguration(): LaunchConfiguration {
    const folders = this.environment.workspace.workspaceFolders ?? [];
    const primaryFolder = folders.length > 0 ? folders[0] : undefined;
    const configuration = this.environment.workspace.getConfiguration('selene', primaryFolder ?? null) as WorkspaceConfiguration & ConfigurationAccessor;
    return resolveLaunchConfiguration(configuration, primaryFolder);
  }

  private handleStateChange(client: LanguageClientLike, event: StateChangeEventLike): void {
    if (!this.client || this.client.instance !== client) {
      return;
    }
    switch (event.newState) {
      case ClientState.Running:
        this.client.ready = true;
        this.status.text = 'Selene LSP: ready';
        this.status.tooltip = 'Selene language server is running';
        break;
      case ClientState.Starting:
        this.client.ready = false;
        this.status.text = 'Selene LSP: starting…';
        this.status.tooltip = 'Starting the Selene language server';
        break;
      case ClientState.Stopped:
        this.client.ready = false;
        this.status.text = 'Selene LSP: stopped';
        this.status.tooltip = 'Selene language server is not running';
        break;
      default:
        break;
    }
  }

  private waitForRunning(client: LanguageClientLike): Promise<void> {
    if (client.isRunning()) {
      return Promise.resolve();
    }
    return new Promise((resolve, reject) => {
      const disposable = client.onDidChangeState((event) => {
        if (event.newState === ClientState.Running) {
          disposable.dispose();
          resolve();
        } else if (event.newState === ClientState.Stopped) {
          disposable.dispose();
          reject(new Error('Language server stopped during startup'));
        }
      });
    });
  }
}
