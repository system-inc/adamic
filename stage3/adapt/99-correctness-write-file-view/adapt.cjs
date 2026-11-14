'use strict';
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const ts = require('typescript');
const owners = require('./owners.json');
const digest = text => crypto.createHash('sha256').update(text).digest('hex');
const oldRead = '    const originalWriteFile = sys.writeFile;';
const newRead = '    const writeFileView: { readonly writeFile: System["writeFile"]; } = sys;\n    const originalWriteFile = writeFileView.writeFile;';
function validatePins() {
    assert.equal(ts.version, '6.0.3');
    assert.equal(owners.length, 1, 'expected one whole owner');
    const pin = owners[0];
    assert.equal(pin.file, 'sys.ts');
    assert.equal(pin.owner, 'patchWriteFileEnsuringDirectory');
    assert.equal(digest(pin.before), pin.ownerBeforeSha256, 'before pin hash drift');
    assert.equal(digest(pin.after), pin.ownerAfterSha256, 'after pin hash drift');
    assert.equal(pin.before.split(oldRead).length, 2, 'capture count drift');
    assert.equal(pin.after, pin.before.replace(oldRead, newRead), 'unreviewed owner edit');
    return pin;
}
function parse(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${file}: parse error`);
    return source;
}
function ownerRange(text) {
    const pin = validatePins(), source = parse(pin.file, text), found = [];
    function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === pin.owner) found.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    assert.equal(found.length, 1, `${pin.owner}: owner count drift`);
    assert(found[0].body && found[0].parent === source, `${pin.owner}: moved or bodyless owner`);
    return {start: found[0].getStart(source), end: found[0].end};
}
function locate(text) {
    const pin = validatePins(), range = ownerRange(text);
    const spelling = text.slice(range.start, range.end), hash = digest(spelling);
    assert([pin.ownerBeforeSha256, pin.ownerAfterSha256].includes(hash), `${pin.owner}: changed or half-adapted owner`);
    const adapted = hash === pin.ownerAfterSha256;
    assert.equal(spelling, adapted ? pin.after : pin.before, `${pin.owner}: exact text drift`);
    return {...range, adapted};
}
function transpile(text, removeComments = false) {
    return ts.transpileModule(text, {compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}}).outputText;
}
function assertJavaScript(before, after) {
    for (const removeComments of [false, true]) {
        const a = transpile(before, removeComments), b = transpile(after, removeComments);
        const ar = ownerRange(a), br = ownerRange(b);
        const oldOwner = a.slice(ar.start, ar.end), newOwner = b.slice(br.start, br.end);
        const alias = '    const writeFileView = sys;\n    const originalWriteFile = writeFileView.writeFile;';
        assert.equal(oldOwner.split(oldRead).length, 2, 'emitted capture count drift');
        assert.equal(newOwner, oldOwner.replace(oldRead, alias), 'unexpected owner JavaScript change');
        assert.equal(b.slice(0, br.start) + oldOwner + b.slice(br.end), a, 'JavaScript outside owner changed');
    }
}
function rewrite(tree, restoring = false) {
    const pin = validatePins(), name = path.join(tree, 'src/compiler', pin.file);
    const raw = fs.readFileSync(name, 'utf8'), crlf = raw.includes('\r\n');
    assert(!/\r(?!\n)/.test(raw), 'bare CR line ending');
    assert(!crlf || !raw.replace(/\r\n/g, '').includes('\n'), 'mixed line endings');
    const text = raw.replace(/\r\n/g, '\n'), state = locate(text);
    const before = text.slice(0, state.start) + pin.before + text.slice(state.end);
    const after = text.slice(0, state.start) + pin.after + text.slice(state.end);
    assert.equal(locate(before).adapted, false);
    assert.equal(locate(after).adapted, true);
    assertJavaScript(before, after);
    const wanted = !restoring;
    if (state.adapted === wanted) return {files: 0, edits: 0, javascriptChanging: false, valueIdentical: true};
    let planned = wanted ? after : before;
    if (crlf) planned = planned.replace(/\n/g, '\r\n');
    assert.equal(fs.readFileSync(name, 'utf8'), raw, 'source changed while planning');
    fs.writeFileSync(name, planned);
    return {files: 1, edits: 1, javascriptChanging: true, valueIdentical: true};
}
const adapt = tree => rewrite(tree);
module.exports = {adapt, rewrite, owners, parse, ownerRange, locate, transpile, assertJavaScript, digest};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
