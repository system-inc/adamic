'use strict';
const assert = require('node:assert/strict');
const { createHash } = require('node:crypto');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const [pin] = require('./owners.json');

const original = 'namedCapturingGroupsScopeStack.some(group => group?.has(tokenValue))';
const adapted = 'namedCapturingGroupsScopeStack.some(group => group !== undefined && group.has(tokenValue))';
const digest = text => createHash('sha256').update(text).digest('hex');

function locate(text) {
    assert.equal(ts.version, '6.0.3');
    assert(!/\r(?!\n)/.test(text), `${pin.file}: bare CR`);
    assert(!(text.includes('\r\n') && /(?<!\r)\n/.test(text)), `${pin.file}: mixed line endings`);
    const source = ts.createSourceFile(pin.file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${pin.file}: parse diagnostics`);
    const found = [];
    function visit(node) {
        if (ts.SyntaxKind[node.kind] === pin.kind && node.name?.text === pin.owner) found.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    assert.equal(found.length, 1, `${pin.file}:${pin.owner}: missing or duplicate owner`);
    const owner = found[0];
    return {source, owner, start: owner.getStart(source), end: owner.end};
}

function callback(text, expression) {
    const {source, owner} = locate(text);
    const found = [];
    function visit(node) {
        if (ts.isCallExpression(node) && node.getText(source) === expression) found.push(node);
        ts.forEachChild(node, visit);
    }
    visit(owner);
    assert.equal(found.length, 1, 'expected one pinned some expression');
    const call = found[0];
    const arrow = call.arguments[0];
    assert(call.arguments.length === 1 && ts.isArrowFunction(arrow), 'expected sole arrow argument');
    return arrow.getText(source);
}

function adapt(tree) {
    assert.equal(digest(pin.before), pin.ownerBeforeSha256, 'before pin corrupt');
    assert.equal(digest(pin.after), pin.ownerAfterSha256, 'after pin corrupt');
    assert.equal(pin.before.split(original).length, 2, 'original expression count');
    assert.equal(pin.after.split(adapted).length, 2, 'adapted expression count');
    assert(!pin.before.includes(adapted) && !pin.after.includes(original), 'mixed expressions in pin');
    assert(digest(pin.before.replace(original, adapted)) === pin.ownerAfterSha256 &&
        pin.before.replace(original, adapted) === pin.after, 'expected one expression change');
    const file = path.join(tree, pin.file);
    const raw = fs.readFileSync(file, 'utf8');
    const {start, end} = locate(raw);
    const current = raw.slice(start, end);
    assert(current === pin.before || current === pin.after, `${pin.file}:${pin.owner}: drift or half-adapted owner`);
    const result = current === pin.after ? raw : raw.slice(0, start) + pin.after + raw.slice(end);
    assert.equal(fs.readFileSync(file, 'utf8'), raw, 'source changed while planning');
    const writes = Number(result !== raw);
    if (writes) fs.writeFileSync(file, result);
    return {files: writes, writes, owners: writes, javascriptChanging: true};
}

module.exports = {adapt, callback};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
