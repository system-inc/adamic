#!/usr/bin/env node
'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const crypto = require('node:crypto'), ts = require('typescript');
const {compiler, nodes} = require('../47-host-errors/adapt.cjs');
const inventory = require('./sites.json');
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
const insertion = '\r\n                if (relevantTypeParameter === undefined) return undefined;';
function adapt(tree, check = false) {
    assert.equal(ts.version, '6.0.3');
    const program = compiler(tree, ts), checker = program.getTypeChecker();
    const name = path.join(tree, inventory.file), before = fs.readFileSync(name, 'utf8'), source = program.getSourceFile(name);
    assert(source && !source.parseDiagnostics.length);
    const one = (root, predicate) => { const matches = nodes(ts, root, predicate); assert.equal(matches.length, 1, 'inventoried parsed address'); return matches[0]; };
    const owner = one(source, n => ts.isFunctionDeclaration(n) && n.name?.text === inventory.owner);
    const body = owner.getText(source), fixed = body.includes(insertion);
    assert.equal(hash(body.replace(insertion, '')), inventory.ownerBeforeSha256, 'owner shape moved');
    assert(!owner.modifiers?.length && owner.parent.parent.name?.text === 'createTypeChecker', 'private nested owner');
    const declaration = one(owner, n => ts.isVariableDeclaration(n) && n.name.getText(source) === 'relevantTypeParameter');
    assert.equal(declaration.getText(source), 'relevantTypeParameter = typeParameters[typeArgumentPosition]');
    const read = declaration.initializer;
    assert(ts.isElementAccessExpression(read));
    const array = checker.getSymbolAtLocation(read.expression), index = checker.getSymbolAtLocation(read.argumentExpression);
    assert(array && index && array.declarations.length === 1 && index.declarations.length === 1);
    assert(owner.pos < array.declarations[0].pos && array.declarations[0].end < owner.end);
    assert(owner.pos < index.declarations[0].pos && index.declarations[0].end < owner.end);
    assert.equal(checker.typeToString(checker.getTypeAtLocation(read)), 'TypeParameter | undefined', 'indexed type');
    const symbol = checker.getSymbolAtLocation(declaration.name);
    const call = one(owner, n => ts.isCallExpression(n) && n.expression.getText(source) === 'getConstraintOfTypeParameter');
    assert.equal(call.arguments.length, 1);
    assert.equal(checker.getSymbolAtLocation(call.arguments[0]), symbol, 'argument binding');
    const target = checker.getResolvedSignature(call).declaration;
    assert(ts.isFunctionDeclaration(target) && target.name.text === 'getConstraintOfTypeParameter' && target.getSourceFile() === source);
    assert.equal(checker.typeToString(checker.getTypeAtLocation(call.arguments[0])), fixed ? 'TypeParameter' : 'TypeParameter | undefined', 'guard narrowing');
    const statement = declaration.parent.parent;
    assert(ts.isVariableStatement(statement));
    const statements = statement.parent.statements, position = statements.indexOf(statement);
    if (fixed) assert.equal(statements[position + 1].getText(source), insertion.trim(), 'absence guard position');
    assert.equal(statements[position + (fixed ? 2 : 1)].getText(source), 'const constraint = getConstraintOfTypeParameter(relevantTypeParameter);', 'next read');
    if (check) assert(fixed, 'excess-argument guard missing');
    const changed = fixed ? before : before.slice(0, statement.end) + insertion + before.slice(statement.end);
    assert.equal(ts.createSourceFile(name, changed, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
    assert.equal(fs.readFileSync(name, 'utf8'), before, 'source changed while planning');
    if (changed !== before && !check) fs.writeFileSync(name, changed);
    return {files: Number(changed !== before), edits: Number(changed !== before), javascriptChanging: true};
}
module.exports = {adapt};
if (require.main === module) {
    const check = process.argv[2] === '--check';
    assert.equal(process.argv.length, check ? 4 : 3);
    console.log(JSON.stringify(adapt(path.resolve(process.argv[check ? 3 : 2]), check), null, 2));
}
