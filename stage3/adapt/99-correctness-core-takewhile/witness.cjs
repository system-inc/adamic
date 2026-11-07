#!/usr/bin/env node
"use strict";
const fs = require('node:fs');
const assert = require("node:assert/strict");
const { isDeepStrictEqual } = require("node:util");

// Runtime body from the task's takeWhile owner. Type erasure also removes
// adaptation 30's non-null assertion on array[index]. No adapter is built.
function original(array, predicate) {
    if (array !== undefined) {
        const len = array.length;
        let index = 0;
        while (index < len && predicate(array[index])) {
            index++;
        }
        return array.slice(0, index);
    }
}

// The proposed implementation, with only TypeScript annotations erased.
function proposed(array, predicate) {
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

if (process.argv.length > 2) {
    assert.equal(process.argv.length, 4);
    const ts = require('typescript');
    function extracted(tree) {
        const text = fs.readFileSync(tree + '/src/compiler/core.ts', 'utf8');
        const source = ts.createSourceFile('core.ts', text, ts.ScriptTarget.Latest, true);
        const owners = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === 'takeWhile');
        assert.equal(owners.length, 3);
        const owner = text.slice(owners[0].getStart(source), owners.at(-1).end).replaceAll('export ', '');
        const code = ts.transpileModule(owner, { compilerOptions: { target: ts.ScriptTarget.ES2024 } }).outputText;
        return new Function(code + '; return takeWhile;')();
    }
    original = extracted(process.argv[2]);
    proposed = extracted(process.argv[3]);
}

function describe(value) {
    if (value === undefined) return { kind: "undefined" };
    if (!Array.isArray(value)) return value;
    return {
        length: value.length,
        ownKeys: Object.keys(value),
        slots: Array.from({ length: value.length }, (_, index) => ({
            own: Object.hasOwn(value, index), value: describe(value[index]),
        })),
    };
}
const classes = [
    ["undefined array", () => undefined, () => true],
    ["empty array", () => [], () => true],
    ["all accepted", () => [1, 2, 3], () => true],
    ["none accepted", () => [1, 2, 3], () => false],
    ["middle rejection", () => [1, 2, 3], value => value < 2],
    ["sparse hole rejected", () => new Array(1), value => value !== undefined],
    ["sparse hole accepted", () => new Array(1), value => value === undefined],
];
const rows = classes.map(([name, input, predicate]) => {
    const beforeCalls = [], afterCalls = [];
    const before = original(input(), value => { beforeCalls.push(value); return predicate(value); });
    const after = proposed(input(), value => { afterCalls.push(value); return predicate(value); });
    const sameResult = isDeepStrictEqual(before, after);
    const sameCalls = isDeepStrictEqual(beforeCalls, afterCalls);
    assert.equal(sameCalls, true, name);
    assert.equal(sameResult, name !== "sparse hole accepted", name);
    return { name, input: describe(input()), before: describe(before), after: describe(after),
        beforeCalls: describe(beforeCalls), afterCalls: describe(afterCalls), sameResult, sameCalls };
});
// Independent observable property test, beyond the deep equality comparison.
const sparseBefore = original(new Array(1), () => true);
const sparseAfter = proposed(new Array(1), () => true);
assert.equal(0 in sparseBefore, false);
assert.equal(0 in sparseAfter, true);
console.log(JSON.stringify({ node: process.version, scope: "Node witnesses only; not lane counts",
    counterexamples: rows.filter(row => !row.sameResult).length, rows }, null, 2));
