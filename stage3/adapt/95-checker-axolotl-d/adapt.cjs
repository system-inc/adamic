#!/usr/bin/env node
'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const crypto = require('node:crypto'), ts = require('typescript');
const {compiler, nodes} = require('../47-host-errors/adapt.cjs');
const ownerBeforeSha256 = '57faba456412535bb02760aad9cdbcf7569cc83624729da3f6ddf40b6ac5a80c';
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
function adapt(tree, check = false) {
    assert.equal(ts.version, '6.0.3');
    const program = compiler(tree, ts), checker = program.getTypeChecker();
    const name = path.join(tree, 'src/compiler/checker.ts'), before = fs.readFileSync(name, 'utf8');
    const source = program.getSourceFile(name);
    assert(source && !source.parseDiagnostics.length);
    const one = (root, predicate) => { const found = nodes(ts, root, predicate); assert.equal(found.length, 1, 'parsed address count'); return found[0]; };
    const owner = one(source, n => ts.isFunctionDeclaration(n) && n.name?.text === 'getFlowTypeFromCommonJSExport');
    assert.equal(hash(owner.getText(source).replace('symbol.declarations![0]!', 'symbol.declarations![0]')), ownerBeforeSha256, 'owner shape moved');
    assert(!owner.modifiers?.length && owner.parent.parent.name?.text === 'createTypeChecker', 'private nested owner');
    const parameter = owner.parameters[0], symbol = checker.getSymbolAtLocation(parameter.name);
    assert.equal(checker.typeToString(checker.getTypeAtLocation(parameter.name)), 'Symbol');
    const read = one(owner, n => ts.isElementAccessExpression(n) && n.getText(source) === 'symbol.declarations![0]');
    assert(ts.isNonNullExpression(read.expression) && ts.isPropertyAccessExpression(read.expression.expression));
    assert.equal(checker.getSymbolAtLocation(read.expression.expression.expression), symbol, 'parameter binding');
    assert.equal(checker.typeToString(checker.getTypeAtLocation(read)), 'Declaration | undefined', 'indexed read type');
    const asserted = ts.isNonNullExpression(read.parent), argument = asserted ? read.parent : read;
    const call = argument.parent;
    assert(ts.isCallExpression(call) && call.arguments.length === 1 && call.arguments[0] === argument);
    let callee = checker.getSymbolAtLocation(call.expression);
    if (callee?.flags & ts.SymbolFlags.Alias) callee = checker.getAliasedSymbol(callee);
    assert.equal(callee?.name, 'getSourceFileOfNode');
    assert(callee.declarations.every(d => d.getSourceFile().fileName.endsWith('/src/compiler/utilities.ts')), 'producer resolution');
    assert.equal(checker.typeToString(checker.getTypeAtLocation(call)), asserted ? 'SourceFile' : 'SourceFile | undefined', 'producer overload');
    const ownerSymbol = checker.getSymbolAtLocation(owner.name);
    const callers = nodes(ts, source, n => ts.isCallExpression(n) && checker.getSymbolAtLocation(n.expression) === ownerSymbol);
    assert.equal(callers.length, 1, 'all callers inventoried');
    const caller = callers[0];
    assert.equal(caller.getText(source), 'getFlowTypeFromCommonJSExport(exportSymbol)');
    assert(ts.isConditionalExpression(caller.parent) && caller.parent.whenTrue === caller);
    assert.equal(caller.parent.condition.getText(source), 'exportSymbol?.declarations && isDuplicatedCommonJSExport(exportSymbol.declarations) && symbol.declarations!.length');
    const duplicate = one(source, n => ts.isFunctionDeclaration(n) && n.name?.text === 'isDuplicatedCommonJSExport');
    assert.equal(duplicate.body.statements[0].expression.getText(source), 'declarations\r\n            && declarations.length > 1\r\n            && declarations.every(d => isInJSFile(d) && isAccessExpression(d) && (isExportsIdentifier(d.expression) || isModuleExportsAccessExpression(d.expression)))');
    if (check) assert(asserted, 'required declaration read missing');
    const changed = asserted ? before : before.slice(0, read.end) + '!' + before.slice(read.end);
    assert.equal(ts.createSourceFile(name, changed, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
    for (const removeComments of [false, true]) {
        const options = {fileName: name, compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}};
        assert.equal(ts.transpileModule(before, options).outputText, ts.transpileModule(changed, options).outputText, 'JavaScript changed');
    }
    assert.equal(fs.readFileSync(name, 'utf8'), before, 'source changed while planning');
    if (changed !== before && !check) fs.writeFileSync(name, changed);
    return {files: Number(before !== changed), assertions: Number(before !== changed), ownerBeforeSha256, javascriptIdentical: true};
}
module.exports = {adapt};
if (require.main === module) {
    const check = process.argv[2] === '--check';
    assert.equal(process.argv.length, check ? 4 : 3);
    console.log(JSON.stringify(adapt(path.resolve(process.argv[check ? 3 : 2]), check), null, 2));
}
