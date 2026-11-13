'use strict';
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const ts = require('typescript');
const owners = require('./owners.json');
const digest = text => crypto.createHash('sha256').update(text).digest('hex');
function parse(text) {
    const source = ts.createSourceFile('path.ts', text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'path.ts: parse error');
    return source;
}
function locate(text) {
    const source = parse(text), pin = owners[0];
    const found = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === pin.owner);
    assert([pin.beforeDeclarations, pin.afterDeclarations].includes(found.length), 'forEachAncestorDirectory: owner count drift');
    const first = found[0].getStart(source);
    // Pin the section comment, documentation, complete overload set and implementation.
    const start = text.lastIndexOf('//// Path Traversal', first);
    assert(start >= 0 && start <= first, 'forEachAncestorDirectory: missing owner prefix');
    const end = found.at(-1).end, spelling = text.slice(start, end), hash = digest(spelling);
    assert([pin.ownerBeforeSha256, pin.ownerAfterSha256].includes(hash), 'forEachAncestorDirectory: changed or half-adapted owner');
    const adapted = hash === pin.ownerAfterSha256;
    assert.equal(spelling, adapted ? pin.after : pin.before, 'forEachAncestorDirectory: exact owner text drift');
    return {start, end, adapted};
}
function transpile(text) {
    return ts.transpileModule(text, {compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext}}).outputText;
}
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const name = path.join(tree, 'src/compiler/path.ts'), raw = fs.readFileSync(name, 'utf8');
    const text = raw.replace(/\r\n/g, '\n'), state = locate(text);
    if (state.adapted) return {files: 0, edits: 0, javascriptChanging: false};
    let after = text.slice(0, state.start) + owners[0].after + text.slice(state.end);
    assert(locate(after).adapted, 'path.ts: invalid planned owner');
    assert.equal(transpile(after), transpile(text), 'path.ts: JavaScript changed');
    if (raw.includes('\r\n')) after = after.replace(/\n/g, '\r\n');
    assert.equal(fs.readFileSync(name, 'utf8'), raw, 'path.ts: source changed while planning');
    fs.writeFileSync(name, after);
    return {files: 1, edits: 1, javascriptChanging: false};
}
module.exports = {adapt, owners, locate, digest, transpile};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
