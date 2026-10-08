#!/usr/bin/env node
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require(process.env.TSC_ADAPT_TYPESCRIPT || 'typescript');
assert.equal(ts.version, '6.0.3');
const [tree, mode] = process.argv.slice(2);
assert(tree && ['wrong-type', 'drop-owner', 'wrong-json-type', 'drop-json-owner'].includes(mode));
const file = path.join(tree, 'src/compiler/types.ts');
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
const edits = [];
function visit(n) {
    if (ts.isInterfaceDeclaration(n) && n.name.text === (mode.includes('json') ? 'JsonSourceFile' : 'Type')) {
        for (const m of n.members) {
            const properties = mode.includes('json') ? ['extendedSourceFiles', 'configFileSpecs'] : ['resolvedBaseConstraint', 'resolvedIndexType', 'resolvedStringIndexType'];
            if (!properties.includes(m.name?.getText(source))) continue;
            if (mode.startsWith('drop-')) edits.push({ start: m.getFullStart(), end: m.end, text: '' });
            else if (mode === 'wrong-json-type' && m.name.getText(source) === 'extendedSourceFiles') edits.push({ start: m.type.getStart(source), end: m.type.end, text: 'number[]' });
            else if (mode === 'wrong-type' && m.name.getText(source) === 'resolvedBaseConstraint') edits.push({ start: m.type.getStart(source), end: m.type.end, text: 'boolean | undefined' });
        }
    }
    ts.forEachChild(n, visit);
}
visit(source);
assert.equal(edits.length, mode === 'drop-owner' ? 3 : mode === 'drop-json-owner' ? 2 : 1);
let result = text;
for (const edit of edits.sort((a,b) => b.start-a.start)) result=result.slice(0,edit.start)+edit.text+result.slice(edit.end);
fs.writeFileSync(file,result);
