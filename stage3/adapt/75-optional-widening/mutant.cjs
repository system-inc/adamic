#!/usr/bin/env node
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require(process.env.TSC_ADAPT_TYPESCRIPT || 'typescript');
assert.equal(ts.version, '6.0.3');
const [tree, mode] = process.argv.slice(2);
assert(tree && ['wrong-type', 'drop-owner'].includes(mode));
const file = path.join(tree, 'src/compiler/types.ts');
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
const edits = [];
function visit(n) {
    if (ts.isInterfaceDeclaration(n) && n.name.text === 'Type') {
        for (const m of n.members) {
            if (!['resolvedBaseConstraint', 'resolvedIndexType', 'resolvedStringIndexType'].includes(m.name?.getText(source))) continue;
            if (mode === 'drop-owner') edits.push({ start: m.getFullStart(), end: m.end, text: '' });
            else if (m.name.getText(source) === 'resolvedBaseConstraint') edits.push({ start: m.type.getStart(source), end: m.type.end, text: 'boolean | undefined' });
        }
    }
    ts.forEachChild(n, visit);
}
visit(source);
assert.equal(edits.length, mode === 'drop-owner' ? 3 : 1);
let result = text;
for (const edit of edits.sort((a,b) => b.start-a.start)) result=result.slice(0,edit.start)+edit.text+result.slice(edit.end);
fs.writeFileSync(file,result);
