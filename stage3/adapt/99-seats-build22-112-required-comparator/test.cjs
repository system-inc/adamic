'use strict';

const assert = require('node:assert/strict');
const { execFileSync, spawnSync } = require('node:child_process');
const { createHash } = require('node:crypto');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('typescript');

assert.equal(process.argv.length, 3, 'usage: node test.cjs <full stage3 applied tree>');
assert.equal(ts.version, '6.0.3', 'pinned TypeScript parser');
const applied = path.resolve(process.argv[2]);
const ref = 'refs/stage3/99-seats-build22-112-required-comparator';
const utilities = 'src/compiler/utilities.ts';
const core = 'src/compiler/core.ts';
const namespace = 'src/compiler/_namespaces/ts.ts';
function snapshot(side, file) {
    return execFileSync('git', ['-C', applied, 'show', ref + '-' + side + ':' + file], { maxBuffer: 8 * 1024 * 1024 });
}

function hash(bytes) {
    return createHash('sha256').update(bytes).digest('hex');
}

function replaceOnce(text, before, after) {
    assert.equal(text.split(before).length, 2, 'expected one occurrence of ' + before);
    return text.replace(before, after);
}

function checkHelper(file, bytes, coreBytes) {
    const source = ts.createSourceFile(file, bytes.toString('utf8'), ts.ScriptTarget.Latest, true);
    const helpers = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === 'toSortedStrings');
    assert.equal(helpers.length, 1, 'generated helper declaration');
    const js = ts.transpileModule(helpers[0].getText(source), { reportDiagnostics: true });
    assert.deepEqual(js.diagnostics, [], 'generated helper transpiles without diagnostics');
    const emptyArray = [];
    const context = vm.createContext({ emptyArray, exports: {} });
    const sort = vm.runInContext(js.outputText + '\ntoSortedStrings', context);
    const coreSource = ts.createSourceFile(core, coreBytes.toString('utf8'), ts.ScriptTarget.Latest, true);
    const originals = coreSource.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === 'toSorted');
    assert.equal(originals.length, 1, 'original toSorted declaration');
    const originalJs = ts.transpileModule(originals[0].getText(coreSource), {
        compilerOptions: { module: ts.ModuleKind.CommonJS },
        reportDiagnostics: true,
    });
    assert.deepEqual(originalJs.diagnostics, [], 'original toSorted transpiles without diagnostics');
    const originalSort = vm.runInContext(originalJs.outputText + '\ntoSorted', context);
    const input = ['b1', 'a2', 'a1', 'b2'];
    let calls = 0;
    const comparer = (a, b) => {
        calls++;
        return a.charCodeAt(0) - b.charCodeAt(0);
    };
    const result = sort(input, comparer);
    assert.deepEqual(Array.from(result), ['a2', 'a1', 'b1', 'b2'], 'stable nonempty order with tied keys');
    assert.notStrictEqual(result, input, 'nonempty return is a fresh array');
    assert.notStrictEqual(result, emptyArray, 'nonempty return is distinct from the sentinel');
    assert.deepEqual(input, ['b1', 'a2', 'a1', 'b2'], 'input stays unsorted');
    assert(calls > 0, 'nonempty branch calls the comparator');
    const originalResult = originalSort(input, comparer);
    assert.deepEqual(Array.from(originalResult), ['a2', 'a1', 'b1', 'b2'], 'original stable nonempty order with tied keys');
    assert.deepEqual(Array.from(result), Array.from(originalResult), 'helper and original nonempty order agree');
    assert.notStrictEqual(originalResult, input, 'original nonempty return is a fresh array');
    assert.notStrictEqual(originalResult, emptyArray, 'original nonempty return is distinct from the sentinel');
    assert.notStrictEqual(result, originalResult, 'nonempty calls allocate separate arrays');
    assert.deepEqual(input, ['b1', 'a2', 'a1', 'b2'], 'input stays unsorted after original sort');
    console.log('stable nonempty order, fresh copy and untouched input passed for both sorts');
    const beforeEmpty = calls;
    const emptyInput = [];
    assert.strictEqual(sort(emptyInput, comparer), emptyArray, 'empty sort must return the shared emptyArray');
    assert.equal(calls, beforeEmpty, 'empty branch calls no comparator');
    assert.strictEqual(originalSort(emptyInput, comparer), emptyArray, 'original empty sort must return the shared emptyArray');
    assert.equal(calls, beforeEmpty, 'original empty branch calls no comparator');
    console.log('empty identity and comparator elision passed for both sorts');
}

