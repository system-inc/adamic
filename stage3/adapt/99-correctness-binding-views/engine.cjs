'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require('typescript');
const digest = text => crypto.createHash('sha256').update(text).digest('hex');
function normalize(raw) {
    const text = raw.replace(/\r\n/g, '\n');
    assert(!text.includes('\r'), 'bare carriage return');
    assert(!raw.includes('\r\n') || !raw.replace(/\r\n/g, '').includes('\n'), 'mixed line endings');
    return text;
}
function parse(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${file}: parse error`);
    return source;
}
function inspect(pin, text) {
    const source = parse(pin.file, text);
    const hash = digest(text);
    assert([pin.ownerBeforeSha256, pin.ownerAfterSha256].includes(hash), `${pin.file}: owner drift, duplicate or half-adapted text`);
    const adapted = hash === pin.ownerAfterSha256;
    assert.equal(text, adapted ? pin.after : pin.before, `${pin.file}: exact text drift`);
    for (const owner of [...(pin.editedOwners || []), ...(pin.insertedOwner ? [pin.insertedOwner] : [])]) {
        const expected = adapted ? owner.after : owner.before;
        const found = source.statements.filter(node => node.name?.text === owner.owner);
        assert.equal(found.length, expected ? 1 : 0, `${owner.owner}: owner count drift`);
        const spelling = found.length ? found[0].getText(source) : '';
        assert.equal(spelling, expected, `${owner.owner}: exact owner drift`);
        assert.equal(digest(spelling), adapted ? owner.ownerAfterSha256 : owner.ownerBeforeSha256);
    }
    return adapted;
}
function transpile(text, removeComments) {
    return ts.transpileModule(text, {compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, verbatimModuleSyntax: true, removeComments}}).outputText;
}
function rewrite(tree, pins, restoring = false) {
    assert.equal(ts.version, '6.0.3');
    const plans = pins.map(pin => {
        const name = path.join(tree, 'src/compiler', pin.file);
        const raw = fs.readFileSync(name, 'utf8'), text = normalize(raw);
        const state = inspect(pin, text), after = restoring ? pin.before : pin.after;
        for (const removeComments of [false, true]) assert.equal(transpile(text, removeComments), transpile(after, removeComments), `${pin.file}: JavaScript changed`);
        assert.equal(inspect(pin, after), !restoring);
        return {name, raw, text, state, after: raw.includes('\r\n') ? after.replace(/\n/g, '\r\n') : after};
    });
    assert(plans.every(plan => plan.state === plans[0].state), 'half-adapted tree');
    for (const plan of plans) assert.equal(fs.readFileSync(plan.name, 'utf8'), plan.raw, 'source changed while planning');
    const changes = plans.filter(plan => plan.raw !== plan.after);
    for (const plan of changes) fs.writeFileSync(plan.name, plan.after);
    return {files: changes.length, edits: changes.length, javascriptChanging: false};
}
module.exports = {rewrite, normalize, inspect, parse, transpile, digest};
