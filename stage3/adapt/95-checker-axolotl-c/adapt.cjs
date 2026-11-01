#!/usr/bin/env node
'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const crypto = require('node:crypto'), ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const {compiler, nodes} = require('../47-host-errors/adapt.cjs');
const inventory = require('./sites.json');
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
function adapt(tree, check = false) {
    assert.equal(ts.version, '6.0.3');
    const program = compiler(tree, ts), checker = program.getTypeChecker(), plans = [], evidence = [];
    const checkerFile = program.getSourceFile(path.join(tree, 'src/compiler/checker.ts'));
    const uniqueFunction = (name, body = true) => {
        const found = nodes(ts, checkerFile, n => ts.isFunctionDeclaration(n) && n.name?.text === name && (!body || n.body));
        assert.equal(found.length, 1, 'implementation count: ' + name);
        return found[0];
    };
    const labelImplementation = uniqueFunction('getTupleElementLabel');
    assert.equal(labelImplementation.parameters[2].getText(checkerFile), 'elementFlags = ElementFlags.Fixed', 'existing default');
    assert(!labelImplementation.modifiers?.length && labelImplementation.parent.parent.name?.text === 'createTypeChecker');

    for (const relative of [...new Set(inventory.map(row => row.file))]) {
        const file = program.getSourceFile(path.join(tree, relative));
        assert(file && !file.parseDiagnostics.length, 'source must parse');
        const edits = [];
        for (const row of inventory.filter(row => row.file === relative)) {
            const owners = nodes(ts, file, n => (ts.isFunctionDeclaration(n) || ts.isInterfaceDeclaration(n)) && n.name?.text === row.owner);
            assert.equal(owners.length, row.ownerCount, 'owner count: ' + row.owner);
            const owner = owners[row.ownerOccurrence];
            const normalized = owner.getText(file).replace(row.after, row.before);
            assert.equal(hash(normalized), row.ownerBeforeSha256, 'owner shape: ' + row.owner);
            if (ts.isFunctionDeclaration(owner)) {
                assert(!owner.modifiers?.length, 'file-private function');
                if (row.nested) assert.equal(owner.parent.parent.name?.text, 'createTypeChecker');
            } else assert(ts.getJSDocTags(owner).some(tag => tag.tagName.text === 'internal'), 'internal interface');
            const candidates = nodes(ts, owner, n => [row.kind, row.kindAfter].includes(ts.SyntaxKind[n.kind]) && n.name?.getText(file) === row.name);
            assert.equal(candidates.length, 1, 'parsed declaration count');
            const node = candidates[0], text = node.getText(file);
            assert([row.before, row.after].includes(text), 'declaration shape');
            const symbol = checker.getSymbolAtLocation(node.name);
            assert(symbol?.declarations?.includes(node), 'declaration symbol');
            assert.equal(checker.typeToString(checker.getTypeAtLocation(node.name)), text === row.before ? row.typeBefore : row.typeAfter, 'declaration type');
            if (row.name === 'tupleLabelDeclaration') {
                const producer = uniqueFunction('createTupleTargetType');
                const writes = nodes(ts, producer, n => ts.isPropertyAccessExpression(n) && n.name.text === row.name);
                assert.equal(writes.length, 1);
                assert.equal(checker.getSymbolAtLocation(writes[0].name), symbol, 'tuple slot writer binding');
                assert.equal(writes[0].parent.getText(checkerFile), 'property.links.tupleLabelDeclaration = namedMemberDeclarations?.[i]', 'intentional absent label');
                assert(checker.getTypeAtLocation(writes[0].parent.right).isUnion());
            }
            if (row.name === 'readFile') {
                const factory = uniqueFunction('createBasicNodeBuilderModuleSpecifierResolutionHost');
                const target = checker.getReturnTypeOfSignature(checker.getSignatureFromDeclaration(factory));
                assert.equal(checker.getPropertyOfType(target, 'readFile'), symbol, 'returned host slot binding');
                const property = nodes(ts, factory, n => ts.isPropertyAssignment(n) && n.name.getText(checkerFile) === 'readFile');
                assert.equal(property.length, 1);
                assert.equal(property[0].initializer.getText(checkerFile), 'host.readFile ? (fileName => host.readFile!(fileName)) : undefined', 'intentional absent callback');
            }
            if (row.owner === 'getMergedSymbol') {
                const reads = nodes(ts, owner, n => ts.isIdentifier(n) && checker.getSymbolAtLocation(n) === symbol);
                assert.equal(reads.length, 3, 'merged declaration, assignment, result');
            }
            if (text === row.before) {
                assert(!check, 'missing declaration repair: ' + row.name);
                edits.push({start: node.getStart(file), end: node.end, text: row.after});
            }
            evidence.push({owner: row.owner, name: row.name, edited: text === row.before});
        }
        let changed = file.text, boundary = changed.length;
        for (const edit of edits.sort((a,b) => b.start-a.start)) {
            assert(edit.end <= boundary, 'overlapping edits');
            changed = changed.slice(0,edit.start) + edit.text + changed.slice(edit.end);
            boundary = edit.start;
        }
        assert.equal(ts.createSourceFile(file.fileName, changed, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
        for (const removeComments of [false,true]) {
            const options = {fileName: file.fileName, compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}};
            assert.equal(ts.transpileModule(file.text,options).outputText, ts.transpileModule(changed,options).outputText, 'JavaScript changed');
        }
        plans.push({file, changed});
    }
    // Every declaration and output is validated before the first write.
    for (const {file} of plans) assert.equal(fs.readFileSync(file.fileName,'utf8'),file.text,'source changed while planning');
    for (const {file,changed} of plans) if (!check && changed !== file.text) fs.writeFileSync(file.fileName,changed);
    return {files: plans.filter(p => p.changed !== p.file.text).length, declarations: evidence.filter(e=>e.edited).length, javascriptIdentical: true, evidence};
}
module.exports = {adapt};
if (require.main === module) {
    const check = process.argv[2] === '--check';
    assert.equal(process.argv.length,check ? 4 : 3,'usage: adapt.cjs [--check] TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[check ? 3 : 2]),check),null,2));
}
