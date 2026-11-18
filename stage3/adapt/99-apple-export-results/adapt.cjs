'use strict';
// Type-only contracts for refusal family 5 (#pncjpqy), one owner at a time: each edit in owners.json is located through
// the parsed source, never by a coordinate. A "member" edit adds one member to a named interface; a "replace" edit
// swaps an exact text a stated number of times inside a named function. Before any write the whole set is read: every
// edit not yet made means adapt, every edit made means nothing to do, and anything else (a drifted owner, a missing or
// duplicate owner, half the set made) is refused with the file untouched.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');
const owners = require('./owners.json');

// The one declaration an edit's owner names: an interface (optionally picked by its heritage clause, for a name merged
// across declarations) or a function with a body.
function owner(source, edit) {
    const found = [];
    ts.forEachChild(source, function visit(node) {
        if (edit.kind === 'member' && ts.isInterfaceDeclaration(node) && node.name.text === edit.owner) {
            const heritage = (node.heritageClauses || []).map(clause => clause.getText(source)).join(' ');
            if (edit.heritage === undefined || heritage === edit.heritage) found.push(node);
        }
        if (edit.kind === 'replace' && ts.isFunctionDeclaration(node) && node.name?.text === edit.owner && node.body) found.push(node);
        ts.forEachChild(node, visit);
    });
    assert.equal(found.length, 1, `${edit.candidate} ${edit.file}: ${found.length} declarations of ${edit.owner}, want one`);
    return found[0];
}

function count(text, needle) {
    let n = 0;
    for (let at = text.indexOf(needle); at >= 0; at = text.indexOf(needle, at + needle.length)) n++;
    return n;
}

// state reads one edit on LF text: 'before', 'after', or a refusal.
function state(text, edit) {
    const source = ts.createSourceFile(edit.file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${edit.file}: parse error`);
    const node = owner(source, edit);
    if (edit.kind === 'member') {
        const named = node.members.filter(member => member.name && member.name.getText(source) === edit.member);
        if (named.length === 0) return {state: 'before', node, source};
        assert.equal(named.length, 1, `${edit.candidate} ${edit.owner}: ${named.length} members named ${edit.member}`);
        assert.equal(named[0].getFullText(source).trim(), edit.text, `${edit.candidate} ${edit.owner}.${edit.member}: a member of that name that isn't this contract`);
        return {state: 'after', node, source};
    }
    const span = text.slice(node.getStart(source), node.end);
    const old = count(span, edit.old), made = count(span, edit.new);
    if (old === edit.count && made === 0) return {state: 'before', node, source};
    if (old === 0 && made === edit.count) return {state: 'after', node, source};
    throw new assert.AssertionError({message: `${edit.candidate} ${edit.owner}: ${old} of the old text and ${made} of the new, want ${edit.count} and 0 or 0 and ${edit.count}`});
}

function apply(text, edit) {
    const {node, source} = state(text, edit);
    if (edit.kind === 'member') {
        const close = node.end - 1;
        assert.equal(text[close], '}', `${edit.owner}: no closing brace`);
        const first = node.members[0];
        const indent = first ? text.slice(text.lastIndexOf('\n', first.getStart(source)) + 1, first.getStart(source)) : '    ';
        const lineStart = text.lastIndexOf('\n', close) + 1;
        return text.slice(0, lineStart) + indent + edit.text + '\n' + text.slice(lineStart);
    }
    const start = node.getStart(source), end = node.end;
    return text.slice(0, start) + text.slice(start, end).split(edit.old).join(edit.new) + text.slice(end);
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const files = [...new Set(owners.edits.map(edit => edit.file))];
    const raw = new Map(files.map(file => [file, fs.readFileSync(path.join(tree, file), 'utf8')]));
    const texts = new Map([...raw].map(([file, text]) => [file, text.replace(/\r\n/g, '\n')]));
    const states = owners.edits.map(edit => state(texts.get(edit.file), edit).state);
    if (states.every(s => s === 'after')) return {files: 0, edits: 0, javascriptChanging: false};
    assert(states.every(s => s === 'before'), `half adapted: ${owners.edits.map((edit, i) => `${edit.candidate} ${states[i]}`).join(', ')}`);
    for (const edit of owners.edits) texts.set(edit.file, apply(texts.get(edit.file), edit));
    for (const edit of owners.edits) assert.equal(state(texts.get(edit.file), edit).state, 'after', `${edit.candidate}: invalid planned edit`);
    for (const file of files) {
        let after = texts.get(file);
        if (raw.get(file).includes('\r\n')) after = after.replace(/\n/g, '\r\n');
        assert.equal(fs.readFileSync(path.join(tree, file), 'utf8'), raw.get(file), `${file}: changed while planning`);
        fs.writeFileSync(path.join(tree, file), after);
    }
    return {files: files.length, edits: owners.edits.length, javascriptChanging: false};
}

// before is the text a tree's adapted files had before this adaptation: each edit undone, in reverse.
function before(texts) {
    const out = new Map(texts);
    for (const edit of [...owners.edits].reverse()) {
        const text = out.get(edit.file);
        if (edit.kind === 'member') {
            const at = text.indexOf(edit.text);
            assert(at >= 0 && count(text, edit.text) === 1, `${edit.candidate}: the member isn't there once`);
            const lineStart = text.lastIndexOf('\n', at) + 1, lineEnd = text.indexOf('\n', at) + 1;
            out.set(edit.file, text.slice(0, lineStart) + text.slice(lineEnd));
        } else {
            const {node, source} = state(text, edit);
            const start = node.getStart(source), end = node.end;
            out.set(edit.file, text.slice(0, start) + text.slice(start, end).split(edit.new).join(edit.old) + text.slice(end));
        }
    }
    return out;
}

module.exports = {adapt, before, owners, state};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
