'use strict';
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const ts = require('typescript');
const owners = require('./owners.json');
const digest = text => crypto.createHash('sha256').update(text).digest('hex');
function locate(text) {
    const source = ts.createSourceFile('sys.ts', text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'sys.ts: parse error');
    const pin = owners[0];
    const found = source.statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === pin.owner);
    assert.equal(found.length, 1, 'setCustomPollingValues: missing or duplicate owner');
    const start = found[0].getStart(source), end = found[0].end;
    const spelling = text.slice(start, end), hash = digest(spelling);
    assert([pin.ownerBeforeSha256, pin.ownerAfterSha256].includes(hash), 'setCustomPollingValues: changed or half-adapted owner');
    const adapted = hash === pin.ownerAfterSha256;
    assert.equal(spelling, adapted ? pin.after : pin.before, 'setCustomPollingValues: exact owner text drift');
    return {start, end, adapted};
}
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const file = path.join(tree, owners[0].file), raw = fs.readFileSync(file, 'utf8');
    const text = raw.replace(/\r\n/g, '\n'), state = locate(text);
    if (state.adapted) return {files: 0, edits: 0, javascriptChanging: true};
    let after = text.slice(0, state.start) + owners[0].after + text.slice(state.end);
    assert(locate(after).adapted, 'sys.ts: invalid planned owner');
    if (raw.includes('\r\n')) after = after.replace(/\n/g, '\r\n');
    assert.equal(fs.readFileSync(file, 'utf8'), raw, 'sys.ts: changed while planning');
    fs.writeFileSync(file, after);
    return {files: 1, edits: 1, javascriptChanging: true};
}
module.exports = {adapt, owners, locate, digest};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
