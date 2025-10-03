import { expect } from 'chai';
import type { WorkspaceFolder } from 'vscode';

import { resolveLaunchConfiguration, type ConfigurationAccessor } from '../configuration';

function createConfiguration(values: Record<string, unknown>): ConfigurationAccessor {
  return {
    get<T>(section: string, defaultValue: T): T {
      if (Object.prototype.hasOwnProperty.call(values, section)) {
        return values[section] as T;
      }
      return defaultValue;
    },
  };
}

describe('resolveLaunchConfiguration', () => {
  const fakeFolder: WorkspaceFolder = {
    index: 0,
    name: 'workspace',
    uri: { fsPath: '/workspace/selene', toString: () => 'file:///workspace/selene' } as unknown as WorkspaceFolder['uri'],
  };

  it('uses defaults when configuration is missing', () => {
    const config = resolveLaunchConfiguration(createConfiguration({}), fakeFolder);
    expect(config.command).to.equal('selene');
    expect(config.args).to.deep.equal(['lsp']);
    expect(config.cwd).to.equal('/workspace/selene');
    expect(config.env).to.not.equal(process.env);
  });

  it('trims and validates command', () => {
    const config = resolveLaunchConfiguration(createConfiguration({ languageServerPath: '  ./selene-cli  ' }), fakeFolder);
    expect(config.command).to.equal('./selene-cli');
  });

  it('normalizes arguments', () => {
    const config = resolveLaunchConfiguration(
      createConfiguration({ languageServerArgs: ['  --stdio  ', null, 42, ''] }),
      null,
    );
    expect(config.args).to.deep.equal(['--stdio', '42']);
  });

  it('falls back to default arguments when invalid', () => {
    const config = resolveLaunchConfiguration(createConfiguration({ languageServerArgs: 'oops' }), null);
    expect(config.args).to.deep.equal(['lsp']);
  });

  it('merges environment variables from configuration', () => {
    const config = resolveLaunchConfiguration(
      createConfiguration({ languageServerEnv: { FOO: 'bar', EMPTY: null } }),
      fakeFolder,
    );
    expect(config.env.FOO).to.equal('bar');
    expect(config.env).to.not.have.property('EMPTY');
  });
});
