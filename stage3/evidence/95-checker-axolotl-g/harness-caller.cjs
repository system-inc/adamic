'use strict';
const fs = require('node:fs'), path = require('node:path'), vm = require('node:vm'), parser = require('typescript');
const ts = require(path.resolve(process.argv[2]));
const sourcePath = process.argv[3], source = parser.createSourceFile(sourcePath, fs.readFileSync(sourcePath, 'utf8'), parser.ScriptTarget.Latest, true), methods = [];
function locate(n) { if (parser.isMethodDeclaration(n) && n.name.getText(source) === 'verifySymbolAtLocation') methods.push(n); parser.forEachChild(n, locate); }
locate(source); if (methods.length !== 1) throw new Error('harness method shape moved');
const method = methods[0].getText(source), emitted = parser.transpileModule('class Probe { ' + method + ' }; Probe;', {compilerOptions: {target: parser.ScriptTarget.ES2024}}).outputText;
const Probe = vm.runInNewContext(emitted), result = {methodSource: method, diagnostics: [], calls: []};
const file = '/95g/consumer.ts', data = new Map([[file, "import v = require('./exported');\nv.xyz;"], ['/95g/exported.js', "/** @type {string} */ let key = 'xyz'; class Computed { static [key] = 1; } module.exports = /** @type {typeof Computed} */ (class {}); module.exports.extra = 2;"]]);
const options = {strict: true, allowJs: true, checkJs: true, noEmit: true, target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.CommonJS, moduleResolution: ts.ModuleResolutionKind.Node10, ignoreDeprecations: '6.0'};
const host = ts.createCompilerHost(options), original = host.getSourceFile, read = host.readFile, exists = host.fileExists;
host.getSourceFile = (f, ...args) => data.has(f) ? ts.createSourceFile(f, data.get(f), ts.ScriptTarget.Latest, true) : original.call(host, f, ...args);
host.readFile = f => data.has(f) ? data.get(f) : read.call(host, f); host.fileExists = f => data.has(f) || exists.call(host, f); host.directoryExists = d => d === '/95g' || ts.sys.directoryExists(d);
const program = ts.createProgram([file], options, host), checker = program.getTypeChecker(), node = program.getSourceFile(file).statements[1].expression;
result.diagnostics = ts.getPreEmitDiagnostics(program).map(d => ({code: d.code, message: ts.flattenDiagnosticMessageText(d.messageText, '\n')}));
try {
    Probe.prototype.verifySymbolAtLocation.call({goToAndGetNode: () => node, getChecker: () => checker, raiseError(message) {result.calls.push({method: 'raiseError', message}); throw new Error(message);}, verifySymbol() {result.calls.push({method: 'verifySymbol'});}}, {}, []);
    result.result = 'returned';
} catch (error) {result.error = error.name + ': ' + error.message; result.stack = error.stack;}
console.log(JSON.stringify(result, null, 2));
