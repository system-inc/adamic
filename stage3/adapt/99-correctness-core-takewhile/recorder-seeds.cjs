'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const { record } = require('./count-preload.cjs');
function original(array, predicate) {
    if (array !== undefined) {
        const len = array.length;
        let index = 0;
        while (index < len && predicate(array[index])) index++;
        return array.slice(0, index);
    }
}
record([{}, {}], () => true, original);
record(new Array(1), () => true, original);
record([{}, {}], () => true, (array, predicate) => original(array, predicate).reverse());
record([1, 2], () => true, (array, predicate) => { const result=original(array,predicate); array[0]=3; return result; });
record([1, 2], () => true, (array, predicate) => { predicate(99); return original(array,predicate); });
const rows = fs.readFileSync(path.join(process.env.TW_CALL_DIR, 'seeds.'+process.pid+'.jsonl'),'utf8').trim().split('\n').map(JSON.parse);
assert(rows[0].sameResult && rows[0].sameSequence && !rows[0].acceptedRangeHasHole);
assert(rows[1].acceptedRangeHasHole && !rows[1].sameResult && rows[1].sameSequence);
assert(!rows[2].sameResult);
assert(!rows[3].stableInput);
assert(!rows[4].sameSequence);
console.log(JSON.stringify({ excludedFromLane: true, checks: ['dense reference identity','accepted hole','reversed result','mutated input','extra predicate call'], pass: true }, null, 2));
