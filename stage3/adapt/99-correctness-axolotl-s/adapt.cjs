'use strict';
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const ts = require('typescript');
const owners = [
  {
    "file": "core.ts",
    "owner": "skipWhile",
    "beforeDeclarations": 3,
    "afterDeclarations": 3,
    "before": "export function skipWhile<T, U extends T>(array: readonly T[], predicate: (element: T) => element is U): Exclude<T, U>[];\n/** @internal */\nexport function skipWhile<T, U extends T>(array: readonly T[] | undefined, predicate: (element: T) => element is U): Exclude<T, U>[] | undefined;\n/** @internal */\nexport function skipWhile<T, U extends T>(array: readonly T[] | undefined, predicate: (element: T) => element is U): Exclude<T, U>[] | undefined {\n    if (array !== undefined) {\n        const len = array.length;\n        let index = 0;\n        while (index < len && predicate(array[index]!)) {\n            index++;\n        }\n        return array.slice(index) as Exclude<T, U>[];\n    }\n}",
    "after": "export function skipWhile<T, U extends T>(array: readonly T[], predicate: (element: T) => element is U): T[];\n/** @internal */\nexport function skipWhile<T, U extends T>(array: readonly T[] | undefined, predicate: (element: T) => element is U): T[] | undefined;\n/** @internal */\nexport function skipWhile<T, U extends T>(array: readonly T[] | undefined, predicate: (element: T) => element is U): T[] | undefined {\n    if (array !== undefined) {\n        const len = array.length;\n        let index = 0;\n        while (index < len && predicate(array[index]!)) {\n            index++;\n        }\n        return array.slice(index);\n    }\n}",
    "ownerBeforeSha256": "3a611286b99646ea467f1cdbb178fe25a8cb63b8d5b940e7fa018df9ab4c9f6d",
    "ownerAfterSha256": "112630165ee7ce029b3e999cae1630585cf422c11c030077a864cd0c3a357fd6",
    "siteEdits": 4
  },
  {
    "file": "core.ts",
    "owner": "filterMutate",
    "beforeDeclarations": 1,
    "afterDeclarations": 1,
    "before": "export function filterMutate<T>(array: T[], f: (x: T, i: number, array: T[]) => boolean): void {\n    let outIndex = 0;\n    for (let i = 0; i < array.length; i++) {\n        if (f(array[i]!, i, array)) {\n            array[outIndex] = array[i]!;\n            outIndex++;\n        }\n    }\n    array.length = outIndex;\n}",
    "after": "export function filterMutate<T>(array: T[], f: (x: T, i: number, array: T[]) => boolean): void {\n    let outIndex = 0;\n    for (let i = 0; i < array.length; i++) {\n        if (f(array[i]!, i, array)) {\n            array[outIndex] = array[i]!;\n            outIndex++;\n        }\n    }\n    array.splice(outIndex);\n}",
    "ownerBeforeSha256": "b3a7d0ae4b108d60dbaeb4f65a7aae3571c02fd5dc8884c34350492f6cc29e04",
    "ownerAfterSha256": "089adb3ecf66c7702a45ec645a13fee228f61148f488ac2a1b97441a9c3183ff",
    "siteEdits": 1
  },
  {
    "file": "core.ts",
    "owner": "clear",
    "beforeDeclarations": 1,
    "afterDeclarations": 1,
    "before": "export function clear(array: unknown[]): void {\n    array.length = 0;\n}",
    "after": "export function clear(array: unknown[]): void {\n    array.splice(0);\n}",
    "ownerBeforeSha256": "cc9708ccbf8476c9a1e88601ad332bdbfb220ae809ea69e783a16dd82334aef1",
    "ownerAfterSha256": "eff6fc8e41dc143e1913ca9abf97a22b67b784d9c8ba9240f8e59fa0599f1949",
    "siteEdits": 1
  },
  {
    "file": "core.ts",
    "owner": "createQueue",
    "beforeDeclarations": 1,
    "afterDeclarations": 1,
    "before": "export function createQueue<T>(items?: readonly T[]): Queue<T> {\n    const elements: (T | undefined)[] = items?.slice() ?? [];\n    let headIndex = 0;\n\n    function isEmpty() {\n        return headIndex === elements.length;\n    }\n\n    function enqueue(...items: T[]) {\n        elements.push(...items);\n    }\n\n    function dequeue(): T {\n        if (isEmpty()) {\n            throw new Error(\"Queue is empty\");\n        }\n\n        const result = elements[headIndex] as T;\n        elements[headIndex] = undefined; // Don't keep referencing dequeued item\n        headIndex++;\n\n        // If more than half of the queue is empty, copy the remaining elements to the\n        // front and shrink the array (unless we'd be saving fewer than 100 slots)\n        if (headIndex > 100 && headIndex > (elements.length >> 1)) {\n            const newLength = elements.length - headIndex;\n            elements.copyWithin(/*target*/ 0, /*start*/ headIndex);\n\n            elements.length = newLength;\n            headIndex = 0;\n        }\n\n        return result;\n    }\n\n    return {\n        enqueue,\n        dequeue,\n        isEmpty,\n    };\n}",
    "after": "export function createQueue<T>(items?: readonly T[]): Queue<T> {\n    const elements: (T | undefined)[] = items?.slice() ?? [];\n    let headIndex = 0;\n\n    function isEmpty() {\n        return headIndex === elements.length;\n    }\n\n    function enqueue(...items: T[]) {\n        elements.push(...items);\n    }\n\n    function dequeue(): T {\n        if (isEmpty()) {\n            throw new Error(\"Queue is empty\");\n        }\n\n        const result = elements[headIndex] as T;\n        elements[headIndex] = undefined; // Don't keep referencing dequeued item\n        headIndex++;\n\n        // If more than half of the queue is empty, copy the remaining elements to the\n        // front and shrink the array (unless we'd be saving fewer than 100 slots)\n        if (headIndex > 100 && headIndex > (elements.length >> 1)) {\n            const newLength = elements.length - headIndex;\n            elements.copyWithin(/*target*/ 0, /*start*/ headIndex);\n\n            elements.splice(newLength);\n            headIndex = 0;\n        }\n\n        return result;\n    }\n\n    return {\n        enqueue,\n        dequeue,\n        isEmpty,\n    };\n}",
    "ownerBeforeSha256": "f267f710a503ed24b60ac4c42774a8600864594c1f68caa7a695c0aac90e8dd2",
    "ownerAfterSha256": "19ae68b55f7e220fbfc88ae005692ded05c4dfa42d6cd5b194c44a617ae3ca09",
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
        if (raw.includes('\r\n')) after = after.replace(/\n/g, '\r\n');
        return {name, raw, after, ...state};
    });
    assert(plans.every(plan => plan.adapted === plans[0].adapted), 'half-adapted tree');
    for (const plan of plans) assert.equal(fs.readFileSync(plan.name, 'utf8'), plan.raw, 'source changed while planning');
    if (!plans[0].adapted) for (const plan of plans) fs.writeFileSync(plan.name, plan.after);
    return {files: plans[0].adapted ? 0 : plans.length, edits: plans.reduce((sum, plan) => sum + plan.edits.length, 0), javascriptChanging: true};
}
module.exports = {adapt, owners, parse, transpile, digest};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
