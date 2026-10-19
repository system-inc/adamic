'use strict';
// Validate the readable table against the saved, independently captured tsc CLI stream.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const tablePath = path.resolve(process.argv[2] || path.join(__dirname, 'checker-stops-2026-10-19.json'));
const table = JSON.parse(fs.readFileSync(tablePath, 'utf8'));
const evidence = path.dirname(tablePath);
const records = new Map(table.diagnostics.map(diagnostic => [diagnostic.id, diagnostic]));
const classes = new Map(table.classes.map(entry => [entry.id, entry]));
assert.equal(records.size, table.diagnostics.length, 'duplicate diagnostic ID');
const counts = values => Object.fromEntries([...values.reduce((result, value) => result.set(value, (result.get(value) || 0) + 1), new Map())].sort());
const builder = diagnostic => diagnostic.file === 'src/compiler/builder.ts' && diagnostic.line === 1246 && diagnostic.column === 69 && diagnostic.code === 'TS2345';
for (const run of table.runs) {
    const diagnostics = run.diagnosticIds.map(id => {
        assert(records.has(id), 'unknown diagnostic ID ' + id);
        const diagnostic = records.get(id);
        assert(classes.has(diagnostic.classification), 'unknown class ' + id);
        assert(diagnostic.observedIn.includes(run.name), 'missing observation ' + id);
        return diagnostic;
    });
    const text = diagnostics.map(diagnostic => `${diagnostic.file}(${diagnostic.line},${diagnostic.column}): error ${diagnostic.code}: ${diagnostic.message}\n`).join('');
    const log = fs.readFileSync(path.join(evidence, run.log), 'utf8');
    assert.equal(text, log, run.name + ': table differs from tsc CLI diagnostics');
    assert.equal(crypto.createHash('sha256').update(log).digest('hex'), run.logSha256, run.name + ': CLI log changed');
    assert.equal(diagnostics.length, run.diagnosticCount, run.name + ': diagnostic count');
    assert.deepEqual(counts(diagnostics.map(diagnostic => diagnostic.classification)), run.countsByClass, run.name + ': class counts');
    assert.deepEqual(counts(diagnostics.map(diagnostic => diagnostic.code)), run.countsByCode, run.name + ': code counts');
    assert.equal(run.exit, diagnostics.length ? 2 : 0, run.name + ': tsc exit');
    assert.equal(run.rootCount, table.roots.length, run.name + ': root count');
    const sources = new Set(run.sourceFiles.map(source => source.file));
    for (const root of table.roots) assert(sources.has(root), run.name + ': omitted source root ' + root);
    assert(!run.sourceFiles.some(source => source.file.includes('/@types/node/')), run.name + ': Node types loaded');
    for (const entry of table.classes) assert.equal(entry.countsByRun[run.name], run.countsByClass[entry.id] || 0, run.name + ': class summary ' + entry.id);
    if (run.name.includes('filtered-prelude') || run.name.endsWith('no-prelude')) assert(!diagnostics.some(builder), run.name + ': excluded iterator defect returned');
}
for (const [left, right] of [['stock-no-prelude', 'loader-no-prelude'], ['stock-filtered-prelude', 'loader-filtered-prelude']]) {
    assert.deepEqual(table.runs.find(run => run.name === left).diagnosticIds, table.runs.find(run => run.name === right).diagnosticIds, 'library profiles disagree');
}
const required = table.runs.find(run => run.name === (process.argv[3] || 'loader-filtered-prelude'));
assert(required, 'unknown required run');
assert(!required.diagnosticIds.map(id => records.get(id)).some(builder), 'required filtered census includes builder.ts:1246:69 TS2345');
const mutant = table.runs.find(run => run.name === 'loader-original-prelude-mutant');
assert(mutant.diagnosticIds.map(id => records.get(id)).some(builder), 'restored-next mutant did not reproduce builder diagnostic');
console.log(`Verified ${table.runs.length} CLI streams, ${table.diagnostics.length} unique diagnostics, every class and root; filtered prelude excludes the builder defect.`);
