'use strict';
const fs = require('node:fs'), path = require('node:path'), os = require('node:os'), assert = require('node:assert/strict');
const adapter = require('./adapt.cjs'), pin = adapter.owners[0];
assert.equal(process.argv.length, 3, 'usage: contract.cjs BEFORE_TREE');
const raw = fs.readFileSync(path.join(process.argv[2], 'src/compiler/path.ts'), 'utf8');
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'path-components-contract-'));
function tree(name, text = raw) {
    const root = path.join(scratch, name); fs.mkdirSync(path.join(root, 'src/compiler'), {recursive: true});
    fs.writeFileSync(path.join(root, 'src/compiler/path.ts'), text); return root;
}
function snapshot(root) {
    const name = path.join(root, 'src/compiler/path.ts');
    return [adapter.digest(fs.readFileSync(name)), String(fs.statSync(name, {bigint: true}).mtimeNs)];
}
const first = tree('first'), initial = adapter.adapt(first), after = snapshot(first);
const write = fs.writeFileSync; let writes = 0;
function observe(fn) {
    writes = 0; fs.writeFileSync = function (...args) { writes++; return write.apply(this, args); };
    try { return fn(); } finally { fs.writeFileSync = write; }
}
const replay = observe(() => adapter.adapt(first));
assert.equal(writes, 0); assert.equal(replay.edits, 0); assert.deepEqual(snapshot(first), after);
const refusals = [];
function refused(name, text) {
    const root = tree(name, text), before = snapshot(root);
    observe(() => assert.throws(() => adapter.adapt(root)));
    assert.equal(writes, 0); assert.deepEqual(snapshot(root), before); refusals.push(name);
}
const before = raw.replace(/\r\n/g, '\n');
refused('changed-owner', before.replace('return "" as T;', 'return ("" as T);'));
refused('half-empty-cast', before.replace('return "" as T;', 'return "";'));
refused('half-final-cast', before.replace('.join(directorySeparator) as T;', '.join(directorySeparator);'));
refused('missing-owner', before.replace(pin.before, ''));
refused('moved-owner', before.replace(pin.before, 'namespace moved {\n' + pin.before + '\n}'));
refused('duplicate-owner', before + '\n' + pin.before);
refused('parse-error', before + '\nfunction {');
const adapted = before.replace(pin.before, pin.after);
refused('changed-after-owner', adapted.replace('if (pathComponents.length === 0)', 'if (pathComponents.length < 1)'));
refused('missing-after-overload', adapted.replace('/** @internal */ export function getPathFromPathComponents(pathComponents: readonly Path[], length?: number): Path;\n', ''));
refused('moved-section-comment', adapted.replace('//// Path Formatting\n\n', '').replace('/** @internal */ export function getPathFromPathComponents', '//// Path Formatting\n\n/** @internal */ export function getPathFromPathComponents'));
for (const [name, text] of [['lf', before], ['crlf', before.replace(/\n/g, '\r\n')]]) {
    const root = tree(name, text); adapter.adapt(root);
    const result = fs.readFileSync(path.join(root, 'src/compiler/path.ts'), 'utf8');
    assert.equal(result, text.replace(name === 'crlf' ? pin.before.replace(/\n/g, '\r\n') : pin.before, name === 'crlf' ? pin.after.replace(/\n/g, '\r\n') : pin.after));
}
const addition = '\nfunction unrelatedPathComponentsContract() { return 99; }\n';
const unrelated = tree('unrelated', before + addition); adapter.adapt(unrelated);
assert(fs.readFileSync(path.join(unrelated, 'src/compiler/path.ts'), 'utf8').endsWith(addition));
console.log(JSON.stringify({initial, replay, replayWrites: 0, replayBytesAndMtimeUnchanged: true, refusals, unrelatedPreserved: true, lineEndingsPreserved: true}, null, 2));
