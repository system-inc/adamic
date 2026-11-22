// Prints the declaration a TypeScript compiler emits for one JavaScript file, or fails on any diagnostic.
// Usage: node declaration.cjs <compiler module> <source file>
// The compiler's default library is the one beside it, so stock and adapted each read their own.
'use strict';

const assert = require('node:assert/strict');
const path = require('node:path');

const [compiler, source] = process.argv.slice(2).map(file => path.resolve(file));
const ts = require(compiler);
const options = {
    allowJs: true,
    checkJs: true,
    declaration: true,
    emitDeclarationOnly: true,
    target: ts.ScriptTarget.ES2020,
    module: ts.ModuleKind.CommonJS,
    types: [],
};
const program = ts.createProgram([source], options);
const diagnostics = ts.getPreEmitDiagnostics(program);
for (const diagnostic of diagnostics) {
    console.error('TS' + diagnostic.code, ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n'));
}
assert.equal(diagnostics.length, 0, 'the fixture must check clean');
const written = [];
const result = program.emit(undefined, (_file, text) => written.push(text));
assert.equal(result.emitSkipped, false);
assert.equal(written.length, 1, 'one declaration file');
process.stdout.write(written[0]);
