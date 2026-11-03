'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require('typescript');
const {compiler, nodes} = require('../47-host-errors/adapt.cjs');
const manifest = require('./proofs/targets.json');
const local = '_crypto';
const importText = 'import * as _crypto from "node:crypto";';
const declarationText = 'let _crypto: typeof import("crypto") | undefined;';
const blockLines = [
    '        // crypto can be absent on reduced node installations',
    '        ' + declarationText,
    '        try {',
    '            _crypto = require("crypto");',
    '        }',
    '        catch {',
    '            _crypto = undefined;',
    '        }',
];
function hash(text) { return crypto.createHash('sha256').update(text).digest('hex'); }
function one(found, label) { assert.equal(found.length, 1, `${label}: want one, got ${found.length}`); return found[0]; }
function validateBindings(program, file, done) {
    const checker = program.getTypeChecker();
    const nodeSystem = one(nodes(ts, file, n => ts.isFunctionDeclaration(n) && n.name?.text === 'getNodeSystem'), 'getNodeSystem');
    const imports = file.statements.filter(ts.isImportDeclaration);
    const existing = imports.filter(n => n.moduleSpecifier.text === 'node:crypto');
    const locals = nodeSystem.body.statements.filter(ts.isVariableStatement).filter(s => s.declarationList.declarations.some(d => d.name.getText(file) === local));
    let binding, declaration;
    if (done) {
        assert.equal(locals.length, 0, 'crypto local remains');
        const imported = one(existing, 'crypto import');
        assert.equal(imported.getText(file), importText, 'crypto import moved');
        assert(ts.isNamespaceImport(imported.importClause.namedBindings), 'crypto must be a namespace import');
        const at = imports.indexOf(imported);
        assert.deepEqual(imports.slice(at - 3, at).map(n => n.getText(file)), [
            'import * as _fs from "node:fs";', 'import * as _path from "node:path";', 'import * as _os from "node:os";',
        ], '93 import neighbors moved');
        binding = imported.importClause.namedBindings.name;
    } else {
        assert.equal(existing.length, 0, 'crypto already imported');
        declaration = one(locals, 'crypto local');
        assert.equal(declaration.getText(file), declarationText, 'crypto declaration moved');
        assert.equal(declaration.declarationList.flags & ts.NodeFlags.Let, ts.NodeFlags.Let);
        assert.equal(declaration.declarationList.declarations.length, 1);
        binding = declaration.declarationList.declarations[0].name;
    }
    const symbol = checker.getSymbolAtLocation(binding);
    assert(symbol, 'crypto binding has no symbol');
    const identifiers = nodes(ts, file, n => ts.isIdentifier(n) && n.text === local);
    const uses = identifiers.filter(n => n !== binding);
    const roles = [];
    for (const use of uses) {
        const parent = use.parent;
        let role;
        if (done) {
            assert(ts.isPropertyAccessExpression(parent) && parent.expression === use, 'collision or bare namespace use: _crypto');
            assert.equal(parent.name.text, 'createHash', 'crypto member moved');
            role = 'qualified receiver';
        } else if (ts.isBinaryExpression(parent) && parent.left === use && parent.operatorToken.kind === ts.SyntaxKind.EqualsToken) {
            assert(['require("crypto")', 'undefined'].includes(parent.right.getText(file)), 'crypto assignment moved');
            role = 'assignment';
        } else if (ts.isConditionalExpression(parent) && parent.condition === use) {
            assert.equal(parent.whenTrue.getText(file), 'createSHA256Hash');
            assert(['generateDjb2Hash', 'undefined'].includes(parent.whenFalse.getText(file)), 'crypto conditional moved');
            role = 'condition';
        } else {
            assert(ts.isNonNullExpression(parent) && ts.isPropertyAccessExpression(parent.parent) && parent.parent.expression === parent && parent.parent.name.text === 'createHash', 'collision or bare namespace use: _crypto');
            role = 'checked qualified receiver';
        }
        assert.equal(checker.getSymbolAtLocation(use), symbol, 'captured or shadowed use: _crypto');
        roles.push(role);
    }
    assert.deepEqual(roles.sort(), done ? ['qualified receiver'] : ['assignment', 'assignment', 'checked qualified receiver', 'condition', 'condition'], 'crypto reference roles moved');
    const sha = one(nodes(ts, nodeSystem, n => ts.isFunctionDeclaration(n) && n.name?.text === 'createSHA256Hash'), 'createSHA256Hash function');
    const receiver = one(nodes(ts, sha, n => ts.isIdentifier(n) && n.text === local), 'hash receiver');
    const access = done ? receiver.parent : receiver.parent.parent;
    assert(ts.isCallExpression(access.parent) && access.parent.expression === access && access.parent.arguments.length === 1 && access.parent.arguments[0].getText(file) === '"sha256"', 'SHA-256 call moved');
    return {declaration, nodeSystem, receiver, evidence: {binding: done ? 'namespace import' : 'local let', references: uses.length, roles, allUsesResolveToBinding: true, otherDeclarations: 0, qualifiedUses: done ? 1 : 0, bareNamespaceUses: 0}};
}
function plannedProgram(program, file, text) {
    const host = ts.createCompilerHost(program.getCompilerOptions(), true);
    const original = host.getSourceFile.bind(host);
    host.getSourceFile = (name, language, onError, shouldCreateNewSourceFile) => path.resolve(name) === path.resolve(file.fileName)
        ? ts.createSourceFile(name, text, language, true)
        : original(name, language, onError, shouldCreateNewSourceFile);
    return ts.createProgram(program.getRootFileNames(), program.getCompilerOptions(), host);
}
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const program = compiler(tree, ts);
    const file = program.getSourceFile(path.join(tree, manifest[0].file));
    assert(file && !file.parseDiagnostics.length, 'cannot parse sys.ts');
    const digest = hash(file.text);
    assert([manifest[0].beforeSha256, manifest[0].afterSha256].includes(digest), 'target source moved: sys.ts');
    const done = digest === manifest[0].afterSha256;
    const checked = validateBindings(program, file, done);
    const edits = [];
    if (!done) {
        const newline = file.text.includes('\r\n') ? '\r\n' : '\n';
        const block = blockLines.join(newline) + newline;
        const start = file.text.indexOf(block);
        assert(start >= 0 && file.text.indexOf(block, start + 1) < 0, 'crypto optional-load block moved');
        edits.push({start, end: start + block.length, text: ''});
        const os = one(file.statements.filter(ts.isImportDeclaration).filter(n => n.getText(file) === 'import * as _os from "node:os";'), '93 os import');
        edits.push({start: os.end, end: os.end, text: newline + importText});
        for (const [before, after] of [
            ['createHash: _crypto ? createSHA256Hash : generateDjb2Hash', 'createHash: createSHA256Hash'],
            ['createSHA256Hash: _crypto ? createSHA256Hash : undefined', 'createSHA256Hash'],
        ]) {
            const property = one(nodes(ts, checked.nodeSystem, n => ts.isPropertyAssignment(n) && n.getText(file) === before), before);
            edits.push({start: property.getStart(file), end: property.end, text: after});
        }
        const nonNull = checked.receiver.parent;
        edits.push({start: nonNull.getStart(file), end: nonNull.end, text: local});
    }
    let text = file.text, boundary = text.length;
    for (const edit of edits.sort((a, b) => b.start - a.start)) {
        assert(edit.end <= boundary, 'overlapping crypto edits');
        text = text.slice(0, edit.start) + edit.text + text.slice(edit.end);
        boundary = edit.start;
    }
    assert.equal(hash(text), manifest[0].afterSha256, 'unreviewed crypto import change');
    assert.equal(ts.createSourceFile(file.fileName, text, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
    const planned = plannedProgram(program, file, text);
    const post = validateBindings(planned, planned.getSourceFile(file.fileName), true);
    assert.equal(fs.readFileSync(file.fileName, 'utf8'), file.text, 'source changed while planning');
    // Validate all original bindings, planned bindings and full bytes before the sole write.
    if (edits.length) fs.writeFileSync(file.fileName, text);
    return {files: edits.length ? 1 : 0, javascriptChanging: true, before: checked.evidence, after: post.evidence};
}
module.exports = {adapt, validateBindings, plannedProgram, blockLines, importText};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
