#!/usr/bin/env node
'use strict';
// Undo exactly one selected range input in a disposable tree.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3' || process.argv.length !== 3) throw new Error('usage: node mutant.cjs <scratch-tree>');
const name = path.join(path.resolve(process.argv[2]), 'src/compiler/utilities.ts');
const text = fs.readFileSync(name, 'utf8'), sf = ts.createSourceFile(name, text, ts.ScriptTarget.Latest, true);
const found = [];
function visit(n) {
    if (ts.isFunctionDeclaration(n) && n.name?.text === 'rangeStartPositionsAreOnSameLine') {
        const parameter = n.parameters.find(p => p.name.getText(sf) === 'range1');
        if (parameter?.type?.getText(sf) === 'Readonly<TextRange>') found.push(parameter.type);
    }
    ts.forEachChild(n, visit);
}
visit(sf);
if (found.length !== 1) throw new Error('expected exactly one adapted range input');
const node = found[0];
fs.writeFileSync(name, text.slice(0, node.getStart(sf)) + 'TextRange' + text.slice(node.end));
console.log('undid rangeStartPositionsAreOnSameLine.range1');
