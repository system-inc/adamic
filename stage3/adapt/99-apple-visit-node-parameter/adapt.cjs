'use strict';
// #ytagqs5: visitorPublic.ts's visitNode implementation takes `node: Node | undefined`, the value its body already
// checks first. Overload 1 admits `TIn extends Node | undefined`, and the declared implementation took `node: Node`,
// so the overload census refused every unit that reaches visitNode by overload 1. The body's first statement returns
// `node` when it is undefined, so the edit is type-only: the JavaScript is the same, and no declaration moves, since
// the implementation of an overloaded function is never emitted as a declaration.
// The owner is the one top-level visitNode with a body, right after its two overloads, pinned whole before and after
// (SHA-256 and exact text, on LF). The edit is one exact text, counted once inside it; anything else is refused before
// any write, CRLF is kept, and a second application writes nothing.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require('typescript');
const owners = require('./owners.json');
const digest = text => crypto.createHash('sha256').update(text).digest('hex');

function count(text, needle) {
    let n = 0;
    for (let at = text.indexOf(needle); at >= 0; at = text.indexOf(needle, at + needle.length)) n++;
    return n;
}

// locate reads LF text: the owner's span and whether it is adapted, or a refusal.
function locate(text) {
    const pin = owners[0];
    const source = ts.createSourceFile('visitorPublic.ts', text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'visitorPublic.ts: parse error');
    const named = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === pin.owner);
    const bodies = named.filter(node => node.body);
    assert.equal(bodies.length, 1, 'visitNode: missing or duplicate implementation');
    const implementation = bodies[0];
    // The implementation serves exactly the two overloads that stand right before it.
    const at = source.statements.indexOf(implementation);
    assert.equal(named.length, 3, 'visitNode: not two overloads and an implementation');
    assert(at >= 2 && source.statements[at - 2] === named[0] && source.statements[at - 1] === named[1] && named[2] === implementation,
        'visitNode: the implementation does not follow its two overloads');
    assert(implementation.modifiers?.some(modifier => modifier.kind === ts.SyntaxKind.ExportKeyword), 'visitNode: not exported');
    const start = implementation.getStart(source), end = implementation.end;
    const spelling = text.slice(start, end), hash = digest(spelling);
    assert([pin.ownerBeforeSha256, pin.ownerAfterSha256].includes(hash), 'visitNode: changed or half-adapted owner');
    const adapted = hash === pin.ownerAfterSha256;
    assert.equal(spelling, adapted ? pin.after : pin.before, 'visitNode: exact owner text drift');
    assert.deepEqual([count(spelling, pin.old), count(spelling, pin.new)], adapted ? [0, pin.count] : [pin.count, 0],
        'visitNode: the edit is not its one text');
    return {start, end, adapted};
}

// typeOnly holds the planned file to the before file's JavaScript and declarations, both byte for byte.
function typeOnly(before, after) {
    for (const target of [ts.ScriptTarget.ES2020, ts.ScriptTarget.ESNext]) {
        for (const removeComments of [true, false]) {
            const emit = text => ts.transpileModule(text, {fileName: owners[0].file, compilerOptions: {target, module: ts.ModuleKind.ESNext, removeComments}}).outputText;
            assert.equal(emit(after), emit(before), `visitorPublic.ts: JavaScript changed at ${ts.ScriptTarget[target]}`);
        }
    }
    for (const stripInternal of [true, false]) {
        const emit = text => ts.transpileDeclaration(text, {fileName: owners[0].file, compilerOptions: {stripInternal}}).outputText;
        assert.equal(emit(after), emit(before), 'visitorPublic.ts: declarations changed');
    }
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const file = path.join(tree, owners[0].file), raw = fs.readFileSync(file, 'utf8');
    const text = raw.replace(/\r\n/g, '\n'), state = locate(text);
    if (state.adapted) return {files: 0, edits: 0, javascriptChanging: false};
    let after = text.slice(0, state.start) + owners[0].after + text.slice(state.end);
    assert(locate(after).adapted, 'visitorPublic.ts: invalid planned owner');
    assert.equal(after, text.slice(0, state.start) + owners[0].before.replace(owners[0].old, owners[0].new) + text.slice(state.end),
        'visitorPublic.ts: the planned owner is not the one edit');
    typeOnly(text, after);
    if (raw.includes('\r\n')) after = after.replace(/\n/g, '\r\n');
    assert.equal(fs.readFileSync(file, 'utf8'), raw, 'visitorPublic.ts: changed while planning');
    fs.writeFileSync(file, after);
    return {files: 1, edits: 1, javascriptChanging: false};
}

module.exports = {adapt, owners, locate, typeOnly, digest};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
