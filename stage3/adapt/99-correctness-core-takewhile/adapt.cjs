'use strict';
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const ts = require('typescript');
const owners = [
    {
        "file": "core.ts",
        "owner": "takeWhile",
        "beforeDeclarations": 3,
        "afterDeclarations": 3,
        "before": "export function takeWhile<T, U extends T>(array: readonly T[], predicate: (element: T) => element is U): U[];\n/** @internal */\nexport function takeWhile<T, U extends T>(array: readonly T[] | undefined, predicate: (element: T) => element is U): U[] | undefined;\nexport function takeWhile<T, U extends T>(array: readonly T[] | undefined, predicate: (element: T) => element is U): U[] | undefined {\n    if (array !== undefined) {\n        const len = array.length;\n        let index = 0;\n        while (index < len && predicate(array[index]!)) {\n            index++;\n        }\n        return array.slice(0, index) as U[];\n    }\n}",
        "after": "export function takeWhile<T, U extends T>(array: readonly T[], predicate: (element: T) => element is U): U[];\n/** @internal */\nexport function takeWhile<T, U extends T>(array: readonly T[] | undefined, predicate: (element: T) => element is U): U[] | undefined;\nexport function takeWhile<T, U extends T>(array: readonly T[] | undefined, predicate: (element: T) => element is U): U[] | undefined {\n    if (array !== undefined) {\n        const result: U[] = [];\n        const len = array.length;\n        let index = 0;\n        while (index < len) {\n            const element = array[index]!;\n            if (!predicate(element)) break;\n            result.push(element);\n            index++;\n        }\n        return result;\n    }\n}",
        "ownerBeforeSha256": "c2d18eaee22538a20b2107e70b0ba35797197225e2531be38dcd81a61da4cd37",
        "ownerAfterSha256": "3fc68dd2d17d7425b633e02eddadf131168bfe1fd32a9b01a7462a21143e390f",
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
        // Convert edit offsets back to raw bytes, preserving even mixed line
        // endings outside this owner. Normalize only the replacement owner.
        const rawOffset = position => {
            let index = 0;
            for (let normalized = 0; normalized < position; normalized++, index++) {
                if (raw[index] === '\r' && raw[index + 1] === '\n') index++;
            }
            return index;
        };
        let preserved = raw;
        for (const edit of state.edits.sort((a, b) => b.start - a.start)) {
            const start = rawOffset(edit.start), end = rawOffset(edit.end);
            const ending = raw.slice(start, end).includes('\r\n') ? '\r\n' : '\n';
            const replacement = edit.after.replace(/\n/g, ending);
            preserved = preserved.slice(0, start) + replacement + preserved.slice(end);
        }
        assert.equal(preserved.replace(/\r\n/g, '\n'), after, 'raw edit mapping drift');
        return {name, raw, after: preserved, ...state};
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
