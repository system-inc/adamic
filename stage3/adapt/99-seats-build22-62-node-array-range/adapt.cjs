#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

if (ts.version !== '6.0.3' || process.argv.length !== 3) {
    throw new Error('usage: node adapt.cjs <TypeScript 6.0.3 tree>');
}

const fileName = path.resolve(process.argv[2], 'src/compiler/visitorPublic.ts');
const text = fs.readFileSync(fileName, 'utf8');
const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
const functions = source.statements.filter(node =>
    ts.isFunctionDeclaration(node) && node.name?.text === 'addDefaultValueAssignmentsIfNeeded');
if (functions.length !== 1 || !functions[0].body) throw new Error('expected one default assignment worker');
const branches = functions[0].body.statements.filter(node =>
    ts.isIfStatement(node) && node.expression.getText(source) === 'result' && !node.elseStatement);
if (branches.length !== 1) throw new Error('expected one result branch');
const branch = branches[0];
const before = [
    'if (result) {',
    '        return setTextRange(context.factory.createNodeArray(result, parameters.hasTrailingComma), parameters);',
    '    }',
];
const after = [
    'if (result) {',
    '        const array = context.factory.createNodeArray(result, parameters.hasTrailingComma);',
    '        const pos = parameters.pos;',
    '        const end = parameters.end;',
    '        array.pos = pos;',
    '        array.end = end;',
    '        return array;',
    '    }',
];
const newline = text.includes('\r\n') ? '\r\n' : '\n';
const written = branch.getText(source);
if (written !== after.join(newline)) {
    if (written !== before.join(newline)) throw new Error('unexpected result branch: ' + written);
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error('source changed: ' + fileName);
    fs.writeFileSync(fileName, text.slice(0, branch.getStart(source)) + after.join(newline) + text.slice(branch.end));
}
console.log(JSON.stringify({ edits: written === before.join(newline) ? 1 : 0 }));
