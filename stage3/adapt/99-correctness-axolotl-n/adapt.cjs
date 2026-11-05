'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require('typescript');
const pins = require('./pins.json');
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
function functions(text) {
    const source = ts.createSourceFile('debug.ts', text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'debug.ts does not parse');
    const found = new Map();
    function visit(node, ancestry) {
        const named = (ts.isFunctionDeclaration(node) && node.body || ts.isModuleDeclaration(node)) && node.name;
        const chain = named ? [...ancestry, node.name.text] : ancestry;
        if (ts.isFunctionDeclaration(node) && node.body && node.name) {
            const key = chain.join('.');
            assert(!found.has(key), `duplicate owner ${key}`);
            found.set(key, {node, start: node.getStart(source), end: node.end, text: node.getText(source)});
        }
        ts.forEachChild(node, child => visit(child, chain));
    }
    visit(source, []);
    return found;
}
function inspect(text) {
    const owners = functions(text);
    return pins.map(pin => {
        const owner = owners.get(pin.owner);
        assert(owner, `missing owner ${pin.owner}`);
        const digest = hash(owner.text);
        const state = digest === pin.beforeSha256 ? 'before' : digest === pin.afterSha256 ? 'after' : undefined;
        assert(state, `changed owner ${pin.owner}`);
        return {pin, owner, state};
    });
}
function javascript(text, removeComments) {
    return ts.transpileModule(text, {compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}}).outputText;
}
function plan(text) {
    assert.equal(ts.version, '6.0.3');
    const sites = inspect(text);
    assert(sites.every(site => site.state === sites[0].state), 'half-adapted debug.ts');
    if (sites[0].state === 'after') return {after: text, edits: 0};
    let after = text;
    let edits = 0;
    for (const {pin, owner} of [...sites].sort((a,b) => b.owner.start - a.owner.start)) {
        let replacement = owner.text;
        for (const edit of pin.edits) {
            assert.equal(replacement.split(edit.before).length - 1, 1, `exact text moved in ${pin.owner}`);
            replacement = replacement.replace(edit.before, edit.after);
            edits++;
        }
        assert.equal(hash(replacement), pin.afterSha256, `planned owner changed ${pin.owner}`);
        after = after.slice(0, owner.start) + replacement + after.slice(owner.end);
    }
    assert(inspect(after).every(site => site.state === 'after'), 'incomplete planned tree');
    for (const removeComments of [false, true]) assert.equal(javascript(after, removeComments), javascript(text, removeComments), 'debug.ts JavaScript changed');
    return {after, edits};
}
function adapt(tree) {
    const file = path.join(tree, 'src/compiler/debug.ts');
    const before = fs.readFileSync(file, 'utf8');
    const result = plan(before);
    assert.equal(fs.readFileSync(file, 'utf8'), before, 'source changed while planning');
    if (result.edits) fs.writeFileSync(file, result.after);
    return {files: result.edits ? 1 : 0, edits: result.edits, javascriptChanging: false};
}
module.exports = {adapt, plan, inspect, functions, hash};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
