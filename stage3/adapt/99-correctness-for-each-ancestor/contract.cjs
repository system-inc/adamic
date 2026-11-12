'use strict';
const fs = require('node:fs'), path = require('node:path'), os = require('node:os'), assert = require('node:assert/strict');
const adapter = require('./adapt.cjs'), pin = adapter.owners[0];
assert.equal(process.argv.length, 3, 'usage: contract.cjs BEFORE_TREE');
const raw = fs.readFileSync(path.join(process.argv[2], 'src/compiler/path.ts'), 'utf8');
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'ancestor-contract-'));
function tree(name, text = raw) {
    const root = path.join(scratch, name); fs.mkdirSync(path.join(root, 'src/compiler'), {recursive: true});
    fs.writeFileSync(path.join(root, 'src/compiler/path.ts'), text); return root;
}
function snapshot(root) {
    const name = path.join(root, 'src/compiler/path.ts');
    return [adapter.digest(fs.readFileSync(name)), String(fs.statSync(name, {bigint: true}).mtimeNs)];
}
const first = tree('first'), initial = adapter.adapt(first), after = snapshot(first);
assert.deepEqual(initial, {files: 1, edits: 1, javascriptChanging: false});
const write = fs.writeFileSync; let writes = 0;
function observe(fn) {
    writes = 0; fs.writeFileSync = function (...args) { writes++; return write.apply(this, args); };
    try { return fn(); } finally { fs.writeFileSync = write; }
}
const replay = observe(() => adapter.adapt(first));
assert.equal(writes, 0); assert.deepEqual(replay, {files: 0, edits: 0, javascriptChanging: false}); assert.deepEqual(snapshot(first), after);
assert.equal(adapter.digest(pin.before), pin.ownerBeforeSha256); assert.equal(adapter.digest(pin.after), pin.ownerAfterSha256);
const refusals = [];
function refused(name, text) {
    const root = tree(name, text), before = snapshot(root);
    observe(() => assert.throws(() => adapter.adapt(root)));
    assert.equal(writes, 0); assert.deepEqual(snapshot(root), before); refusals.push(name);
}
const before = raw.replace(/\r\n/g, '\n');
const pathOverload = 'export function forEachAncestorDirectory<T>(directory: Path, callback: (directory: Path) => T | undefined): T | undefined;\n';
refused('changed-body', before.replace('const parentPath = getDirectoryPath(directory) as P;', 'const parentPath = (getDirectoryPath(directory)) as P;'));
refused('half-cast', before.replace('getDirectoryPath(directory) as P', 'getDirectoryPath(directory)'));
refused('half-overload', before.replace(pathOverload, ''));
refused('half-implementation', before.replace('<T, P extends string>(directory: P, callback: (directory: P)', '<T>(directory: string, callback: (directory: string)'));
refused('missing-owner', before.replace(pin.before, ''));
refused('moved-owner', before.replace(pin.before, 'namespace moved {\n' + pin.before + '\n}'));
refused('duplicate-before-owner', before + '\n' + pin.before);
refused('parse-error', before + '\nfunction {');
const adapted = before.replace(pin.before, pin.after);
refused('changed-after-body', adapted.replace('const parentPath = getDirectoryPath(directory);', 'const parentPath = (getDirectoryPath(directory));'));
refused('duplicate-after-owner', adapted + '\n' + pin.after);
refused('restored-path-overload', adapted.replace(pin.after, pathOverload + pin.after));
refused('missing-string-overload', adapted.replace('export function forEachAncestorDirectory<T>(directory: string, callback: (directory: string) => T | undefined): T | undefined;\n', ''));
refused('changed-comment', adapted.replace('returning the first defined result.', 'returning a defined result.'));
refused('missing-section-comment', adapted.replace('//// Path Traversal\n\n', ''));
for (const [name, text] of [['lf', before], ['crlf', before.replace(/\n/g, '\r\n')]]) {
    const root = tree(name, text); adapter.adapt(root);
    const result = fs.readFileSync(path.join(root, 'src/compiler/path.ts'), 'utf8');
    assert.equal(result, text.replace(name === 'crlf' ? pin.before.replace(/\n/g, '\r\n') : pin.before, name === 'crlf' ? pin.after.replace(/\n/g, '\r\n') : pin.after));
}
const addition = '\nfunction unrelatedAncestorContract() { return 99; }\n';
const unrelated = tree('unrelated', before + addition); adapter.adapt(unrelated);
assert(fs.readFileSync(path.join(unrelated, 'src/compiler/path.ts'), 'utf8').endsWith(addition));
console.log(JSON.stringify({initial, replay, replayWrites: 0, replayBytesAndMtimeUnchanged: true, refusals, unrelatedPreserved: true, lineEndingsPreserved: true}, null, 2));
fs.rmSync(scratch, {recursive: true, force: true});
