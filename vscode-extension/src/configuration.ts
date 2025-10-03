import type { WorkspaceFolder } from 'vscode';

export interface ConfigurationAccessor {
  get<T>(section: string, defaultValue: T): T;
}

export interface LaunchConfiguration {
  command: string;
  args: string[];
  env: NodeJS.ProcessEnv;
  cwd?: string;
}

function coerceStringArray(value: unknown, fallback: string[]): string[] {
  if (!Array.isArray(value)) {
    return [...fallback];
  }
  const coerced = value
    .map((item) => {
      if (item === undefined || item === null) {
        return '';
      }
      return String(item).trim();
    })
    .filter((item) => item.length > 0);
  return coerced.length > 0 ? coerced : [...fallback];
}

function coerceEnvironment(value: unknown, base: NodeJS.ProcessEnv): NodeJS.ProcessEnv {
  const result: NodeJS.ProcessEnv = { ...base };
  if (value && typeof value === 'object') {
    for (const [key, raw] of Object.entries(value as Record<string, unknown>)) {
      if (!key || raw === undefined || raw === null) {
        continue;
      }
      result[key] = String(raw);
    }
  }
  return result;
}

export function resolveLaunchConfiguration(
  configuration: ConfigurationAccessor,
  folder?: WorkspaceFolder | null,
): LaunchConfiguration {
  const rawCommand = configuration.get('languageServerPath', 'selene');
  const command = typeof rawCommand === 'string' && rawCommand.trim().length > 0
    ? rawCommand.trim()
    : 'selene';

  const args = coerceStringArray(configuration.get('languageServerArgs', ['lsp']), ['lsp']);
  const env = coerceEnvironment(configuration.get('languageServerEnv', {}), process.env);

  const cwd = folder?.uri.fsPath;

  return {
    command,
    args,
    env,
    cwd,
  };
}