const before = snapshot('before', utilities);
const expected = snapshot('after', utilities);
const original = before.toString('utf8');
const tree = fs.mkdtempSync(path.join(os.tmpdir(), 'adamic-required-comparator-'));
const file = path.join(tree, utilities);
const namespaceFile = path.join(tree, namespace);
const namespaceBefore = snapshot('before', namespace);

function applyCase(label, input, rejection, namespaceInput = namespaceBefore) {
    const bytes = Buffer.from(input);
    fs.writeFileSync(file, bytes);
    const namespaceBytes = Buffer.from(namespaceInput);
    fs.writeFileSync(namespaceFile, namespaceBytes);
    const run = spawnSync(process.execPath, [path.join(__dirname, 'adapt.cjs'), tree], { encoding: 'utf8' });
    if (run.error) throw run.error;
    const written = fs.readFileSync(file);
    const namespaceWritten = fs.readFileSync(namespaceFile);
    fs.writeFileSync(namespaceFile, namespaceBefore);
    console.log("%s namespace before %s after %s", label, hash(namespaceBytes), hash(namespaceWritten));
    assert(namespaceWritten.equals(namespaceBytes), label + ": namespace bytes changed");
    console.log('%s exit %s before %s after %s', label, run.status, hash(bytes), hash(written));
    assert.equal(run.signal, null, label + ': adapter must exit normally');
    if (rejection === undefined) {
        assert.equal(run.status, 0, label + ': ' + run.stderr);
    } else {
        assert(run.status > 0, label + ': expected rejection, adapter exit ' + run.status);
        assert.equal(hash(written), hash(bytes), label + ': rejected input SHA-256');
        assert(written.equals(bytes), label + ': rejected input bytes changed');
        assert(run.stderr.includes(rejection), label + ': expected ' + rejection + ' in stderr; got ' + run.stderr);
        console.log('%s rejected with %s', label, rejection);
    }
    return written;
}

const baseCall = 'toSorted<string>(files, compareStringsCaseSensitive)';
const headCall = 'toSortedStrings(files, compareStringsCaseSensitive)';
const baseLoop = 'for (const current of ' + baseCall + ') {';
const headLoop = 'for (const current of ' + headCall + ') {';
const shadow = 'const toSortedStrings = (items: readonly string[], _compare: (a: string, b: string) => number) => items;\r\n        ';
const third = 'for (const item of toSorted<string>(files.slice(), compareStringsCaseSensitive)) {}\r\n        ';
const comment = '// EOF drift\r\n';

