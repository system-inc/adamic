'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {dump} = require('./binder-dump.cjs');
const api = require(process.env.STEP31_TYPESCRIPT);
const request = JSON.parse(fs.readFileSync(path.join(__dirname, 'fixtures/projects.json'), 'utf8'));
for (const file of request.projects[0].files) assert.equal(file.text, fs.readFileSync(path.join(__dirname, 'fixtures', file.path), 'utf8'));
const output = dump(api, request);
assert.equal(output, fs.readFileSync(path.join(__dirname, 'fixtures/golden.jsonl'), 'utf8'));
assert.equal(dump(api, request), output, 'allocation IDs must not enter output');
const reversed = structuredClone(request);
reversed.projects[0].files.reverse();
assert.equal(dump(api, reversed), output, 'file ordering must be stable');
const rows = output.trimEnd().split('\n').map(JSON.parse);
const files = new Map(rows.filter(row => row.record === 'file').map(row => [row.path, row]));
function symbol(file, name, flags) {
    const matches = files.get(file).symbols.filter(item => item.name === name && item.flags === flags);
    assert.ok(matches.length > 0, `${file}: ${name} must have flags ${flags}`);
    return matches[0];
}
symbol('scopes.a', 'outer', api.SymbolFlags.Function);
symbol('scopes.a', 'delta', api.SymbolFlags.BlockScopedVariable);
symbol('members.a', 'Counter', api.SymbolFlags.Class);
assert.ok(symbol('members.a', 'Counter', api.SymbolFlags.Class).members.some(([name]) => name === 'increment'));
symbol('aliases.a', 'ImportedCounter', api.SymbolFlags.Alias);
assert.ok(files.get('scopes.a').nodes.some(node => node.locals?.some(([name]) => name === 'delta')));
for (const row of files.values()) {
    assert.deepEqual(row.parseDiagnostics, []);
    assert.deepEqual(row.bindDiagnostics, []);
}
const diagnosticRequest = {projects: [{id: 'diagnostics', options: {strict: true}, files: [{path: 'duplicate.ts', text: 'export {}; const value = 1; const value = 2;'}]}]};
const diagnosticRows = dump(api, diagnosticRequest).trimEnd().split('\n').map(JSON.parse);
assert.deepEqual(diagnosticRows[1].bindDiagnostics.map(item => item.code), [2451, 2451]);
assert.ok(diagnosticRows[1].bindDiagnostics.every(item => item.start !== null && item.length > 0));
const changed = Buffer.from(output);
changed[changed.indexOf(Buffer.from('"flags":')) + 8] ^= 1;
assert.notDeepEqual(changed, Buffer.from(output), 'one-byte output mutant must fail comparison');
assert.throws(() => dump({...api, version: 'wrong'}, request), /expected TypeScript/);
console.log('binder fixture goldens, flags, members, scope, diagnostics, ordering, repeatability and one-byte mutant pass');
