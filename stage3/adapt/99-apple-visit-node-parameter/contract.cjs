'use strict';
// The owner contract, run on an adapted tree: node contract.cjs TREE. It rebuilds visitorPublic.ts with the pinned
// before owner in a scratch copy, then checks the edit under LF and CRLF, a zero-edit replay that keeps bytes and
// mtime, and that a renamed, duplicated or drifted owner (four drifts, each one token), or a third overload, is refused
// with the file untouched.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const {adapt, locate, owners} = require('./adapt.cjs');
const pin = owners[0];
assert.equal(process.argv.length, 3, 'usage: node contract.cjs TREE');
const raw = fs.readFileSync(path.join(process.argv[2], pin.file), 'utf8');
const original = raw.replace(/\r\n/g, '\n');
const state = locate(original);
assert(state.adapted, 'contract: the tree must hold the adapted owner');
const before = original.slice(0, state.start) + pin.before + original.slice(state.end);
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'visit-node-owner-'));
const file = path.join(scratch, pin.file);
fs.mkdirSync(path.dirname(file), {recursive: true});
for (const newline of ['\n', '\r\n']) {
    const input = before.replace(/\n/g, newline);
    fs.writeFileSync(file, input);
    assert.equal(adapt(scratch).edits, 1);
    const expected = original.replace(/\n/g, newline);
    assert.equal(fs.readFileSync(file, 'utf8'), expected);
    const stamp = fs.statSync(file, {bigint: true}).mtimeNs;
    assert.equal(adapt(scratch).edits, 0);
    assert.equal(fs.readFileSync(file, 'utf8'), expected);
    assert.equal(fs.statSync(file, {bigint: true}).mtimeNs, stamp);
}
const overload = before.slice(0, state.start).match(/export function visitNode<[^]*?;\n/);
assert(overload, 'contract: no overload of visitNode to duplicate');
for (const mutant of [
    before.replace('export function visitNode(\n    node: Node,', 'export function visitNodeRenamed(\n    node: Node,'),
    before + '\n' + pin.before + '\n',
    before.replace(pin.before, pin.before.replace('if (node === undefined) {', 'if (node == null) {')),
    before.replace(pin.before, pin.before.replace('return visitedNode;', 'return visited as Node;')),
    before.replace(pin.before, pin.before.replace('    node: Node,\n', '    node: Node | null,\n')),
    before.replace(pin.before, pin.before.replace('): Node | undefined {', '): Node {')),
    before.replace(pin.before, overload[0] + pin.before),
]) {
    assert.notEqual(mutant, before, 'a contract mutant changed nothing');
    fs.writeFileSync(file, mutant);
    assert.throws(() => adapt(scratch));
    assert.equal(fs.readFileSync(file, 'utf8'), mutant);
}
fs.writeFileSync(file, before + '\n// unrelated owner stays\n');
adapt(scratch);
assert(fs.readFileSync(file, 'utf8').endsWith('// unrelated owner stays\n'));
console.log('whole-owner drift, renamed, duplicate, third overload, LF/CRLF, unrelated bytes and zero-edit replay: pass');
