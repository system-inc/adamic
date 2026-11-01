'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');

// createWatchProgram's close() writes undefined into every cached entry's fileWatcher, a required field on FilePresentOnHost (#qg1ry2t).
// The write is guarded to the unknown variant, and that variant's optional field admits undefined (README.md).
const file = 'src/compiler/watchPublic.ts';
const declaration = {before: 'FileWatcher', after: 'FileWatcher | undefined'};
const write = {before: 'value.fileWatcher = undefined;', after: 'if (isFilePresenceUnknownOnHost(value)) value.fileWatcher = undefined;'};

function nodes(root, test) {
    const found = [];
    (function visit(node) {
        if (test(node)) found.push(node);
        ts.forEachChild(node, visit);
    })(root);
    return found;
}

function only(found, what) {
    assert.equal(found.length, 1, `${found.length} ${what}, not one`);
    return found[0];
}

function property(owner, name) {
    return only(owner.members.filter(m => ts.isPropertySignature(m) && ts.isIdentifier(m.name) && m.name.text === name), `${owner.name.text}.${name} members`);
}

function shapeOf(text, known, what) {
    const shape = Object.keys(known).find(key => known[key] === text);
    assert(shape, `${what} is ${text}, not an inventoried shape`);
    return shape;
}

function locate(fileName, text) {
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'cannot parse watchPublic.ts');
    // The implementation is the one createWatchProgram with a body; the other two are its overloads.
    const program = only(source.statements.filter(s => ts.isFunctionDeclaration(s) && s.name?.text === 'createWatchProgram' && s.body), 'createWatchProgram implementations');
    const statements = program.body.statements;
    const local = (test, name, what) => only(statements.filter(s => test(s) && s.name?.text === name), what);

    // FilePresentOnHost keeps its required field: this repair never touches it.
    const present = local(ts.isInterfaceDeclaration, 'FilePresentOnHost', 'FilePresentOnHost interfaces');
    const presentWatcher = property(present, 'fileWatcher');
    assert(!presentWatcher.questionToken && presentWatcher.type?.getText(source) === 'FileWatcher', `FilePresentOnHost.fileWatcher is ${presentWatcher.getText(source)}, not the required fileWatcher: FileWatcher`);
    assert.equal(nodes(source, n => ts.isInterfaceDeclaration(n) && n.name.text === 'FilePresenceUnknownOnHost').length, 1, 'another FilePresenceUnknownOnHost appeared');
    const unknown = local(ts.isInterfaceDeclaration, 'FilePresenceUnknownOnHost', 'FilePresenceUnknownOnHost interfaces');
    assert.equal(unknown.members.length, 2, 'FilePresenceUnknownOnHost grew or shrank');
    assert.equal(property(unknown, 'version').getText(source), 'version: false;', 'FilePresenceUnknownOnHost.version moved');
    const unknownWatcher = property(unknown, 'fileWatcher');
    assert(unknownWatcher.questionToken && unknownWatcher.type, `FilePresenceUnknownOnHost.fileWatcher is ${unknownWatcher.getText(source)}, not optional`);
    const declared = shapeOf(unknownWatcher.type.getText(source), declaration, 'FilePresenceUnknownOnHost.fileWatcher');

    // The guard's predicate: the unknown variant is the one whose version is a boolean.
    const predicate = local(ts.isFunctionDeclaration, 'isFilePresenceUnknownOnHost', 'isFilePresenceUnknownOnHost functions');
    assert.equal(predicate.parameters.map(p => p.getText(source)).join(', '), 'hostSourceFile: FileMayBePresentOnHost', 'isFilePresenceUnknownOnHost takes something else');
    assert.equal(predicate.type?.getText(source), 'hostSourceFile is FilePresenceUnknownOnHost', 'isFilePresenceUnknownOnHost is no longer the unknown variant\'s predicate');

    // close()'s clearMap over sourceFilesCache, and the write at the end of its callback.
    const close = local(ts.isFunctionDeclaration, 'close', 'close functions');
    const clearing = only(nodes(source, n => ts.isCallExpression(n) && ts.isIdentifier(n.expression) && n.expression.text === 'clearMap' && n.arguments.length === 2 && ts.isIdentifier(n.arguments[0]) && n.arguments[0].text === 'sourceFilesCache'), 'clearMap calls over sourceFilesCache');
    assert(ts.isExpressionStatement(clearing.parent) && clearing.parent.parent === close.body, 'the clearMap over sourceFilesCache left close()');
    const callback = clearing.arguments[1];
    assert(ts.isArrowFunction(callback) && callback.parameters.length === 1 && callback.parameters[0].getText(source) === 'value' && ts.isBlock(callback.body), 'the clearMap callback is no longer value => { ... }');
    assert.equal(callback.body.statements.length, 1, 'the clearMap callback grew or shrank');
    const branch = callback.body.statements[0];
    assert(ts.isIfStatement(branch) && !branch.elseStatement && branch.expression.getText(source) === 'value && value.fileWatcher' && ts.isBlock(branch.thenStatement), 'the callback no longer tests value && value.fileWatcher');
    const steps = branch.thenStatement.statements;
    assert.equal(steps.length, 2, 'the callback\'s branch grew or shrank');
    assert.equal(steps[0].getText(source), 'value.fileWatcher.close();', 'the watcher is no longer closed first');
    const site = steps[1];
    const written = shapeOf(site.getText(source), write, 'the write');
    if (written === 'after') {
        // Spelled exactly, but also the right nodes: a call to the predicate on value, then the bare write, no else.
        assert(ts.isIfStatement(site) && !site.elseStatement && ts.isCallExpression(site.expression) && ts.isIdentifier(site.expression.expression) && site.expression.expression.text === 'isFilePresenceUnknownOnHost' && ts.isExpressionStatement(site.thenStatement), 'the guard is not the predicate on value');
    }
    else assert(ts.isExpressionStatement(site), 'the write is not a statement');
    // No other write of undefined into a fileWatcher anywhere in the file.
    const writes = nodes(source, n => ts.isBinaryExpression(n) && n.operatorToken.kind === ts.SyntaxKind.EqualsToken && ts.isPropertyAccessExpression(n.left) && n.left.name.text === 'fileWatcher' && ts.isIdentifier(n.right) && n.right.text === 'undefined');
    assert.equal(writes.length, 1, `${writes.length} writes of undefined into a fileWatcher, not one`);
    // Either untouched or already adapted, never half: a widening without its guard writes into the required field unchecked.
    assert.equal(declared, written, `watchPublic.ts is partly adapted: the declaration is ${declared}, the write ${written}`);
    return {shape: written, declaration: unknownWatcher.type, site, source};
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const fileName = path.join(tree, file);
    const text = fs.readFileSync(fileName, 'utf8');
    const found = locate(fileName, text);
    const edits = found.shape === 'before' ? [
        {start: found.declaration.getStart(found.source), end: found.declaration.end, text: declaration.after},
        {start: found.site.getStart(found.source), end: found.site.end, text: write.after},
    ] : [];
    let result = text;
    for (const edit of edits.sort((a, b) => b.start - a.start)) result = result.slice(0, edit.start) + edit.text + result.slice(edit.end);
    assert.equal(locate(fileName, result).shape, 'after', 'the planned edits do not adapt the site');
    const grown = edits.length ? declaration.after.length - declaration.before.length + write.after.length - write.before.length : 0;
    assert.equal(result.length - text.length, grown, 'the planned edits touch other bytes');
    // The edits hold no line break, so every line ending stays as it was, CRLF included.
    assert.equal(result.split('\r\n').length, text.split('\r\n').length, 'the planned edits moved a CRLF');
    assert.equal(result.split('\n').length, text.split('\n').length, 'the planned edits moved a line break');
    assert.equal(fs.readFileSync(fileName, 'utf8'), text, 'source changed while planning');
    if (edits.length) fs.writeFileSync(fileName, result);
    return {files: edits.length ? 1 : 0, edits: edits.length, javascriptChanging: true, behaviorChanging: false};
}
module.exports = {adapt, locate, file, declaration, write};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
