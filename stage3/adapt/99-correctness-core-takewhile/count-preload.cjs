'use strict';
const fs = require('node:fs'), path = require('node:path'), vm = require('node:vm');
let sequence = 0, nextId = 0;
const ids = new WeakMap();
function encode(value) {
    if (value === undefined) return { type: 'undefined' };
    if (value === null) return null;
    if (typeof value === 'object' || typeof value === 'function') {
        if (!ids.has(value)) ids.set(value, ++nextId);
        return { type: typeof value, identity: ids.get(value), kind: value.kind };
    }
    if (typeof value === 'number' && !Number.isFinite(value)) return { type: 'number', value: String(value) };
    if (typeof value === 'bigint' || typeof value === 'symbol') return { type: typeof value, value: String(value) };
    return value;
}
function describe(array) {
    if (array === undefined) return { undefined: true };
    return { length: array.length, slots: Array.from({ length: array.length }, (_, i) => ({
        present: i in array, own: Object.hasOwn(array, i), value: encode(array[i])
    })) };
}
function candidate(array, predicate) {
    if (array !== undefined) {
        const result = [];
        const len = array.length;
        let index = 0;
        while (index < len) {
            const element = array[index];
            if (!predicate(element)) break;
            result.push(element);
            index++;
        }
        return result;
    }
}
function record(array, predicate, original, replay = candidate) {
    if (!process.env.TW_CALL_DIR) return original(array, predicate);
    // Copy slots without cloning elements. A replay uses the original predicate's
    // answers and checks its actual arguments, never calls the real callback twice.
    const input = describe(array), snapshot = array === undefined ? undefined : Array.prototype.slice.call(array);
    const calls = [], answers = [], holes = [];
    let index = 0;
    const before = original(array, element => {
        const present = index in array;
        calls.push(element);
        const answer = predicate(element);
        answers.push(answer);
        if (answer && !present) holes.push(index);
        index++;
        return answer;
    });
    const replayCalls = [];
    const after = replay(snapshot, element => {
        replayCalls.push(element);
        return answers[replayCalls.length - 1];
    });
    const sameSequence = calls.length === replayCalls.length && calls.every((value, i) => Object.is(value, replayCalls[i]));
    const sameResult = before === undefined || after === undefined ? before === after :
        Object.getPrototypeOf(before) === Object.getPrototypeOf(after) &&
        Reflect.ownKeys(before).length === Reflect.ownKeys(after).length &&
        Reflect.ownKeys(before).every((key, i) => key === Reflect.ownKeys(after)[i] && Object.is(before[key], after[key])) &&
        before.length === after.length && Array.from({ length: before.length }, (_, i) => i).every(i =>
            (i in before) === (i in after) && Object.hasOwn(before, i) === Object.hasOwn(after, i) && Object.is(before[i], after[i]));
    const stableInput = JSON.stringify(input) === JSON.stringify(describe(array));
    const row = { shard: process.env.TW_SHARD || 'whole', pid: process.pid, sequence: ++sequence,
        predicateName: predicate.name, input, acceptedRangeHoles: holes, acceptedRangeHasHole: holes.length !== 0,
        before: describe(before), after: describe(after),
        predicateArguments: calls.map(encode), replayPredicateArguments: replayCalls.map(encode), answers,
        sameResult, sameSequence, stableInput };
    fs.appendFileSync(path.join(process.env.TW_CALL_DIR, row.shard + '.' + process.pid + '.jsonl'), JSON.stringify(row) + '\n');
    return before;
}
globalThis.__takeWhileCensus = record;
function inject(context) {
    if (context && typeof context === 'object') Object.defineProperty(context, '__takeWhileCensus', { value: record, configurable: true });
    return context;
}
const create = vm.createContext;
vm.createContext = function(context = {}, ...rest) { return Reflect.apply(create, this, [inject(context), ...rest]); };
const run = vm.runInNewContext;
vm.runInNewContext = function(code, context = {}, ...rest) { return Reflect.apply(run, this, [code, inject(context), ...rest]); };
const script = vm.Script.prototype.runInNewContext;
vm.Script.prototype.runInNewContext = function(context = {}, ...rest) { return Reflect.apply(script, this, [inject(context), ...rest]); };
module.exports = { record, candidate };
