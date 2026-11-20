'use strict';
// node witness.cjs TREE: the adapted tree holds makeReverseMap's dense owner, and the dense build reads, at every
// index, what the sparse build read, over tsc's own two maps.
// Stock TypeScript 6.0.3's tokenToString and regularExpressionFlagToCharacterCode are upstream's sparse builds read
// by index, so reading each at every index from 0 to 1023 is upstream's array, and the map each was built from is
// recovered as the entries that read back (a name written twice reads its last write, which both builds keep, since
// both walk the map in one order). The pinned before and after owners are transpiled and run on that map. For each,
// the lengths are equal, every index from 0 to the length plus 16 reads the same value by Object.is, the after
// build holds every index below its length (no hole), and the after build read at each index is stock's answer.
// The dense body runs guarded as rulings.md:318 guards it, so a write it makes where its array holds no index fails
// here, as it would under Adamic, instead of growing the array as Node does; upstream's body, guarded the same, must
// stop, which shows the guard sees such a write. WITNESS_MUTANT=off-by-one sizes the dense array to the largest value
// instead of one past it, and must fail.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const {owner, pins} = require('./adapt.cjs');
assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: node witness.cjs TREE');
let failed = 0;
const verdict = (ok, line) => { console.log(`witness: ${ok ? 'PASS' : 'FAIL'} ${line}`); if (!ok) failed++; };

const tree = path.resolve(process.argv[2]);
verdict(owner(fs.readFileSync(path.join(tree, pins.file), 'utf8')).state === 'after', `${pins.file}: the owner is the pinned dense build`);

let after = pins.after;
if (process.env.WITNESS_MUTANT === 'off-by-one') {
    assert(after.includes('length = value + 1;'));
    after = after.replace('length = value + 1;', 'length = value;');
    console.log('witness: mutant off-by-one planted');
}
const compile = text => new Function(`${ts.transpileModule(text, {compilerOptions: {target: ts.ScriptTarget.ES2024}}).outputText}\nreturn makeReverseMap;`)();
// guarded holds a body to rulings.md:318 as Adamic runs it: a write lands only at an index the array holds, and
// anything else throws, where Node grows the array. The dense body runs guarded, so a sizing that falls short of a
// value can't be covered by Node growing the array; upstream's body runs as Node runs it.
const guarded = text => {
    assert.equal(text.split('result[value] = name;').length - 1, 1, 'witness: the owner writes result[value] once');
    return text.replace('result[value] = name;',
        'if (!(Number.isInteger(value) && value >= 0 && value < result.length)) throw new RangeError(`index ${value} is outside an array of length ${result.length}`);\r\n        result[value] = name;');
};
const sparse = compile(pins.before), dense = compile(guarded(after)), sparseGuarded = compile(guarded(pins.before));

const readers = {
    tokenToString: k => ts.tokenToString(k),
    regularExpressionFlagToCharacterCode: f => ts.regularExpressionFlagToCharacterCode(f),
};
for (const [name, read] of Object.entries(readers)) {
    const stock = Array.from({length: 1024}, (_, index) => read(index));
    const map = new Map();
    stock.forEach((value, index) => { if (value !== undefined) map.set(value, index); });
    let stopped;
    try { sparseGuarded(map); } catch (error) { stopped = error.message; }
    verdict(stopped !== undefined, `${name}: upstream's body, guarded as :318 guards it, stops (${stopped})`);
    const before = sparse(map);
    let built;
    try { built = dense(map); } catch (error) {
        verdict(false, `${name}: the dense body wrote where its array holds no index: ${error.message}`);
        continue;
    }
    const holes = Array.from({length: before.length}, (_, index) => index).filter(index => !(index in before)).length;
    let differing = [];
    for (let index = 0; index < Math.max(before.length, built.length) + 16; index++)
        if (!Object.is(before[index], built[index])) differing.push(index);
    const stockDiffering = stock.map((value, index) => index).filter(index => !Object.is(built[index], stock[index]));
    const denseHoles = Array.from({length: built.length}, (_, index) => index).filter(index => !(index in built));
    verdict(before.length === built.length && differing.length === 0 && stockDiffering.length === 0 && denseHoles.length === 0,
        `${name}: ${map.size} entries, sparse length ${before.length} with ${holes} holes, dense length ${built.length} with ${denseHoles.length} holes; ` +
        `indexes reading differently ${JSON.stringify(differing.slice(0, 8))}, against stock ${JSON.stringify(stockDiffering.slice(0, 8))}`);
}
const empty = dense(new Map());
verdict(Array.isArray(empty) && empty.length === 0, 'an empty map builds an empty array, as upstream does');
console.log(failed === 0 ? 'witness: PASS' : `witness: FAIL (${failed})`);
process.exit(failed === 0 ? 0 : 1);
