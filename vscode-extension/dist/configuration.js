"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.resolveLaunchConfiguration = resolveLaunchConfiguration;
function coerceStringArray(value, fallback) {
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
function coerceEnvironment(value, base) {
    const result = { ...base };
    if (value && typeof value === 'object') {
        for (const [key, raw] of Object.entries(value)) {
            if (!key || raw === undefined || raw === null) {
                continue;
            }
            result[key] = String(raw);
        }
    }
    return result;
}
function resolveLaunchConfiguration(configuration, folder) {
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
//# sourceMappingURL=configuration.js.map