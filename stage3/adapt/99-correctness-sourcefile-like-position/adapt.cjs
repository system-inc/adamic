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
function locate(text) {
    const pin = owners[0], source = parse(pin.file, text);
    const found = source.statements.filter(node => ts.isInterfaceDeclaration(node) && node.name.text === pin.owner);
    assert.equal(found.length, 1, `${pin.owner}: owner count drift`);
    const start = found[0].getStart(source), end = found[0].end;
    const spelling = text.slice(start, end), hash = digest(spelling);
    assert([pin.ownerBeforeSha256, pin.ownerAfterSha256].includes(hash), `${pin.owner}: changed or half-adapted owner`);
    const adapted = hash === pin.ownerAfterSha256;
    assert.equal(spelling, adapted ? pin.after : pin.before, `${pin.owner}: exact text drift`);
    return {start, end, adapted};
}
function transpile(text, removeComments = false) {
    return ts.transpileModule(text, {compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}}).outputText;
}
function rewrite(tree, restoring = false) {
    assert.equal(ts.version, '6.0.3');
    const pin = owners[0], name = path.join(tree, 'src/compiler', pin.file);
    const raw = fs.readFileSync(name, 'utf8'), text = raw.replace(/\r\n/g, '\n');
    const crlf = raw.includes('\r\n');
    assert(!crlf || !raw.replace(/\r\n/g, '').includes('\n'), 'mixed line endings');
    const state = locate(text), wanted = !restoring;
    if (state.adapted === wanted) return {files: 0, edits: 0, javascriptChanging: false};
    let after = text.slice(0, state.start) + (wanted ? pin.after : pin.before) + text.slice(state.end);
    assert.equal(locate(after).adapted, wanted, 'invalid planned result');
    for (const removeComments of [false, true]) assert.equal(transpile(after, removeComments), transpile(text, removeComments), 'types.ts JavaScript changed');
    if (crlf) after = after.replace(/\n/g, '\r\n');
    assert.equal(fs.readFileSync(name, 'utf8'), raw, 'source changed while planning');
    fs.writeFileSync(name, after);
    return {files: 1, edits: 1, javascriptChanging: false};
}
const adapt = tree => rewrite(tree);
module.exports = {adapt, rewrite, owners, parse, locate, transpile, digest};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
