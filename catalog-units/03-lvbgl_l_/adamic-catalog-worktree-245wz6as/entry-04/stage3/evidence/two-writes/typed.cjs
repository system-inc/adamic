// Compile a virtual .ts view of the authored .a using the measured 6.0.3 API.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
require('./probe.cjs');
const ts = require(process.argv[2]);
const file = path.join(__dirname, 'typed-witness.ts');
const source = fs.readFileSync(path.join(__dirname, 'typed-witness.a'), 'utf8').replace('__TSC_PUBLIC__',
    path.join(path.dirname(process.argv[2]), 'typescript').replaceAll('\\', '/'));
const options = { strict: true, noEmit: true, target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS, types: [] };
const host = ts.createCompilerHost(options);
const read = host.readFile;
const exists = host.fileExists;
host.readFile = name => name === file ? source : read(name);
host.fileExists = name => name === file || exists(name);
let diagnostics = ts.getPreEmitDiagnostics(ts.createProgram([file], options, host));
assert.deepEqual(diagnostics.map(d => ts.flattenDiagnosticMessageText(d.messageText, '\n')), []);
const broken = source.replace('const promisedFlag: NodeFlags.Synthesized = changed.flags;', 'const promisedFlag: string = changed.flags;');
host.readFile = name => name === file ? broken : read(name);
diagnostics = ts.getPreEmitDiagnostics(ts.createProgram([file], options, host));
assert.ok(diagnostics.some(d => d.code === 2322));
console.log('legal narrowed holders: 0 diagnostics; wrong string-domain mutant: TS2322');
