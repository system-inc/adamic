'use strict';
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const ts = require('typescript');
const owners = [
    {
        "file": "core.ts",
        "owner": "flatten",
        "beforeDeclarations": 1,
        "afterDeclarations": 1,
        "before": "export function flatten<T>(array: T[][] | readonly (T | readonly T[] | undefined)[]): T[] {\n    const result = [];\n    for (let i = 0; i < array.length; i++) {\n        const v = array[i];\n        if (v) {\n            if (isArray(v)) {\n                addRange(result, v);\n            }\n            else {\n                result.push(v);\n            }\n        }\n    }\n    return result;\n}",
        "after": "export function flatten<T>(array: T[][] | readonly (T | readonly T[] | undefined)[]): T[] {\n    const result: T[] = [];\n    for (let i = 0; i < array.length; i++) {\n        const v = array[i];\n        if (v) {\n            if (isArray(v)) {\n                addRange(result, v);\n            }\n            else {\n                result.push(v);\n            }\n        }\n    }\n    return result;\n}",
        "ownerBeforeSha256": "ce3e295002878e282155c3e62fb39fa01ce965545a8af17b59043df3c7b037e3",
        "ownerAfterSha256": "921c5e87381130c3df3cfea87e14a2183653078bebfb53c3e8779dee1c4bd1fe"
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
