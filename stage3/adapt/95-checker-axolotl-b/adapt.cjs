#!/usr/bin/env node
'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const crypto = require('node:crypto'), ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const {compiler, nodes} = require('../47-host-errors/adapt.cjs');
const manifest = require('./sites.json');
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
function enclosing(node) {
    while (node && !ts.isFunctionDeclaration(node)) node = node.parent;
    return node?.name?.text;
}
function typeShape(checker, type) {
    if (type.isUnion()) return 'union:' + type.types.map(t => typeShape(checker,t)).sort().join('|');
    if (checker.isTupleType(type)) return 'tuple:' + checker.getTypeArguments(type).map(t => typeShape(checker,t)).join(',');
    if (checker.isArrayType(type)) return 'array:' + checker.getTypeArguments(type).map(t => typeShape(checker,t)).join(',');
    const calls = type.getCallSignatures();
    if (calls.length) {
        assert.equal(calls.length,1,'read callback overloads moved');
        assert.equal(calls[0].parameters.length,0,'read callback parameters moved');
        return 'callback:' + typeShape(checker,checker.getReturnTypeOfSignature(calls[0]));
    }
    return type.symbol ? 'symbol:' + type.symbol.name : checker.typeToString(type);
}
function adapt(tree, check = false) {
    assert.equal(ts.version, '6.0.3');
    const program = compiler(tree, ts), checker = program.getTypeChecker();
    const source = program.getSourceFile(path.join(tree, 'src/compiler/checker.ts'));
    assert(source && !source.parseDiagnostics.length, 'checker must parse');
    const oneFunction = name => {
        const found = nodes(ts, source, node => ts.isFunctionDeclaration(node) && node.name?.text === name && !!node.body);
        assert.equal(found.length, 1, 'function inventory: ' + name);
        return found[0];
    };
    for (const dependency of manifest.dependencies) assert.equal(hash(oneFunction(dependency.name).getText(source)), dependency.sha256, 'caller/arity dependency moved: ' + dependency.name);
    for (const dependency of manifest.variables) {
        const found = nodes(ts, source, node => ts.isVariableDeclaration(node) && node.name.getText(source) === dependency.name);
        assert.equal(found.length, 1, 'resolver/fallback inventory');
        assert.equal(hash(found[0].getText(source)), dependency.sha256, 'resolver/fallback moved: ' + dependency.name);
    }
    const mapper = oneFunction('makeDeferredTypeMapper'), mapperSymbol = checker.getSymbolAtLocation(mapper.name);
    const references = nodes(ts, source, n => ts.isIdentifier(n) && checker.getSymbolAtLocation(n) === mapperSymbol);
    assert.equal(references.length, 4, 'deferred mapper must have only its definition and three direct callers');
    const calls = nodes(ts, source, n => ts.isCallExpression(n) && checker.getSymbolAtLocation(n.expression) === mapperSymbol);
    assert.deepEqual(calls.map(n => ({owner: enclosing(n), text: n.getText(source)})), manifest.mapperCallers, 'deferred mapper callers moved');
    for (const contract of manifest.referenceCalls) {
        const declaration = oneFunction(contract.name), symbol = checker.getSymbolAtLocation(declaration.name);
        assert.equal(nodes(ts,source,n => ts.isIdentifier(n) && checker.getSymbolAtLocation(n) === symbol).length,
            contract.references, 'reference/alias inventory: ' + contract.name);
        const calls = nodes(ts,source,n => ts.isCallExpression(n) && checker.getSymbolAtLocation(n.expression) === symbol);
        assert.deepEqual(calls.map(n => ({owner: enclosing(n),text: n.getText(source)})),contract.calls,'caller inventory: ' + contract.name);
    }
    const edits = [];
    for (const contract of manifest.owners) {
        const owner = oneFunction(contract.name), ownerText = owner.getText(source);
        assert(!owner.modifiers?.length, 'no public function edit');
        assert([contract.beforeSha256, contract.afterSha256].includes(hash(ownerText)), 'uninventoried owner: ' + contract.name);
        const planned = [];
        for (const site of manifest.sites.filter(row => row.owner === contract.name)) {
            const found = nodes(ts, owner, node => {
                if (node.getText(source) !== site.expression) return false;
                if (ts.isElementAccessExpression(node)) return true;
                const parent = ts.isNonNullExpression(node.parent) ? node.parent.parent : node.parent;
                return ts.isIdentifier(node) && ts.isCallExpression(parent) &&
                    (parent.arguments[0] === node || ts.isNonNullExpression(parent.arguments[0]) && parent.arguments[0].expression === node) &&
                    parent.expression.getText(source) === 'resolver.resolveIterationType';
            });
            assert.equal(found.length, site.total, 'read occurrence count: ' + site.expression);
            const node = found[site.occurrence - 1];
            const bindingNode = ts.isIdentifier(node) ? node : ts.isIdentifier(node.expression) ? node.expression : node.expression.expression;
            const binding = checker.getSymbolAtLocation(bindingNode);
            assert(binding && binding.name === site.binding && enclosing(binding.valueDeclaration) === site.bindingOwner, 'read binding moved');
            assert.equal(typeShape(checker,checker.getTypeAtLocation(node)), site.shape, 'read type moved');
            const asserted = ts.isNonNullExpression(node.parent);
            if (!asserted) {
                assert(!check, 'required assertion missing: ' + site.owner + ':' + site.occurrence);
                planned.push({at: node.end, text: '!'});
            }
        }
        let expected = ownerText;
        for (const edit of [...planned].sort((a,b) => b.at - a.at)) {
            const at = edit.at - owner.getStart(source);
            expected = expected.slice(0, at) + edit.text + expected.slice(at);
        }
        assert.equal(hash(expected), contract.afterSha256, 'only inventoried owner edits');
        edits.push(...planned);
    }
    let after = source.text;
    for (const edit of edits.sort((a,b) => b.at - a.at)) after = after.slice(0,edit.at) + edit.text + after.slice(edit.at);
    assert.equal(ts.createSourceFile(source.fileName, after, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
    for (const removeComments of [false, true]) {
        const options = {fileName: source.fileName, compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}};
        assert.equal(ts.transpileModule(after,options).outputText === ts.transpileModule(source.text,options).outputText, true, 'JavaScript changed');
    }
    assert.equal(fs.readFileSync(source.fileName,'utf8'), source.text, 'source changed while planning');
    if (after !== source.text && !check) fs.writeFileSync(source.fileName,after);
    return {files: Number(after !== source.text), assertions: edits.length, owners: manifest.owners.length, javascriptIdentical: true};
}
module.exports = {adapt, typeShape};
if (require.main === module) {
    const check = process.argv[2] === '--check';
    assert.equal(process.argv.length, check ? 4 : 3, 'usage: node adapt.cjs [--check] TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[check ? 3 : 2]),check),null,2));
}
