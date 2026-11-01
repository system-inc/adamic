#!/usr/bin/env node
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const inventory = require('./sites.json');
const {compiler, nodes} = require('../47-host-errors/adapt.cjs');
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
function adapt(tree, check = false) {
    assert.equal(ts.version, '6.0.3');
    const name = path.join(tree, inventory.file), before = fs.readFileSync(name, 'utf8');
    assert([inventory.beforeSha256, inventory.afterSha256].includes(hash(before)), 'uninventoried checker source');
    const program = compiler(tree, ts), checker = program.getTypeChecker();
    const source = program.getSourceFile(name);
    assert(source && !source.parseDiagnostics.length, 'checker must parse');
    const owners = nodes(ts, source, n => ts.isFunctionDeclaration(n) && n.name?.text === inventory.owner);
    assert.equal(owners.length, 1, 'owner count');
    const owner = owners[0];
    assert.equal(hash(owner.getText(source).replaceAll('parameter!', 'parameter')), inventory.ownerBeforeSha256, 'owner shape moved');
    assert(!owner.modifiers?.length, 'private function must have no modifiers');
    assert.equal(owner.parent.parent.name?.text, 'createTypeChecker', 'private nested owner');
    const statements = owner.body.statements;
    assert.equal(statements[0].getText(source), 'const parameter = node.parameters[0];', 'absence-handling initializer');
    const guard = statements[1];
    assert(ts.isIfStatement(guard));
    assert.equal(guard.expression.getText(source), 'node.parameters.length !== 1');
    assert.equal(nodes(ts, guard, ts.isReturnStatement).length, 2, 'both invalid-arity arms return');
    const declaration = statements[0].declarationList.declarations[0];
    const symbol = checker.getSymbolAtLocation(declaration.name);
    assert(symbol, 'parameter symbol');
    const reads = statements.slice(2).flatMap(statement => nodes(ts, statement,
        n => ts.isIdentifier(n) && checker.getSymbolAtLocation(n) === symbol));
    assert.equal(reads.length, inventory.sites.length, 'read count');
    let changed = before, edits = 0;
    for (const [index, node] of [...reads.entries()].reverse()) {
        const site = inventory.sites[index];
        assert.equal(site.occurrence, index + 1);
        assert.equal(site.total, reads.length);
        const asserted = ts.isNonNullExpression(node.parent);
        const parent = asserted ? node.parent.parent : node.parent;
        assert.equal(parent.getText(source).replaceAll('parameter!', 'parameter'), site.parent, 'read context');
        const type = checker.getTypeAtLocation(node);
        assert(type.isUnion() && type.types.length === 2 && type.types.some(t => t.flags === ts.TypeFlags.Undefined) &&
            type.types.some(t => t.symbol?.name === 'ParameterDeclaration'), 'parameter type moved');
        if (!asserted) {
            assert(!check, 'required assertion missing: occurrence ' + site.occurrence);
            changed = changed.slice(0, node.end) + '!' + changed.slice(node.end);
            edits++;
        }
    }
    assert.equal(hash(changed), inventory.afterSha256, 'only inventoried reads may change');
    assert.equal(ts.createSourceFile(name, changed, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
    for (const removeComments of [false, true]) {
        const options = {fileName: name, compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}};
        assert.equal(ts.transpileModule(before, options).outputText, ts.transpileModule(changed, options).outputText, 'JavaScript changed');
    }
    assert.equal(fs.readFileSync(name, 'utf8'), before, 'source changed while planning');
    if (changed !== before && !check) fs.writeFileSync(name, changed);
    return {files: Number(changed !== before), assertions: edits, javascriptIdentical: true};
}
module.exports = {adapt};
if (require.main === module) {
    const check = process.argv[2] === '--check';
    assert.equal(process.argv.length, check ? 4 : 3, 'usage: node adapt.cjs [--check] TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[check ? 3 : 2]), check), null, 2));
}
