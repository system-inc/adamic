'use strict';
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const ts = require('typescript');
const owners = require('./owners.json');
const digest = text => crypto.createHash('sha256').update(text).digest('hex');
function locate(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${file}: parse error`);
    const pin = owners[0];
    assert.equal(owners.length, 1, 'arrayOf: expected one owner pin');
    assert.equal(digest(pin.before), pin.ownerBeforeSha256, 'arrayOf: invalid before pin');
    assert.equal(digest(pin.after), pin.ownerAfterSha256, 'arrayOf: invalid after pin');
    const found = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === pin.owner);
    assert.equal(found.length, 1, `arrayOf: expected one top-level owner, found ${found.length}`);
    const start = found[0].getStart(source), end = found[0].end;
    const rawOwner = text.slice(start, end), spelling = rawOwner.replace(/\r\n/g, '\n');
    const eol = rawOwner.includes('\r\n') ? '\r\n' : '\n';
    assert.equal(rawOwner, spelling.replace(/\n/g, eol), 'arrayOf: mixed line endings inside owner');
    const hash = digest(spelling);
    assert([pin.ownerBeforeSha256, pin.ownerAfterSha256].includes(hash), 'arrayOf: changed or half-adapted owner');
    const adapted = hash === pin.ownerAfterSha256;
    assert.equal(spelling, adapted ? pin.after : pin.before, 'arrayOf: exact owner text drift');
    return {start, end, adapted, eol};
}
function adapt(tree) {
    assert.equal(ts.version, '6.0.3', 'arrayOf: expected TypeScript 6.0.3 API');
    const pin = owners[0], name = path.join(tree, 'src/compiler', pin.file);
    const raw = fs.readFileSync(name, 'utf8'), state = locate(name, raw);
    if (state.adapted) return {files: 0, edits: 0, javascriptChanging: true};
    const after = raw.slice(0, state.start) + pin.after.replace(/\n/g, state.eol) + raw.slice(state.end);
    assert(locate(name, after).adapted, 'arrayOf: invalid planned result');
    assert.equal(fs.readFileSync(name, 'utf8'), raw, 'arrayOf: source changed while planning');
    fs.writeFileSync(name, after);
    return {files: 1, edits: 1, javascriptChanging: true};
}
module.exports = {adapt, owners, locate, digest};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