try {
    fs.mkdirSync(path.dirname(namespaceFile), { recursive: true });
    const coreBytes = snapshot('before', core);
    fs.writeFileSync(path.join(tree, core), coreBytes);
    const output = applyCase('clean input', before);
    // Exercise the generated helper before the byte oracle, so a semantic drift
    // reports the behavior it breaks.
    checkHelper(file, output, coreBytes);
    assert(output.equals(expected), 'clean output matches the pipeline after snapshot byte for byte');
    assert(applyCase('clean reapplication', output).equals(output), 'clean reapplication is a byte no-op');

    const adapted = output.toString('utf8');
    const sortShadow = 'function toSorted<T>(items: readonly T[], _compare: (a: T, b: T) => number): readonly T[] { return items; }\r\n        ';
    const comparerShadow = 'const compareStringsCaseSensitive = (a: string, b: string) => (a < b ? 1 : a > b ? -1 : 0);\r\n        ';
    const namespaceText = namespaceBefore.toString('utf8');
    const anchor = 'export { performance };\r\n';
    const retarget = declaration => replaceOnce(namespaceText, anchor, anchor + '\r\n' + declaration);
    applyCase('namespace retarget of original sort', before, 'toSorted binding at visitDirectory files sort',
        retarget('export function toSorted<T>(array: readonly T[], _comparer?: import("../corePublic.js").Comparer<T>): import("../corePublic.js").SortedReadonlyArray<T> {\r\n' +
            '    return array as import("../corePublic.js").SortedReadonlyArray<T>;\r\n}\r\n'));
    applyCase('namespace retarget of comparator on adapted input', output, 'compareStringsCaseSensitive binding at visitDirectory files sort',
        retarget('export function compareStringsCaseSensitive(a: string | undefined, b: string | undefined): number { return 0; }\r\n'));
    applyCase('namespace without its core re-export', before, '_namespaces/ts.ts core value star export',
        replaceOnce(namespaceText, 'export * from "../core.js";\r\n', ''));
    applyCase('namespace retarget of emptyArray on adapted input', output, 'emptyArray binding at toSortedStrings empty branch',
        retarget('export const emptyArray: never[] = [] as never[];\r\n'));
    applyCase('local shadow of original sort', replaceOnce(original, baseLoop, sortShadow + baseLoop), 'toSorted binding at visitDirectory files sort');
    applyCase('local shadow of comparator on base input', replaceOnce(original, baseLoop, comparerShadow + baseLoop), 'compareStringsCaseSensitive binding at visitDirectory files sort');
    applyCase('local shadow of comparator on adapted input', replaceOnce(adapted, headLoop, comparerShadow + headLoop), 'compareStringsCaseSensitive binding at visitDirectory files sort');
    applyCase('local shadow of introduced helper', replaceOnce(original, baseLoop, shadow + baseLoop), 'unexpected toSortedStrings occurrence in utilities.ts');
    applyCase('local shadow of already-adapted helper', replaceOnce(adapted, headLoop, shadow + headLoop), 'unexpected toSortedStrings occurrence in utilities.ts');
    const escapedShadow = replaceOnce(shadow, 'toSortedStrings', 'toSorted\\u0053trings');
    applyCase('escaped local shadow', replaceOnce(original, baseLoop, escapedShadow + baseLoop), 'unexpected toSortedStrings identifier in utilities.ts');
    applyCase('helper name in a comment', replaceOnce(original, baseLoop, '// toSortedStrings\r\n        ' + baseLoop), 'unexpected toSortedStrings occurrence in utilities.ts');
    applyCase('duplicate same-argument loop', replaceOnce(original, baseLoop, 'for (const item of ' + baseCall + ') {}\r\n        ' + baseLoop), 'expected one visitDirectory files sort, found 2');
    applyCase('sort moved out of for-of', replaceOnce(original, baseLoop, 'const ordered = ' + baseCall + ';\r\n        for (const current of ordered) {'), 'expected one visitDirectory files sort, found 0');
    applyCase('drifted comparator argument', replaceOnce(original, baseCall, 'toSorted<string>(files, undefined)'), 'snapshot drift in files sort');
    applyCase('trailing comment on base', original + comment, 'snapshot drift at utilities.ts end of file');
    applyCase('partial adaptation', replaceOnce(original, baseCall, headCall), 'expected one toSortedStrings, found 0');
    const thirdOutput = applyCase('third loop with distinct argument', replaceOnce(original, baseLoop, third + baseLoop));
    assert(thirdOutput.equals(Buffer.from(replaceOnce(adapted, headLoop, third + headLoop))), 'third sort stays generic and unchanged');
    const commented = adapted + comment;
    assert(applyCase('trailing comment on already-adapted input', commented).equals(Buffer.from(commented)), 'adapted trailing comment stays unchanged');
    console.log('All guard inventory checks passed');
} finally {
    fs.rmSync(tree, { recursive: true, force: true });
}
