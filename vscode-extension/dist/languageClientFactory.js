"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.createLanguageClient = createLanguageClient;
const node_1 = require("vscode-languageclient/node");
function createLanguageClient(options) {
    const serverOptions = {
        ...options.serverOptions,
        transport: node_1.TransportKind.stdio,
    };
    return new node_1.LanguageClient(options.id, options.name, serverOptions, options.clientOptions);
}
//# sourceMappingURL=languageClientFactory.js.map