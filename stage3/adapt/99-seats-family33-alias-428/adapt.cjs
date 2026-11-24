#!/usr/bin/env node
'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
const file = path.join(process.argv[2], 'src/compiler/transformers/utilities.ts');
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
assert.equal(source.parseDiagnostics.length, 0, 'utilities.ts must parse');
const owners = source.statements.filter(node => ts.isClassDeclaration(node) && node.name?.text === 'IdentifierNameMap');
assert.equal(owners.length, 1, 'expected one IdentifierNameMap class');
const methods = owners[0].members.filter(node => ts.isMethodDeclaration(node) && node.name.getText(source) === 'toKey');
assert.equal(methods.length, 1, 'expected one IdentifierNameMap.toKey method');

const candidates = [];
const aliases = [];
function visit(node) {
    if (ts.isVariableDeclaration(node)) {
        if (node.name.getText(source) === 'baseName' && node.initializer && ts.isConditionalExpression(node.initializer)) {
            candidates.push(node);
        }
        if (node.name.getText(source) === 'nodeForId') aliases.push(node);
    }
    ts.forEachChild(node, visit);
}
visit(methods[0]);
assert.equal(candidates.length, 1, 'expected one conditional baseName declaration');

const statement = candidates[0].parent.parent;
assert(ts.isVariableStatement(statement) && ts.isBlock(statement.parent));
const statements = statement.parent.statements;
const index = statements.indexOf(statement);
const original = 'const baseName = isMemberName(node) && node !== name ? IdentifierNameMap.toKey(node) : `(generated@${getNodeId(node)})`;';
const adapted = original.replace('getNodeId(node)', 'getNodeId(nodeForId)');
const alias = 'const nodeForId = node;';
const initializer = 'const node = getNodeForGeneratedName(name);';

if (statement.getText(source) === adapted) {
    assert.equal(aliases.length, 1, 'expected one alias');
    assert.equal(statements[index - 1]?.getText(source), alias, 'alias snapshot changed');
    assert.equal(statements[index - 2]?.getText(source), initializer, 'initializer snapshot changed');
    console.log('family 33 alias at 428 already applied');
} else {
    assert.equal(aliases.length, 0, 'nodeForId already exists');
    assert.equal(statement.getText(source), original, 'conditional snapshot changed');
    assert.equal(statements[index - 1]?.getText(source), initializer, 'initializer snapshot changed');
    const start = statement.getStart(source);
    const lineStart = text.lastIndexOf('\n', start - 1) + 1;
    const indent = text.slice(lineStart, start);
    assert.match(indent, /^[ \t]+$/);
    const newline = text.slice(lineStart - 2, lineStart) === '\r\n' ? '\r\n' : '\n';
    fs.writeFileSync(file, text.slice(0, start) + alias + newline + indent + adapted + text.slice(statement.end));
    console.log('applied family 33 alias at 428');
}
