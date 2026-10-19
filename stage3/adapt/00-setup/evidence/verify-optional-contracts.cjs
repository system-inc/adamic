'use strict';
// Evidence checks use the census and stock parser as independent inputs.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('../../../api/node_modules/typescript');
const directory = __dirname;
const table = JSON.parse(fs.readFileSync(process.argv[2] || path.join(directory, 'optional-contracts-2026-10-19.json')));
const census = JSON.parse(fs.readFileSync(path.join(directory, 'checker-stops-2026-10-19.json')));
const expected = census.diagnostics.filter(row => row.classification === 'optional-value-contract');
assert.equal(table.count, 62);
assert.equal(table.rows.length, 62);
assert.deepEqual(table.rows.map(row => row.id).sort(), expected.map(row => row.id).sort());
const grouped = table.groups.flatMap(group => {
    assert.equal(group.count, group.ids.length);
    for (const id of group.ids) {
        const row = table.rows.find(row => row.id === id);
        assert(row);
        assert.equal(row.reasonGroup, group.reason);
        assert.equal(row.owners.some(owner => owner.publicApi), group.visibility === 'public');
        assert.equal(row.owners.some(owner => owner.kind === 'MethodSignature'), group.form === 'method');
    }
    return group.ids;
});
assert.deepEqual(grouped.sort(), expected.map(row => row.id).sort());
const ledger = JSON.parse(fs.readFileSync(path.join(directory, 'optional-contracts-20-before-2026-10-19.json')));
for (const row of table.rows) {
    const original = expected.find(item => item.id === row.id);
    for (const key of ['file', 'line', 'column', 'code', 'message', 'sourceLine', 'observedIn']) assert.deepEqual(row[key], original[key]);
    assert(row.owners.length);
    for (const key of ['declined', 'initially_declined']) {
        for (const entry of row.adaptation20[key]) assert(ledger[key].some(actual => JSON.stringify(actual) === JSON.stringify(entry)));
    }
}
function ownerPath(node) {
    const names = [];
    for (let current = node; current; current = current.parent) {
        if (current.name && !(ts.isModuleDeclaration(current) && current.name.getText() === 'ts')) names.unshift(current.name.getText());
    }
    return names;
}
function members(file) {
    const source = ts.createSourceFile(file, fs.readFileSync(file, 'utf8'), ts.ScriptTarget.Latest, true);
    const result = [];
    function visit(node) {
        if (ts.isPropertySignature(node) || ts.isPropertyDeclaration(node) || ts.isMethodSignature(node)) {
            result.push({ node, line: source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1,
                kind: ts.SyntaxKind[node.kind], path: ownerPath(node) });
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
    return result;
}
if (process.argv[3]) {
    const tree = path.resolve(process.argv[3]);
    const api = members(path.join(tree, 'built/local/typescript.d.ts'));
    const files = new Map();
    for (const row of table.rows) for (const owner of row.owners) {
        if (!files.has(owner.file)) files.set(owner.file, members(path.join(tree, owner.file)));
        const candidates = files.get(owner.file).filter(entry => entry.line === owner.line && entry.kind === owner.kind && JSON.stringify(entry.path) === JSON.stringify(owner.path));
        assert.equal(candidates.length, 1, `${row.id}: ${owner.file}:${owner.line}`);
        assert.equal(candidates[0].node.getText().slice(0, 400), owner.text);
        const publicEntries = api.filter(entry => entry.kind === owner.kind && JSON.stringify(entry.path) === JSON.stringify(owner.path));
        assert.equal(publicEntries.length > 0, owner.publicApi, `${row.id}: API owner`);
        assert.deepEqual(publicEntries.map(entry => entry.line), owner.publicApiLines);
    }
}
console.log('62 exact census rows; complete disjoint groups; recorded decline entries; owner/API checks passed');
