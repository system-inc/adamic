'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const base = __dirname;
const snapshot = JSON.parse(fs.readFileSync(path.join(base, 'evidence/after.json')));
const ledger = JSON.parse(fs.readFileSync(path.join(base, 'evidence/classification.json')));
const tree = process.argv[2];
assert(tree, 'provide the unchanged adapted upstream tree');
function coverage(data) {
    assert.deepEqual(data.sites.map(s => s.id), snapshot.census.map((_, i) => i + 1), 'site coverage');
}
function membership(data) {
    const ids = data.classes.flatMap(c => c.siteIds).sort((a, b) => a - b);
    assert.deepEqual(ids, snapshot.census.map((_, i) => i + 1), 'disjoint class membership');
    for (const c of data.classes) for (const id of c.siteIds) {
        assert.equal(data.sites[id - 1].classification, c.id, 'class assignment');
    }
}
function counts(data) {
    for (const c of data.classes) assert.equal(c.count, c.siteIds.length, 'class count');
    assert.equal(data.tokens, snapshot.census.length, 'token count');
    assert.equal(data.classes.reduce((n, c) => n + c.count, 0), data.tokens, 'class total');
    assert.equal(data.files, new Set(data.sites.map(s => s.file)).size, 'file count');
}
function locations(data) {
    data.sites.forEach((s, i) => {
        const { id, classification, ...location } = s;
        assert.deepEqual(location, snapshot.census[i], 'exact census location and source');
    });
}
const hashes = {};
for (const file of Object.keys(snapshot.sources)) {
    hashes[file] = crypto.createHash('sha256').update(fs.readFileSync(path.join(tree, file))).digest('hex');
}
function sources(actual) { assert.deepEqual(actual, snapshot.sources, 'unchanged compiler sources'); }
coverage(ledger); membership(ledger); counts(ledger); locations(ledger); sources(hashes);
console.log(JSON.stringify({ tokens: ledger.tokens, files: ledger.files, classes: ledger.classes.length, sourceFiles: Object.keys(hashes).length, sourceChanges: 0 }));
function mutant(name, check, value, mutate) {
    const changed = structuredClone(value);
    mutate(changed);
    assert.throws(() => check(changed), assert.AssertionError, 'mutant must fail: ' + name);
    console.log('Caught mutant: ' + name);
}
mutant('missing token', coverage, ledger, d => d.sites.pop());
mutant('duplicate class membership', membership, ledger, d => d.classes[0].siteIds.push(d.classes[0].siteIds[0]));
mutant('wrong class count', counts, ledger, d => d.classes[0].count++);
mutant('wrong token location', locations, ledger, d => d.sites[0].column++);
const scratch = fs.mkdtempSync(path.join(require('node:os').tmpdir(), 'adamic-classification-'));
const file = Object.keys(hashes)[0];
const input = path.join(scratch, 'source-mutant.txt');
try {
    fs.copyFileSync(path.join(tree, file), input);
    fs.appendFileSync(input, '\n// classification source preservation mutant\n');
    mutant('changed compiler source input', sources, hashes, d => {
        d[file] = crypto.createHash('sha256').update(fs.readFileSync(input)).digest('hex');
    });
} finally {
    fs.unlinkSync(input);
    fs.rmdirSync(scratch);
}
