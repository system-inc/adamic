#!/usr/bin/env node
'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');

assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
const file = path.join(process.argv[2], 'src/compiler/checker.ts');
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
assert.equal(source.parseDiagnostics.length, 0);

function nodes(root, predicate) {
    const result = [];
    function visit(node) {
        if (predicate(node)) result.push(node);
        ts.forEachChild(node, visit);
    }
    visit(root);
    return result;
}
function one(items, description) {
    assert.equal(items.length, 1, description);
    return items[0];
}
const name = 'resolvedImplementsTypes';
const owner = one(nodes(source, node => ts.isFunctionDeclaration(node) &&
    node.name?.text === 'getImplementsTypes'), 'expected one getImplementsTypes');
assert(owner.body && owner.type?.getText(source) === 'readonly BaseType[]',
    'getImplementsTypes return or body drift');
const binding = one(nodes(owner.body, node => ts.isVariableDeclaration(node) &&
    ts.isIdentifier(node.name) && node.name.text === name), 'expected one local binding');
assert(ts.isVariableDeclarationList(binding.parent) &&
    (binding.parent.flags & ts.NodeFlags.BlockScoped) === ts.NodeFlags.Let &&
    binding.initializer?.getText(source) === 'emptyArray', 'local declaration drift');
const branch = one(nodes(owner.body, node => ts.isIfStatement(node) &&
    node.expression.getText(source) === name + ' === emptyArray'), 'expected one sentinel guard');
assert(ts.isBlock(branch.thenStatement) && branch.thenStatement.statements.length === 1 &&
    branch.thenStatement.statements[0].getText(source) === name + ' = [implementsType as ObjectType];',
    'equal branch drift');
assert(branch.elseStatement && ts.isBlock(branch.elseStatement) &&
    branch.elseStatement.statements.length === 1, 'unequal branch drift');
const writer = branch.elseStatement.statements[0];
const before = [name + '.push(implementsType);', 'BaseType[]'];
const after = [name + ' = [...' + name + ', implementsType];', 'readonly BaseType[]'];
const targets = [writer, binding.type];
assert(binding.type && targets[0].getStart(source) > targets[1].getStart(source), 'missing or reordered targets');
const actual = targets.map(node => node.getText(source));
if (actual.every((value, index) => value === after[index])) {
    console.log('getImplementsTypes.resolvedImplementsTypes: already adapted');
} else {
    assert.deepEqual(actual, before, 'local type or guarded writer drift');
    let updated = text;
    for (const [index, node] of targets.entries()) {
        updated = updated.slice(0, node.getStart(source)) + after[index] + updated.slice(node.end);
    }
    assert.equal(fs.readFileSync(file, 'utf8'), text, 'checker.ts changed during adaptation');
    fs.writeFileSync(file, updated);
    console.log('getImplementsTypes.resolvedImplementsTypes: adapted local type and unequal branch');
}
