import { expect } from 'chai';
import sinon from 'sinon';

import { SeleneClientManager } from '../clientManager';
import type { LaunchConfiguration } from '../configuration';
import type { LanguageClientLike } from '../languageClientFactory';
import type { ConfigurationChangeEvent, OutputChannel, StatusBarItem, WorkspaceFolder } from 'vscode';

type StateChangeListener = (event: { oldState: number; newState: number }) => void;

class FakeClient implements LanguageClientLike {
  public started = false;
  public stopped = false;
  private running = false;
  private listeners: StateChangeListener[] = [];

  constructor(private readonly shouldFail = false) {}

  async start(): Promise<void> {
    this.started = true;
    this.emit({ oldState: 1, newState: 3 });
    if (this.shouldFail) {
      this.emit({ oldState: 3, newState: 1 });
      return;
    }
    this.running = true;
    this.emit({ oldState: 3, newState: 2 });
  }

  isRunning(): boolean {
    return this.running;
  }

  onDidChangeState(listener: StateChangeListener): { dispose(): void } {
    this.listeners.push(listener);
    return {
      dispose: () => {
        this.listeners = this.listeners.filter((candidate) => candidate !== listener);
      },
    };
  }

  async stop(): Promise<void> {
    this.stopped = true;
    if (this.running) {
      this.running = false;
      this.emit({ oldState: 2, newState: 1 });
    }
  }

  private emit(event: { oldState: number; newState: number }): void {
    for (const listener of [...this.listeners]) {
      listener(event);
    }
  }
}

function createWorkspaceFolder(path: string): WorkspaceFolder {
  const uri = { fsPath: path, toString: () => `file://${path}` } as unknown as WorkspaceFolder['uri'];
  return { name: 'root', index: 0, uri };
}

function createEnvironment(overrides: Partial<LaunchConfiguration> = {}) {
  const workspaceFolders: WorkspaceFolder[] = overrides.cwd
    ? [createWorkspaceFolder(overrides.cwd)]
    : [];
  const configuration = {
    get<T>(key: string, defaultValue: T): T {
      switch (key) {
        case 'languageServerPath':
          return (overrides.command as T) ?? defaultValue;
        case 'languageServerArgs':
          return (overrides.args as T) ?? defaultValue;
        case 'languageServerEnv':
          return (overrides.env as T) ?? defaultValue;
        default:
          return defaultValue;
      }
    },
  };

  const configurationStub = sinon.stub().returns(configuration);
  const onDidChangeConfiguration = sinon.stub().returns({ dispose: sinon.spy() });
  const onDidChangeWorkspaceFolders = sinon.stub().returns({ dispose: sinon.spy() });
  const fileSystemWatcher = { dispose: sinon.spy() };
  const createFileSystemWatcher = sinon.stub().returns(fileSystemWatcher);

  return {
    environment: {
      workspace: {
        workspaceFolders,
        getConfiguration: configurationStub,
        onDidChangeConfiguration,
        onDidChangeWorkspaceFolders,
        createFileSystemWatcher,
      },
      window: {
        showErrorMessage: sinon.stub().resolves(undefined),
        showInformationMessage: sinon.stub().resolves(undefined),
      },
    },
    fileSystemWatcher,
    configurationStub,
    onDidChangeConfiguration,
    onDidChangeWorkspaceFolders,
  };
}

describe('SeleneClientManager', () => {
  it('starts and stops the language client', async () => {
    const fakeClient = new FakeClient();
    const clientFactory = sinon.stub().returns(fakeClient);
    const env = createEnvironment({ cwd: '/workspace/project' });
    const output = { appendLine: sinon.spy(), show: sinon.spy() } as unknown as OutputChannel;
    const status = {
      text: '',
      tooltip: '',
      command: '',
      show: sinon.spy(),
      hide: sinon.spy(),
    } as unknown as StatusBarItem;

    const manager = new SeleneClientManager(env.environment, output, status, clientFactory);
    await manager.activate();

    expect(clientFactory.calledOnce).to.be.true;
    expect(fakeClient.started).to.be.true;
    expect(status.text).to.equal('Selene LSP: ready');

    await manager.dispose();
    expect(fakeClient.stopped).to.be.true;
    expect(status.text).to.equal('Selene LSP: stopped');
  });

  it('restarts when configuration changes', async () => {
    const fakeClient = new FakeClient();
    const clientFactory = sinon.stub().returns(fakeClient);
    const env = createEnvironment();
    const output = { appendLine: sinon.spy(), show: sinon.spy() } as unknown as OutputChannel;
    const status = {
      text: '',
      tooltip: '',
      command: '',
      show: sinon.spy(),
      hide: sinon.spy(),
    } as unknown as StatusBarItem;

    const manager = new SeleneClientManager(env.environment, output, status, clientFactory);
    await manager.activate();

    const restartSpy = sinon.spy(manager, 'restart');
    const configHandler = env.onDidChangeConfiguration.getCall(0).args[0];
    const event: ConfigurationChangeEvent = {
      affectsConfiguration: (section) => section === 'selene.languageServerPath',
    };
    await configHandler(event);

    expect(restartSpy.calledOnce).to.be.true;
  });

  it('shows error when client fails to start', async () => {
    const failingClient = new FakeClient(true);
    const clientFactory = sinon.stub().returns(failingClient);
    const env = createEnvironment();
    const output = { appendLine: sinon.spy(), show: sinon.spy() } as unknown as OutputChannel;
    const status = {
      text: '',
      tooltip: '',
      command: '',
      show: sinon.spy(),
      hide: sinon.spy(),
    } as unknown as StatusBarItem;

    const manager = new SeleneClientManager(env.environment, output, status, clientFactory);
    await manager.activate();

    expect(env.environment.window.showErrorMessage.calledOnce).to.be.true;
    expect(status.text).to.equal('Selene LSP: failed');
  });

  it('announces restart when requested', async () => {
    const clientFactory = sinon.stub().returns(new FakeClient());
    const env = createEnvironment();
    const output = { appendLine: sinon.spy(), show: sinon.spy() } as unknown as OutputChannel;
    const status = {
      text: '',
      tooltip: '',
      command: '',
      show: sinon.spy(),
      hide: sinon.spy(),
    } as unknown as StatusBarItem;

    const manager = new SeleneClientManager(env.environment, output, status, clientFactory);
    await manager.activate();

    await manager.restart({ announce: true });
    expect(env.environment.window.showInformationMessage.calledOnce).to.be.true;
  });
});
