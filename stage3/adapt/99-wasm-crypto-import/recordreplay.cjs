'use strict';
// The compiler owns this optional absent member already. This declaration lets
// stock tsc read the same member without a cast; it promises no runtime value.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');
const importText = 'import type {} from "../recordreplay.js";';
const declaration = 'export {};\n\ndeclare global {\n    namespace NodeJS {\n        interface Process {\n            recordreplay?: unknown;\n        }\n    }\n}\n';
function declarationPath(file) { return path.resolve(path.dirname(file), '../recordreplay.d.ts'); }
function accesses(file, owner) {
    assert(ts.isFunctionDeclaration(owner) && owner.name?.text === 'getNodeSystem', 'recordreplay owner moved');
    const stores = [], allReads = [];
    function visit(node) {
        if (ts.isVariableDeclaration(node) && ts.isIdentifier(node.name) && node.name.text === 'nodeSystem'
            && node.initializer && ts.isObjectLiteralExpression(node.initializer)) stores.push(node.initializer);
        if (ts.isPropertyAccessExpression(node) && node.name.text === 'recordreplay') allReads.push(node);
        ts.forEachChild(node, visit);
    }
    visit(owner);
    assert.equal(stores.length, 1, 'recordreplay nodeSystem store count moved');
    const properties = stores[0].properties.filter(n => ts.isPropertyAssignment(n)
        && ts.isIdentifier(n.name) && n.name.text === 'debugMode');
    assert.equal(properties.length, 1, 'recordreplay debugMode property count moved');
    const reads = [];
    function visitDebugMode(node) {
        if (ts.isPropertyAccessExpression(node) && node.name.text === 'recordreplay') reads.push(node);
        ts.forEachChild(node, visitDebugMode);
    }
    visitDebugMode(properties[0].initializer);
    assert.equal(reads.length, 1, 'recordreplay debugMode read count moved');
    assert.equal(allReads.length, 1, 'recordreplay read count moved');
    assert.equal(reads[0], allReads[0], 'recordreplay read must belong to nodeSystem.debugMode');
    return reads[0];
}
function sourceState(file, done) {
    const target = declarationPath(file.fileName);
    if (done) assert.equal(fs.readFileSync(target, 'utf8'), declaration, 'recordreplay declaration changed');
    else assert(!fs.existsSync(target), 'recordreplay declaration already exists before adaptation');
}
function edits(file, owner, os, newline, done) {
    sourceState(file, done);
    const imports = file.statements.filter(ts.isImportDeclaration).filter(n => n.moduleSpecifier.text === '../recordreplay.js');
    assert.equal(imports.length, done ? 1 : 0, 'recordreplay import count moved');
    const read = accesses(file, owner);
    if (done) {
        assert.equal(imports[0].getText(file), importText, 'recordreplay import moved');
        assert.equal(read.getText(file), 'process.recordreplay', 'recordreplay receiver changed');
        return [];
    }
    assert.equal(read.getText(file), '(process as any).recordreplay', 'recordreplay cast moved');
    assert.equal(file.text.slice(os.end, os.end + newline.length * 3), newline.repeat(3), 'reserved blank import line moved');
    return [
        {start: read.getStart(file), end: read.end, text: 'process.recordreplay'},
        {start: os.end + newline.length, end: os.end + newline.length * 2, text: importText + newline},
    ];
}
function importEdit(os, newline, edit) {
    return edit.start === os.end + newline.length && edit.end === os.end + newline.length * 2 && edit.text === importText + newline;
}
function virtualHost(host, file) {
    // Official apply clones upstream without node_modules. Resolve its explicit
    // Node type reference against the API's locked install, only in this virtual
    // validation program. No declaration or runtime import is added to the tree.
    const nodePackage = require.resolve('@types/node/package.json');
    assert.equal(require(nodePackage).version, require('../../api/package.json').dependencies['@types/node']);
    host.resolveTypeReferenceDirectives = (names, containingFile, redirectedReference, options, containingFileMode) =>
        names.map(name => ts.resolveTypeReferenceDirective(typeof name === 'string' ? name : name.fileName,
            containingFile, {...options, typeRoots: [path.dirname(path.dirname(nodePackage))]}, host,
            redirectedReference, undefined, containingFileMode).resolvedTypeReferenceDirective);
    const target = declarationPath(file);
    const exists = host.fileExists.bind(host), read = host.readFile.bind(host), source = host.getSourceFile.bind(host);
    host.fileExists = name => path.resolve(name) === target || exists(name);
    host.readFile = name => path.resolve(name) === target ? declaration : read(name);
    host.getSourceFile = (name, language, onError, fresh) => path.resolve(name) === target
        ? ts.createSourceFile(name, declaration, language, true)
        : source(name, language, onError, fresh);
}
function validate(program, file, owner) {
    const imported = file.statements.filter(ts.isImportDeclaration).filter(n => n.moduleSpecifier.text === '../recordreplay.js');
    assert.equal(imported.length, 1, 'planned recordreplay import missing');
    assert.equal(imported[0].getText(file), importText);
    assert(imported[0].importClause.isTypeOnly, 'recordreplay import must erase');
    const read = accesses(file, owner);
    assert(ts.isIdentifier(read.expression) && read.expression.text === 'process', 'recordreplay must be a direct process read');
    const checker = program.getTypeChecker();
    const symbol = checker.getSymbolAtLocation(read.expression);
    assert(symbol, 'process read has no binding');
    const nodeRoot = path.dirname(require.resolve('@types/node/package.json'));
    assert(symbol.declarations?.some(node => {
        const relative = path.relative(nodeRoot, node.getSourceFile().fileName);
        return !relative.startsWith('..') && !path.isAbsolute(relative) && node.getSourceFile().isDeclarationFile
            && ts.isVariableDeclaration(node) && ts.isIdentifier(node.name) && node.name.text === 'process'
            && node.type && ts.isTypeReferenceNode(node.type) && ts.isQualifiedName(node.type.typeName)
            && node.type.typeName.left.getText() === 'NodeJS' && node.type.typeName.right.text === 'Process';
    }), 'recordreplay must bind to the pinned ambient NodeJS.Process');
    assert.equal(checker.getTypeAtLocation(read).flags, ts.TypeFlags.Unknown, 'recordreplay must retain unknown');
    const declarations = program.getSourceFiles().filter(f => path.resolve(f.fileName) === declarationPath(file.fileName));
    assert.equal(declarations.length, 1, 'recordreplay declaration was not loaded');
    assert.equal(declarations[0].text, declaration, 'virtual recordreplay declaration changed');
    return {typeOnlyImport: true, directRead: true, type: 'unknown', declaration: 'src/recordreplay.d.ts', javascriptChanging: false};
}
function write(file, done) {
    if (!done) fs.writeFileSync(declarationPath(file.fileName), declaration, {flag: 'wx'});
}
module.exports = {accesses, edits, importEdit, virtualHost, validate, write, declaration, importText};
