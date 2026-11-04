'use strict';
const fs = require('node:fs'), path = require('node:path'), os = require('node:os');
const assert = require('node:assert/strict'), crypto = require('node:crypto');
const {adapt, targets} = require('../adapt.cjs');
assert.equal(process.argv.length, 3, 'usage: node owner-contract.cjs PRE_91_TREE');
const before = path.resolve(process.argv[2]), root = fs.mkdtempSync(path.join(os.tmpdir(), 'adapt91-owner-contract-'));
function tree(name) {
    const dest = path.join(root, name); fs.mkdirSync(dest);
    fs.cpSync(path.join(before, 'src'), path.join(dest, 'src'), {recursive: true});
    return dest;
}
function state(dest) {
    return targets.map(t => {
        const p = path.join(dest, t.file);
        return [t.file, crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'), String(fs.statSync(p, {bigint: true}).mtimeNs)];
    });
}
const first = tree('first'), applied = adapt(first);
assert.equal(applied.files, 2); const snapshot = state(first);
assert.equal(adapt(first).files, 0); assert.deepEqual(state(first), snapshot);
const unrelated = tree('unrelated-utility-edit'), utility = path.join(unrelated, 'src/compiler/utilities.ts');
const addition = '\r\nfunction unrelatedAdapt91OwnerFixture() { return 91; }\r\n';
fs.appendFileSync(utility, addition);
assert.equal(adapt(unrelated).files, 2); assert(fs.readFileSync(utility, 'utf8').endsWith(addition));
const unrelatedState = state(unrelated); assert.equal(adapt(unrelated).files, 0); assert.deepEqual(state(unrelated), unrelatedState);
const declaration = tree('adapt32-declaration'), p = path.join(declaration, 'src/compiler/utilities.ts');
const old = 'host: { directoryExists?: (directoryName: string) => boolean; }';
const next = 'host: { directoryExists?: ((directoryName: string) => boolean) | undefined; }';
const raw = fs.readFileSync(p, 'utf8');
const [from, to] = raw.includes(old) ? [old, next] : [next, old];
assert.equal(raw.split(from).length, 2); fs.writeFileSync(p, raw.replace(from, to));
assert.equal(adapt(declaration).files, 2); assert(fs.readFileSync(p, 'utf8').includes(to));
const refused = [];
for (const target of targets) {
    const dest = tree('changed-' + target.owner), file = path.join(dest, target.file), text = fs.readFileSync(file, 'utf8');
    const needle = `forEach(${target.registry}, fn => fn(${target.argument}));`;
    assert.equal(text.split(needle).length, 2);
    fs.writeFileSync(file, text.replace(needle, '/* changed owner */ ' + needle));
    const saved = state(dest); let writes = 0; const write = fs.writeFileSync;
    fs.writeFileSync = function(...args) { writes++; return write.apply(this, args); };
    try { assert.throws(() => adapt(dest), /target owner moved/); } finally { fs.writeFileSync = write; }
    assert.equal(writes, 0); assert.deepEqual(state(dest), saved); refused.push(target.owner);
}
console.log(JSON.stringify({first: applied, noWriteSecondRun: true, unrelatedUtilityEditAccepted: true, directoryHelperParameterEditAccepted: true, changedOwnersRefusedBeforeWrites: refused}, null, 2));
