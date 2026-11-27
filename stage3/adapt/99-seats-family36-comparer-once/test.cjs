'use strict';

const assert = require('node:assert/strict');
const {execFileSync} = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const ts = require('typescript');
const {ownerOf, switchSpan, adaptedSpan, unchanged} = require('./adapt.cjs');

assert.equal(process.argv.length, 3, 'usage: node test.cjs <applied tree>');
assert.equal(ts.version, '6.0.3', 'expected locked TypeScript 6.0.3');

const tree = path.resolve(process.argv[2]);
const ref = 'refs/stage3/99-seats-family36-comparer-once';
const relative = 'src/compiler/core.ts';
const snapshot = side => execFileSync('git', ['-C', tree, 'show', ref + '-' + side + ':' + relative]);
const before = snapshot('before');
const after = snapshot('after');

const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'adamic-comparer-'));
try {
    const file = path.join(scratch, relative);
    fs.mkdirSync(path.dirname(file), {recursive: true});
    fs.writeFileSync(file, before);
    const run = () => JSON.parse(execFileSync(process.execPath, [
        path.join(__dirname, 'adapt.cjs'), scratch,
    ], {encoding: 'utf8'}));
    assert.deepEqual(run(), {files: 1, edits: 1});
    assert.deepEqual(fs.readFileSync(file), after, 'adapter output differs from the after ref');
    const stamp = new Date('2000-01-01T00:00:00Z');
    fs.utimesSync(file, stamp, stamp);
    assert.deepEqual(run(), {files: 0, edits: 0});
    assert.equal(fs.statSync(file).mtimeMs, stamp.getTime(), 'replay wrote to core.ts');
    assert.deepEqual(fs.readFileSync(file), after, 'replay changed core.ts');
    console.log('PASS output equals after ref and replay writes nothing');
} finally {
    fs.rmSync(scratch, {recursive: true, force: true});
}

const originalOwner = ownerOf(before.toString());
const adaptedOwner = ownerOf(after.toString());
assert.deepEqual(unchanged(after.toString(), adaptedSpan(adaptedOwner)),
    unchanged(before.toString(), switchSpan(originalOwner)), 'only the loop switch may change');
for (const owner of [originalOwner, adaptedOwner]) {
    assert.equal(owner.body.statements[1].getText(), 'let last = array[0]!;');
}
console.log('PASS cut preserves all text outside the switch, including last');

const Comparison = {LessThan: -1, EqualTo: 0, GreaterThan: 1};
const Debug = {fail(message) { throw new Error(message); }};
const emptyArray = [];

function evaluate(owner) {
    const {outputText} = ts.transpileModule(owner.getText(), {
        compilerOptions: {target: ts.ScriptTarget.ES2024},
    });
    const load = new Function('Comparison', 'Debug', 'emptyArray',
        outputText + '\nreturn deduplicateSorted;');
    return load(Comparison, Debug, emptyArray);
}

const original = evaluate(originalOwner);
const adapted = evaluate(adaptedOwner);

function drive(fn, array, comparer) {
    const calls = [];
    let result;
    let message;
    try {
        result = fn(array, (next, last) => {
            calls.push([next, last]);
            return comparer(next, last, calls.length);
        });
    } catch (error) {
        message = error.message;
    }
    return {result, calls, message};
}

function sameArray(actual, expected, label) {
    if (expected === undefined) {
        assert.equal(actual, undefined, label);
        return;
    }
    assert(Array.isArray(actual), label);
    assert.equal(actual.length, expected.length, label + ' length');
    expected.forEach((value, index) => assert.equal(actual[index], value, label + ' element ' + index));
}

function sameCalls(actual, expected, label) {
    assert.equal(actual.length, expected.length, label + ' call count');
    expected.forEach(([next, last], index) => {
        assert.equal(actual[index][0], next, label + ' next argument ' + index);
        assert.equal(actual[index][1], last, label + ' last argument ' + index);
    });
}

function check(name, array, comparer, result, calls, message) {
    const beforeRun = drive(original, array, comparer);
    const afterRun = drive(adapted, array, comparer);
    sameArray(beforeRun.result, result, name + ' original array');
    sameCalls(beforeRun.calls, calls, name + ' original');
    assert.equal(beforeRun.message, message, name + ' original throw');
    sameArray(afterRun.result, beforeRun.result, name + ' adapted array');
    sameCalls(afterRun.calls, beforeRun.calls, name + ' adapted');
    assert.equal(afterRun.message, beforeRun.message, name + ' adapted throw');
    console.log('PASS ' + name + ' (' + calls.length + ' calls)');
}

const a = {key: 1};
const b = {key: 1};
const c = {key: 2};
const d = {key: 2};
check('false pushes and numeric zero continues', [a, b, c],
    next => next === b ? false : 0, [a, b], [[b, a], [c, b]]);
check('equality duplicate runs', [a, b, c, d], (next, last) => next.key === last.key,
    [a, c], [[b, a], [c, a], [d, c]]);
check('relational duplicate runs', [a, b, c, d], (next, last) => Math.sign(next.key - last.key),
    [a, c], [[b, a], [c, a], [d, c]]);
check('relational unsorted', [c, a], (next, last) => Math.sign(next.key - last.key),
    undefined, [[a, c]], 'Array is unsorted.');

const witnessComparer = (next, last, calls) => calls === 1 ? true : Math.sign(next - last);
check('round 1 sorted witness', [1, 1, 2, 2, 4], witnessComparer, [1, 2, 4],
    [[1, 1], [2, 1], [2, 2], [4, 2]]);
check('round 1 unsorted witness after four sorted calls', [4, 2, 1],
    (next, last, calls) => witnessComparer(next, last, calls + 4),
    undefined, [[2, 4]], 'Array is unsorted.');
check('empty', [], () => assert.fail('unexpected call'), [], []);
check('singleton', [a], () => assert.fail('unexpected call'), [a], []);

for (const [name, value, action] of [
    ['true', true, 'continue'], ['false', false, 'push'], ['0', 0, 'continue'],
    ['-0', -0, 'continue'], ['-1', -1, 'throw'], ['1', 1, 'push'],
    ['NaN', NaN, 'push'], ['undefined', undefined, 'push'], ['null', null, 'push'],
    ['string "0"', '0', 'push'],
]) {
    const result = action === 'throw' ? undefined : action === 'continue' ? [a] : [a, b];
    check('result ' + name, [a, b], () => value, result, [[b, a]],
        action === 'throw' ? 'Array is unsorted.' : undefined);
}
