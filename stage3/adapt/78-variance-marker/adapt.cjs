'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const {compiler, nodes} = require('../47-host-errors/adapt.cjs');
assert.equal(ts.version, '6.0.3');
function plan(program, tree) {
    const checker = program.getTypeChecker();
    const file = program.getSourceFile(path.join(tree, 'src/compiler/checker.ts'));
    assert(file && file.parseDiagnostics.length === 0, 'checker source');
    function one(root, predicate, label) {
        const found = nodes(ts, root, predicate);
        assert.equal(found.length, 1, label);
        return found[0];
    }
    const marker = one(file, n => ts.isVariableDeclaration(n) && n.name.getText(file) === 'outofbandVarianceMarkerHandler', 'marker declaration');
    const symbol = checker.getSymbolAtLocation(marker.name);
    assert(symbol, 'unresolved marker');
    const references = nodes(ts, file, n => ts.isIdentifier(n) && n.text === 'outofbandVarianceMarkerHandler' && checker.getSymbolAtLocation(n) === symbol);
    assert.equal(references.length, 15, 'marker reference drift');
    const writers = references.filter(n => ts.isBinaryExpression(n.parent) && n.parent.left === n && n.parent.operatorToken.kind === ts.SyntaxKind.EqualsToken);
    assert.deepEqual(writers.map(n => n.parent.right.getText(file).replace(/\r\n/g, '\n')), [
        'onlyUnreliable => {\n                    propagatingVarianceFlags |= onlyUnreliable ? RelationComparisonResult.ReportsUnreliable : RelationComparisonResult.ReportsUnmeasurable;\n                    return originalHandler!(onlyUnreliable);\n                }',
        'originalHandler',
        'onlyUnreliable => onlyUnreliable ? unreliable = true : unmeasurable = true',
        'oldHandler',
    ], 'marker writers drift');
    const calls = references.filter(n => ts.isCallExpression(n.parent) && n.parent.expression === n);
    assert.equal(calls.length, 2, 'marker call drift');
    for (const call of calls) assert(ts.isExpressionStatement(call.parent.parent), 'marker result is no longer discarded');
    const callback = one(marker.type, ts.isFunctionTypeNode, 'marker callback type');
    assert.equal(callback.parameters.length, 1);
    assert.equal(callback.parameters[0].type.kind, ts.SyntaxKind.BooleanKeyword);
    const signature = checker.getSignatureFromDeclaration(callback);
    assert(signature, 'unresolved callback signature');
    const before = callback.type.getText(file);
    assert(['void', 'boolean'].includes(before), 'marker return type drift');
    assert.equal(checker.typeToString(checker.getReturnTypeOfSignature(signature)), before);
    const start = callback.type.getStart(file);
    const changed = file.text.slice(0, start) + 'boolean' + file.text.slice(callback.type.end);
    for (const removeComments of [false, true]) {
        const compilerOptions = {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments};
        assert.equal(ts.transpileModule(changed, {fileName: file.fileName, compilerOptions}).outputText,
            ts.transpileModule(file.text, {fileName: file.fileName, compilerOptions}).outputText, 'checker runtime JavaScript changed');
    }
    return {file: file.fileName, original: file.text, changed, evidence: {
        file: 'src/compiler/checker.ts', line: file.getLineAndCharacterOfPosition(start).line + 1,
        column: file.getLineAndCharacterOfPosition(start).character + 1, before, after: 'boolean',
        references: references.map(n => ({line: file.getLineAndCharacterOfPosition(n.getStart(file)).line + 1, parent: ts.SyntaxKind[n.parent.kind]})),
        javascriptIdentical: true,
    }};
}
function adapt(tree) {
    const result = plan(compiler(tree, ts), tree);
    const files = Number(result.changed !== result.original);
    if (files) fs.writeFileSync(result.file, result.changed);
    return {files, ...result.evidence};
}
module.exports = {adapt, plan};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
