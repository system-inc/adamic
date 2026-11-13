'use strict';
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const ts = require('typescript');
const owners = require('./owners.json');
const digest = text => crypto.createHash('sha256').update(text).digest('hex');
function parse(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${file}: parse error`);
    return source;
}
function locate(file, text) {
    const source = parse(file, text), states = [], edits = [];
    for (const pin of owners.filter(owner => owner.file === file)) {
        const found = [];
        function visit(node) { if ((ts.isFunctionDeclaration(node) && node.body || ts.isClassDeclaration(node) || ts.isInterfaceDeclaration(node)) && node.name?.text === pin.owner) found.push(node); ts.forEachChild(node, visit); }
        visit(source);
        assert.equal(found.length, 1, `${pin.owner}: owner count drift`);
        assert(found[0].parent === source, `${pin.owner}: moved owner`);
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
        const name = path.join(tree, file), raw = fs.readFileSync(name, 'utf8');
        const text = raw.replace(/\r\n/g, '\n'), state = locate(file, text);
        let after = text;
        for (const edit of state.edits.sort((a, b) => b.start - a.start)) after = after.slice(0, edit.start) + edit.after + after.slice(edit.end);
        assert(locate(file, after).adapted, `${file}: invalid planned result`);
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
            preserved = preserved.slice(0, start) + edit.after.replace(/\n/g, ending) + preserved.slice(end);
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
