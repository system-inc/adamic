#!/usr/bin/env node
'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const nonempty = require('../32-indexed-reads-program/nonempty-endings.cjs');

assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
const file = path.join(process.argv[2], 'src/compiler/moduleSpecifiers.ts');
const text = fs.readFileSync(file, 'utf8');
const parse = input => ts.createSourceFile(file, input, ts.ScriptTarget.Latest, true);
const source = parse(text);
assert.equal(source.parseDiagnostics.length, 0);

function one(items, label) {
    assert.equal(items.length, 1, label);
    return items[0];
}

function owner(name) {
    return one(source.statements.filter(node => ts.isFunctionDeclaration(node) &&
        node.name?.text === name && node.body), name);
}

const edits = [];
function edit(start, end, before, after) {
    edits.push({start, end, before, after});
}
function replace(node, before, after) {
    assert(node, `missing ${before}`);
    edit(node.getStart(source), node.end, before, after);
}

const array = 'readonly ModuleSpecifierEnding[]';
const tuple = '[ModuleSpecifierEnding, ...ModuleSpecifierEnding[]]';
for (const [name, before] of [
    ['processEnding', 'readonly ' + tuple],
    ['tryGetModuleNameFromRootDirs', 'readonly ' + tuple],
    ['tryGetModuleNameFromPaths', 'ModuleSpecifierEnding[]'],
]) {
    const parameter = one(owner(name).parameters.filter(node => node.name.getText(source) === 'allowedEndings'), name);
    replace(parameter.type, before, array);
}

for (const [name, call] of [
    ['getLocalModuleSpecifier', 'getAllowedEndingsInPrefererredOrder(importMode)'],
    ['tryGetModuleNameAsNodeModule', 'preferences.getAllowedEndingsInPreferredOrder()'],
]) {
    const local = one(owner(name).body.statements.filter(ts.isVariableStatement)
        .flatMap(node => [...node.declarationList.declarations])
        .filter(node => node.name.getText(source) === 'allowedEndings'), name);
    replace(local, `allowedEndings = ${call} as ${tuple}`, `allowedEndings: ${array} = ${call}`);
}

const statements = owner('processEnding').body.statements;
const endingSwitch = one(statements.filter(ts.isSwitchStatement), 'processEnding switch');
const previous = statements[statements.indexOf(endingSwitch) - 1];
const firstRead = 'const firstEnding = allowedEndings[0]!;';
const newline = text.includes('\r\n') ? '\r\n' : '\n';
// Put the binding in the blank line before the switch to retain source line numbers.
const start = previous.getText(source) === firstRead ? previous.getFullStart() : endingSwitch.getFullStart();
edit(start, endingSwitch.expression.end,
    newline + newline + '    switch (allowedEndings[0]!',
    newline + '    ' + firstRead + newline + '    switch (firstEnding');
const fallback = one(endingSwitch.caseBlock.clauses.filter(ts.isDefaultClause), 'processEnding default');
replace(one(fallback.statements, 'processEnding fallback'),
    'return Debug.assertNever(allowedEndings[0]);', 'return Debug.assertNever(firstEnding);');

edits.sort((a, b) => b.start - a.start);
const before = edits.every(({start, end, before}) => text.slice(start, end) === before);
const after = edits.every(({start, end, after}) => text.slice(start, end) === after);
assert(before || after, 'moduleSpecifiers.ts preference-list snapshot drift');

function withEdits(side) {
    let result = text;
    for (const {start, end, [side]: replacement} of edits) {
        result = result.slice(0, start) + replacement + result.slice(end);
    }
    return result;
}

// Reuse the producer and no-write proofs introduced with the original tuple refinement.
nonempty.validate(ts, before ? source : parse(withEdits('before')));
if (before) {
    assert.equal(fs.readFileSync(file, 'utf8'), text, 'moduleSpecifiers.ts changed during adaptation');
    fs.writeFileSync(file, withEdits('after'));
}
console.log(`moduleSpecifiers.ts preference list: ${before ? edits.length : 0} edits`);
