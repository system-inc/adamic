#!/usr/bin/env node
'use strict';
// Typecheck the .a witness with stock TypeScript, then observe its erased source.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const assert = require('node:assert/strict');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
assert.equal(ts.version, '6.0.3');
const args = process.argv.slice(2);
const name = args.includes('--node-flags') ? 'node-flags' : 'text-range';
assert(args.every(a => ['--node-flags', '--mutant'].includes(a)));
let source = fs.readFileSync(path.join(__dirname, `language-questions/${name}.a`), 'utf8');
if (args.includes('--mutant')) source = name === 'text-range' ?
    source.replace('(range as TextRange).pos = pos;', '(range as TextRange).pos = 0;') :
    source.replace('(node as Mutable<Node>).flags |= 1;', '(node as Mutable<Node>).flags |= 0;');
const file = path.resolve('/unit71-witness/main.ts');
const options = { strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
    verbatimModuleSyntax: true, module: ts.ModuleKind.ESNext, target: ts.ScriptTarget.ES2024,
    lib: ['lib.es2024.d.ts'], types: [] };
const host = ts.createCompilerHost(options), get = host.getSourceFile;
host.getSourceFile = (name, version, onError, fresh) => name === file ?
    ts.createSourceFile(name, source, version, true, ts.ScriptKind.TS) : get(name, version, onError, fresh);
let javascript;
host.writeFile = (name, text) => { if (name.endsWith('main.js')) javascript = text; };
const program = ts.createProgram([file], options, host);
const diagnostics = ts.getPreEmitDiagnostics(program);
assert.equal(diagnostics.length, 0, diagnostics.map(d => ts.flattenDiagnosticMessageText(d.messageText, '\n')).join('\n'));
assert.equal(program.emit().emitSkipped, false);
const output = [];
vm.runInNewContext(javascript, { console: { log: value => output.push(value) } });
assert.deepEqual(output, ['1']);
console.log(JSON.stringify({ diagnostics: 0, promised_type: 0, output, source_write: name === 'text-range' ? 'utilities.ts:10645:5' : 'binder.ts:1113:17' }, null, 2));
