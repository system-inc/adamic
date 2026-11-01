#!/usr/bin/env node
'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict'), crypto = require('node:crypto'), ts = require('typescript');
const inventory = require('./sites.json');
const {compiler, nodes} = require('../47-host-errors/adapt.cjs');
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
function adapt(tree, check = false) {
    assert.equal(ts.version, '6.0.3');
    const program = compiler(tree, ts), checker = program.getTypeChecker(), name = path.join(tree, inventory.file);
    const before = fs.readFileSync(name, 'utf8'), source = program.getSourceFile(name);
    assert(source && !source.parseDiagnostics.length);
    const one = (root, predicate) => { const found = nodes(ts, root, predicate); assert.equal(found.length, 1, 'inventoried parsed address'); return found[0]; };
    const owner = one(source, n => ts.isFunctionDeclaration(n) && n.name?.text === inventory.owner);
    assert(!owner.modifiers?.length && owner.parent.parent.name?.text === 'createTypeChecker', 'private nested owner');
    const old = nodes(ts, owner, n => ts.isExpressionStatement(n) && n.getText(source) === inventory.before);
    const current = nodes(ts, owner, n => ts.isIfStatement(n) && n.getText(source) === inventory.after);
    assert.equal(old.length + current.length, 1, 'parent selection shape moved');
    const fixed = current.length === 1, target = fixed ? current[0] : old[0];
    assert.equal(hash(owner.getText(source).replace(inventory.after, inventory.before)), inventory.ownerBeforeSha256, 'owner shape moved');
    const parameter = checker.getSymbolAtLocation(owner.parameters[0].name);
    assert.equal(checker.typeToString(checker.getTypeAtLocation(owner.parameters[0].name)), 'Type');
    const copy = one(owner, n => ts.isVariableDeclaration(n) && n.name.getText(source) === 'copy');
    const symbol = checker.getSymbolAtLocation(copy.name);
    assert.equal(copy.initializer.getText(source), 'createSymbol(SymbolFlags.Signature, InternalSymbolName.Index)');
    assert.equal(checker.typeToString(checker.getTypeAtLocation(copy.name)), 'TransientSymbol');
    const read = one(target, n => ts.isElementAccessExpression(n) && n.getText(source) === 'copy.declarations[0]');
    assert.equal(checker.getSymbolAtLocation(read.expression.expression), symbol, 'copy binding');
    assert.equal(checker.typeToString(checker.getTypeAtLocation(read)), 'Declaration | undefined');
    const calls = nodes(ts, target, n => ts.isCallExpression(n) && n.expression.getText(source) === 'getSymbolAtLocation');
    assert.equal(calls.length, 1);
    const callee = checker.getResolvedSignature(calls[0]).declaration;
    assert(ts.isFunctionDeclaration(callee) && callee.name.text === 'getSymbolAtLocation' && callee.getSourceFile() === source);
    const conditions = fixed ? [target.expression, target.elseStatement.expression] : [target.expression.right.condition, target.expression.right.whenFalse.condition];
    for (const [index, condition] of conditions.entries()) {
        assert.equal(condition.getText(source), index ? 'type.symbol' : 'type.aliasSymbol');
        assert.equal(checker.getSymbolAtLocation(condition.expression), parameter, 'ordered type binding');
    }
    if (fixed) {
        const fallback = target.elseStatement.elseStatement;
        assert(ts.isBlock(fallback) && fallback.statements.length === 3);
        const declaration = fallback.statements[0].declarationList.declarations[0];
        assert.equal(declaration.name.getText(source), 'declaration');
        const argument = calls[0].arguments[0];
        assert.equal(checker.getSymbolAtLocation(argument.expression), checker.getSymbolAtLocation(declaration.name));
        assert.equal(checker.typeToString(checker.getTypeAtLocation(argument.expression)), 'Declaration', 'fallback guard narrowing');
    }
    if (check) assert(fixed, 'throwing-fallback guard missing');
    const changed = fixed ? before : before.slice(0, target.getStart(source)) + inventory.after + before.slice(target.end);
    assert.equal(ts.createSourceFile(name, changed, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
    assert.equal(fs.readFileSync(name, 'utf8'), before, 'source changed while planning');
    if (changed !== before && !check) fs.writeFileSync(name, changed);
    return {files: Number(changed !== before), edits: Number(changed !== before), javascriptChanging: true, scope: 'third parent-selection arm only'};
}
module.exports = {adapt};
if (require.main === module) {
    const check = process.argv[2] === '--check'; assert.equal(process.argv.length, check ? 4 : 3);
    console.log(JSON.stringify(adapt(path.resolve(process.argv[check ? 3 : 2]), check), null, 2));
}
