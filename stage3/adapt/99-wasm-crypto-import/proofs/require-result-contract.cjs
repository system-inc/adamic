'use strict';
const fs = require('node:fs'), path = require('node:path'), os = require('node:os'), cp = require('node:child_process'), assert = require('node:assert/strict');
const {adapt, requireBefore, requireAfter} = require('../adapt.cjs');
const {restore} = require('../require-result-mutant.cjs');
const baseline = path.resolve(process.argv[2]);
const before = cp.execFileSync('git', ['-C', path.resolve(process.argv[3] || baseline), 'show', 'refs/stage3/99-wasm-crypto-import-before:src/compiler/sys.ts'], {encoding: 'utf8'});
const expected = fs.readFileSync(path.join(baseline, 'src/compiler/sys.ts'), 'utf8');
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'require-result-contract-')), file = path.join(scratch, 'src/compiler/sys.ts');
const cases = [];
function seed(text = before) { fs.writeFileSync(file, text); }
function snapshot() { return [fs.readFileSync(file, 'utf8'), String(fs.statSync(file, {bigint: true}).mtimeNs)]; }
function noWrites(action) {
    const write = fs.writeFileSync; let writes = 0;
    fs.writeFileSync = (...args) => { writes++; return write(...args); };
    try { action(); assert.equal(writes, 0); } finally { fs.writeFileSync = write; }
}
function refused(name, text) {
    seed(text); const old = snapshot();
    noWrites(() => assert.throws(() => adapt(scratch), undefined, name));
    assert.deepEqual(snapshot(), old); cases.push(name);
}
try {
    fs.cpSync(path.join(baseline, 'src'), path.join(scratch, 'src'), {recursive: true});
    for (const newline of ['\n', '\r\n']) {
        seed(before.replace(/\r\n/g, '\n').replace(/\n/g, newline));
        assert.equal(adapt(scratch).files, 1);
        assert.equal(fs.readFileSync(file, 'utf8'), expected.replace(/\r\n/g, '\n').replace(/\n/g, newline));
        const old = snapshot(); noWrites(() => assert.equal(adapt(scratch).files, 0)); assert.deepEqual(snapshot(), old);
        restore(scratch);
        assert(fs.readFileSync(file, 'utf8').includes(requireBefore));
        refused('restored-result-' + JSON.stringify(newline), fs.readFileSync(file, 'utf8'));
    }
    refused('half-adapted-result', before.replace(requireBefore, requireAfter));
    refused('owner-drift', before.replace(requireBefore, requireBefore + ' /* drift */'));
    refused('duplicate-arrow', before.replace(requireBefore, requireBefore + '\n' + requireBefore));
    const owner = require('./targets.json')[0].ownerBeforeText;
    refused('duplicate-owner', before + '\r\n' + owner);
    refused('mixed-line-endings', before.replace('\r\n', '\n'));
    console.log(JSON.stringify({LF: true, CRLF: true, exactAfterText: true, zeroWriteReplay: true, bytesAndMtimePreserved: true, restoringMutant: true, refused: cases}, null, 2));
} finally { fs.rmSync(scratch, {recursive: true, force: true}); }
