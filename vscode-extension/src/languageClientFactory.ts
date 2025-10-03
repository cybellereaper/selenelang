import { LanguageClient, TransportKind } from 'vscode-languageclient/node';
import type { Executable, LanguageClientOptions } from 'vscode-languageclient/node';
import type { Disposable } from 'vscode';

export interface CreateClientOptions {
  id: string;
  name: string;
  serverOptions: Executable;
  clientOptions: LanguageClientOptions;
}

export interface LanguageClientLike {
  start(): Promise<void>;
  stop(): Thenable<void>;
  isRunning(): boolean;
  onDidChangeState(
    listener: (event: { oldState: number; newState: number }) => void,
    thisArgs?: unknown,
    disposables?: Disposable[],
  ): Disposable;
}

export function createLanguageClient(options: CreateClientOptions): LanguageClientLike {
  const serverOptions: Executable = {
    ...options.serverOptions,
    transport: TransportKind.stdio,
  };

  return new LanguageClient(options.id, options.name, serverOptions, options.clientOptions);
}
