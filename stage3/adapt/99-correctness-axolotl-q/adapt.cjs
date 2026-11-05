'use strict';
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const ts = require('typescript');
const owners = [
    {
        "file": "core.ts",
        "owner": "deduplicateRelational",
        "beforeDeclarations": 1,
        "afterDeclarations": 1,
        "before": "function deduplicateRelational<T>(array: readonly T[], equalityComparer: EqualityComparer<T>, comparer: Comparer<T>) {\n    // Perform a stable sort of the array. This ensures the first entry in a list of\n    // duplicates remains the first entry in the result.\n    const indices = indicesOf(array);\n    stableSortIndices(array, indices, comparer);\n\n    let last = array[indices[0]!]!;\n    const deduplicated: number[] = [indices[0]!];\n    for (let i = 1; i < indices.length; i++) {\n        const index = indices[i]!;\n        const item = array[index]!;\n        if (!equalityComparer(last, item)) {\n            deduplicated.push(index);\n            last = item;\n        }\n    }\n\n    // restore original order\n    deduplicated.sort();\n    return deduplicated.map(i => array[i]!);\n}",
        "after": "function deduplicateRelational<T>(array: readonly T[], equalityComparer: EqualityComparer<T>, comparer: Comparer<T>) {\n    // Perform a stable sort of the array. This ensures the first entry in a list of\n    // duplicates remains the first entry in the result.\n    const indices = indicesOf(array);\n    stableSortIndices(array, indices, comparer);\n\n    let last = array[indices[0]!]!;\n    const deduplicated: number[] = [indices[0]!];\n    for (let i = 1; i < indices.length; i++) {\n        const index = indices[i]!;\n        const item = array[index]!;\n        if (!equalityComparer(last, item)) {\n            deduplicated.push(index);\n            last = item;\n        }\n    }\n\n    // restore original order\n    deduplicated.sort();\n    return deduplicated.map<T>(i => array[i]!);\n}",
        "ownerBeforeSha256": "ca2797e92b590ee1ba2fe25f47de55dab4e0a6e2005715dc1f98024b1eb7fe38",
        "ownerAfterSha256": "e13cc74c6313093c3a6b5d4852ae9e57783c9965b160e22679d687d181142b6b",
        "siteEdits": 1
    },
    {
        "file": "core.ts",
        "owner": "indicesOf",
        "beforeDeclarations": 1,
        "afterDeclarations": 1,
        "before": "export function indicesOf(array: readonly unknown[]): number[] {\n    return array.map(selectIndex);\n}",
        "after": "export function indicesOf<T>(array: readonly T[]): number[] {\n    return array.map(selectIndex);\n}",
        "ownerBeforeSha256": "a7c6c1b0fc171942ddf5e205bb21e8da1d1409a235220908028a3fc53e36b214",
        "ownerAfterSha256": "dac2cfe38f47956695330dcaca37192cfff7115596c51f98d27c58a0359a3148",
        "siteEdits": 1
    },
    {
        "file": "core.ts",
        "owner": "cartesianProductWorker",
        "beforeDeclarations": 1,
        "afterDeclarations": 1,
        "before": "function cartesianProductWorker<T>(arrays: readonly (readonly T[])[], result: (readonly T[])[], outer: readonly T[] | undefined, index: number) {\n    for (const element of arrays[index]!) {\n        let inner: T[];\n        if (outer) {\n            inner = outer.slice();\n            inner.push(element);\n        }\n        else {\n            inner = [element];\n        }\n        if (index === arrays.length - 1) {\n            result.push(inner);\n        }\n        else {\n            cartesianProductWorker(arrays, result, inner, index + 1);\n        }\n    }\n}",
        "after": "function cartesianProductWorker<T>(arrays: readonly (readonly T[])[], result: T[][], outer: readonly T[] | undefined, index: number) {\n    for (const element of arrays[index]!) {\n        let inner: T[];\n        if (outer) {\n            inner = outer.slice();\n            inner.push(element);\n        }\n        else {\n            inner = [element];\n        }\n        if (index === arrays.length - 1) {\n            result.push(inner);\n        }\n        else {\n            cartesianProductWorker(arrays, result, inner, index + 1);\n        }\n    }\n}",
        "ownerBeforeSha256": "6c3583e061791c5e761229ab655c75d69401737c678591902d92dd8d3c675e0c",
        "ownerAfterSha256": "60a49d1150fd68287fe482dd0c86a7f4825780b87a43455281de938cd8c839b3",
        "siteEdits": 1
    },
    {
        "file": "core.ts",
        "owner": "isNodeLikeSystem",
        "beforeDeclarations": 1,
        "afterDeclarations": 1,
        "before": "export function isNodeLikeSystem(): boolean {\n    // This is defined here rather than in sys.ts to prevent a cycle from its\n    // use in performanceCore.ts.\n    return typeof process !== \"undefined\"\n        && !!process.nextTick\n        && !(process as any).browser\n        && typeof require !== \"undefined\";\n}",
        "after": "export function isNodeLikeSystem(): boolean {\n    // This is defined here rather than in sys.ts to prevent a cycle from its\n    // use in performanceCore.ts.\n    return typeof process !== \"undefined\"\n        && !!process.nextTick\n        && !process.browser\n        && typeof require !== \"undefined\";\n}",
        "ownerBeforeSha256": "f459dd087504c3cb88278d4dfb98b800485fff4d23f6b6707b24bee5f0001d11",
        "ownerAfterSha256": "204354b270f8515598cdddf741deaa724189f11e0a76f7f9c28a99217bb4e5f8",
        "siteEdits": 1
    }
];
const digest = text => crypto.createHash('sha256').update(text).digest('hex');
function parse(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${file}: parse error`);
    return source;
}
function locate(file, text) {
    const source = parse(file, text), states = [], edits = [];
    for (const pin of owners.filter(owner => owner.file === file)) {
        const found = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === pin.owner);
        assert([pin.beforeDeclarations, pin.afterDeclarations].includes(found.length), `${pin.owner}: owner count drift`);
        let total = 0; function visit(node) { if (ts.isFunctionDeclaration(node) && node.name?.text === pin.owner) total++; ts.forEachChild(node, visit); } visit(source);
        assert.equal(total, found.length, pin.owner + ': nested owner');
        const start = found[0].getStart(source), end = found.at(-1).end;
        const spelling = text.slice(start, end), hash = digest(spelling);
        assert([pin.ownerBeforeSha256, pin.ownerAfterSha256].includes(hash), `${pin.owner}: changed or half-adapted owner`);
        const adapted = hash === pin.ownerAfterSha256;
        assert.equal(spelling, adapted ? pin.after : pin.before, `${pin.owner}: exact text drift`);
        states.push(adapted);
        if (!adapted) edits.push({start, end, after: pin.after});
    }
    assert(states.length && states.every(state => state === states[0]), `${file}: half-adapted file`);
    return {adapted: states[0], edits};
}
function transpile(text, removeComments = false) {
    return ts.transpileModule(text, {compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}}).outputText;
}
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const plans = [...new Set(owners.map(owner => owner.file))].map(file => {
        const name = path.join(tree, 'src/compiler', file), raw = fs.readFileSync(name, 'utf8');
        const text = raw.replace(/\r\n/g, '\n'), state = locate(file, text);
        let after = text;
        for (const edit of state.edits.sort((a, b) => b.start - a.start)) after = after.slice(0, edit.start) + edit.after + after.slice(edit.end);
        assert(locate(file, after).adapted, `${file}: invalid planned result`);
        if (file === 'core.ts') for (const removeComments of [false, true]) assert.equal(transpile(after, removeComments), transpile(text, removeComments), 'core.ts JavaScript changed');
        if (raw.includes('\r\n')) after = after.replace(/\n/g, '\r\n');
        return {name, raw, after, ...state};
    });
    assert(plans.every(plan => plan.adapted === plans[0].adapted), 'half-adapted tree');
    for (const plan of plans) assert.equal(fs.readFileSync(plan.name, 'utf8'), plan.raw, 'source changed while planning');
    if (!plans[0].adapted) for (const plan of plans) fs.writeFileSync(plan.name, plan.after);
    return {files: plans[0].adapted ? 0 : plans.length, edits: plans.reduce((sum, plan) => sum + plan.edits.length, 0), javascriptChanging: false};
}
module.exports = {adapt, owners, parse, transpile, digest};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
