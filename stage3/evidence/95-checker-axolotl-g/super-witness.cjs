'use strict';
const path = require('node:path');
const ts = require(path.resolve(process.argv[2]));
const name = '/95g/super.js';
const text = '/** @type {string} */\nlet key = "xyz";\nclass Computed { static [key] = 1; }\nclass Derived extends Computed { static read() { return super.xyz; } }\n';
const options = {allowJs: true, checkJs: true, noEmit: true, strict: true, target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.CommonJS, moduleResolution: ts.ModuleResolutionKind.Node10, ignoreDeprecations: '6.0'};
const host = {
    getScriptFileNames: () => [name], getScriptVersion: () => '1',
    getScriptSnapshot: file => { const data = file === name ? text : ts.sys.readFile(file); return data === undefined ? undefined : ts.ScriptSnapshot.fromString(data); },
    getCurrentDirectory: () => '/95g', getCompilationSettings: () => options,
    getDefaultLibFileName: ts.getDefaultLibFilePath,
    fileExists: file => file === name || ts.sys.fileExists(file), readFile: file => file === name ? text : ts.sys.readFile(file),
    readDirectory: ts.sys.readDirectory, directoryExists: dir => dir === '/95g' || ts.sys.directoryExists(dir),
};
const service = ts.createLanguageService(host), result = {text};
result.diagnostics = [...service.getSyntacticDiagnostics(name), ...service.getSemanticDiagnostics(name)].map(d => ({code: d.code, message: ts.flattenDiagnosticMessageText(d.messageText, '\n')}));
for (const token of ['super', 'xyz']) {
    const position = token === 'super' ? text.lastIndexOf('super') + 1 : text.lastIndexOf('xyz') + 1;
    try { const value = service.getImplementationAtPosition(name, position); result[token] = value === undefined ? 'undefined' : value; }
    catch (error) { result[token] = {error: error.name + ': ' + error.message, stack: error.stack}; }
}
console.log(JSON.stringify(result, null, 2));
