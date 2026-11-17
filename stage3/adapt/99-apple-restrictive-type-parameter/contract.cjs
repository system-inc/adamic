'use strict';
// The owner contract, run on an adapted tree: node contract.cjs TREE. It rebuilds checker.ts with the pinned before
// owner in a scratch copy, then checks the edit under LF and CRLF, a zero-edit replay that keeps bytes and mtime, and
// that every drifted, missing or duplicated owner is refused with the file untouched.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const {adapt, locate, owners} = require('./adapt.cjs');
const raw = fs.readFileSync(path.join(process.argv[2], owners[0].file), 'utf8');
const original = raw.replace(/\r\n/g, '\n');
const state = locate(original), before = original.slice(0, state.start) + owners[0].before + original.slice(state.end);
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'restrictive-owner-'));
const file = path.join(scratch, owners[0].file);
fs.mkdirSync(path.dirname(file), {recursive: true});
for (const newline of ['\n', '\r\n']) {
    const input = before.replace(/\n/g, newline);
    fs.writeFileSync(file, input);
    assert.equal(adapt(scratch).edits, 1);
    const expected = (before.slice(0, state.start) + owners[0].after + before.slice(state.start + owners[0].before.length)).replace(/\n/g, newline);
    assert.equal(fs.readFileSync(file, 'utf8'), expected);
    const stamp = fs.statSync(file, {bigint: true}).mtimeNs;
    assert.equal(adapt(scratch).edits, 0);
    assert.equal(fs.readFileSync(file, 'utf8'), expected);
    assert.equal(fs.statSync(file, {bigint: true}).mtimeNs, stamp);
}
for (const mutant of [
    before.replace('function getRestrictiveTypeParameter', 'function changedOwner'),
    before + '\n' + owners[0].before,
    before.replace('(tp.restrictiveInstantiation as TypeParameter).constraint = noConstraintType', '(tp.restrictiveInstantiation as TypeParameter).constraint = unknownType'),
    before.replace('tp.restrictiveInstantiation = createTypeParameter(tp.symbol),', 'tp.restrictiveInstantiation = cloneTypeParameter(tp),'),
    before.replace('tp.constraint === noConstraintType ? tp :', 'tp.constraint === circularConstraintType ? tp :'),
]) {
    assert.notEqual(mutant, before, 'a contract mutant changed nothing');
    fs.writeFileSync(file, mutant);
    assert.throws(() => adapt(scratch));
    assert.equal(fs.readFileSync(file, 'utf8'), mutant);
}
fs.writeFileSync(file, before + '\n// unrelated owner stays\n');
adapt(scratch);
assert(fs.readFileSync(file, 'utf8').endsWith('// unrelated owner stays\n'));
console.log('whole-owner drift, duplicate, missing, half-adapted, LF/CRLF, unrelated bytes and zero-edit replay: pass');
