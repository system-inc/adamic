#!/usr/bin/env node
'use strict';
// A reachable counterexample to the untouched destination view's type promise.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
assert.equal(ts.version, '6.0.3');
if (process.argv.length !== 4) throw new Error('usage: node narrow-range.cjs <published.d.ts> <compiler.js>');
const input = `const compiler = (globalThis as unknown as { readonlyRangeCompiler: typeof import('./published') }).readonlyRangeCompiler;
const range = { pos: 0, end: 0 } as const;
const returned = compiler.setTextRange(range, { pos: 1, end: 2 });
const promisedPos: 0 = returned.pos;
const promisedEnd: 0 = range.end;
export const observation = [promisedPos, promisedEnd];
`;
const root = path.resolve('/readonly-range-counterexample'), main = path.join(root, 'main.ts'), published = path.join(root, 'published.d.ts');
const options = { strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
    verbatimModuleSyntax: true, erasableSyntaxOnly: true, moduleDetection: ts.ModuleDetectionKind.Force,
    module: ts.ModuleKind.ESNext, moduleResolution: ts.ModuleResolutionKind.Bundler,
    target: ts.ScriptTarget.ES2024, lib: ['lib.es2024.d.ts'], types: [] };
const host = ts.createCompilerHost(options), read = host.readFile, exists = host.fileExists, getSourceFile = host.getSourceFile;
host.readFile = name => name === main ? input : name === published ? fs.readFileSync(process.argv[2], 'utf8') : read(name);
host.fileExists = name => name === main || name === published || exists(name);
host.getSourceFile = (name, version, onError, fresh) => name === main || name === published ?
    ts.createSourceFile(name, host.readFile(name), version, true) : getSourceFile(name, version, onError, fresh);
host.resolveModuleNames = names => names.map(name => name === './published' ? { resolvedFileName: published, extension: ts.Extension.Dts } : undefined);
let javascript;
host.writeFile = (name, text) => { if (name.endsWith('main.js')) javascript = text; };
const program = ts.createProgram([main, published], options, host);
const diagnostics = ts.getPreEmitDiagnostics(program);
assert.equal(diagnostics.length, 0, diagnostics.map(d => ts.flattenDiagnosticMessageText(d.messageText, '\n')).join('\n'));
assert.equal(program.emit().emitSkipped, false);
assert(javascript);
globalThis.readonlyRangeCompiler = require(path.resolve(process.argv[3]));
assert.equal(globalThis.readonlyRangeCompiler.version, '6.0.3');
(async () => {
    try {
        const result = await import('data:text/javascript;base64,' + Buffer.from(javascript).toString('base64'));
        assert.deepEqual(result.observation, [1, 2]);
        console.log(JSON.stringify({ input, diagnostics: [], promised_types: [0, 0], observed: result.observation,
            classification: 'reachable through the public API; no narrow-position instance established among compiler-internal callers' }, null, 2));
    } finally { delete globalThis.readonlyRangeCompiler; }
})().catch(error => { process.exitCode = 1; console.error(error); });
