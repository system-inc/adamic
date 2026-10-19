'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const cp = require('node:child_process');
const ts = require('../../api/node_modules/typescript');
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'optional-return-check-'));
const compiler = path.join(scratch, 'src/compiler');
fs.mkdirSync(compiler, { recursive: true });
const fixture = fs.readFileSync(path.join(__dirname, 'testdata/nullable-return.a'), 'utf8');
const file = path.join(compiler, 'fixture.ts');
fs.writeFileSync(file, fixture);
const options = { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext };
const emit = source => ts.transpileModule(source, { compilerOptions: options, fileName: 'fixture.ts' }).outputText;
const report = JSON.parse(cp.execFileSync(process.execPath, [process.argv[2] || path.join(__dirname, 'adapt.cjs'), scratch], {
    env: { ...process.env, TSC_ADAPT_TYPESCRIPT: require.resolve('../../api/node_modules/typescript') },
    timeout: 60000, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'],
}));
assert.equal(report.declarations_changed, 1, 'only the single object nullable return owner is eligible');
assert.deepEqual(report.declarations[0].owner_path, ['ReturnStorage', 'value']);
assert(report.resolved_findings.some(row => row.owners.some(owner => owner.name === 'value')));
const after = fs.readFileSync(file, 'utf8');
assert(after.includes('interface MethodStorage { callback?(): number; }'));
assert(after.includes('interface GenericStorage<T> { value?: T; }'));
assert(after.includes('interface LeftStorage { kind: "left"; left?: number; }'));
assert(after.includes('interface ArgumentStorage { value?: number; }'));
assert.equal(emit(fixture), emit(after), 'transpile output bytes');
const originalJs = path.join(scratch, 'before.mjs'), adaptedJs = path.join(scratch, 'after.mjs');
fs.writeFileSync(originalJs, emit(fixture));
fs.writeFileSync(adaptedJs, emit(after));
const run = script => cp.execFileSync(process.execPath, [script], { timeout: 10000, encoding: 'utf8' });
assert.equal(run(originalJs), '[["value"],true,null]\n');
assert.equal(run(adaptedJs), run(originalJs));
// The byte comparator must reject an actual emitted runtime expression.
assert.notEqual(emit(fixture), emit(after + '\nconsole.log("mutant");\n'));
console.log(JSON.stringify({ status: 'pass', declarations: report.declarations,
    node: '[["value"],true,null]', runtimeExpressionMutant: 'caught by transpile byte comparator', scratch }));
