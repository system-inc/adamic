'use strict';
// Validate the bare read through CLI's Node-prelude augmentation, without
// adding a declaration or import to the adapted tree.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');
// Read CLI's augmentation verbatim; validation must follow the compiler's prelude.
const loader = fs.readFileSync(path.resolve(__dirname, '../../../internal/load/node_library.go'), 'utf8');
const augmentation = JSON.parse(loader.match(/^const nodeProcessAugmentation = (".*")$/m)[1]);
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
function edits(file, owner, os, newline, done) {
    assert(!fs.existsSync(path.resolve(path.dirname(file.fileName), '../recordreplay.d.ts')), 'adaptation declaration must be absent');
    assert(!file.statements.some(n => ts.isImportDeclaration(n) && n.moduleSpecifier.text === '../recordreplay.js'), 'adaptation import must be absent');
    const read = accesses(file, owner);
    assert.equal(read.getText(file), done ? 'process.recordreplay' : '(process as any).recordreplay', 'recordreplay receiver changed');
    return done ? [] : [{start: read.getStart(file), end: read.end, text: 'process.recordreplay'}];
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
    const source = host.getSourceFile.bind(host);
    host.getSourceFile = (name, language, onError, fresh) => {
        const result = source(name, language, onError, fresh);
        if (result && path.resolve(name) === path.join(path.dirname(nodePackage), 'globals.d.ts'))
            return ts.createSourceFile(name, result.text + augmentation, language, true);
        return result;
    };
}

function validate(program, file, owner) {
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
    const member = checker.getPropertyOfType(checker.getTypeAtLocation(read.expression), 'recordreplay');
    assert(member?.declarations?.some(d => d.getSourceFile().text.endsWith(augmentation)), 'read must type through CLI Node-prelude augmentation');
    return {directRead: true, type: 'unknown', nodePreludeAugmentation: true, adaptationDeclaration: false};
}
module.exports = {accesses, edits, virtualHost, validate, augmentation};
