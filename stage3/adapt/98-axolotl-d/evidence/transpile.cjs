'use strict';
// Transpiles each file the adapter edits from two trees with stock TypeScript 6.0.3's transpileModule and
// reports, per file, whether the JavaScript is byte-identical and, if not, the differing lines.
// Usage: node transpile.cjs <before-tree> <after-tree>
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');
const {declarations} = require('./adapt.cjs');

assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 4, 'usage: node transpile.cjs <before-tree> <after-tree>');
const [before, after] = process.argv.slice(2).map(tree => path.resolve(tree));
const files = [...new Set([...declarations.map(row => row.file), 'src/compiler/moduleNameResolver.ts', 'src/compiler/sourcemap.ts'])];
const options = {compilerOptions: {target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.ESNext, removeComments: false}};

function transpile(tree, file) {
    return ts.transpileModule(fs.readFileSync(path.join(tree, file), 'utf8'), {...options, fileName: file}).outputText;
}

const report = [];
for (const file of files) {
    const left = transpile(before, file).split('\n');
    const right = transpile(after, file).split('\n');
    const row = {file, identical: left.join('\n') === right.join('\n'), bytes: [left.join('\n').length, right.join('\n').length]};
    if (!row.identical) {
        // The edits never add or remove a line outside sourcemap.ts, so compare by aligning the common prefix and suffix.
        let start = 0;
        while (start < left.length && start < right.length && left[start] === right[start]) start++;
        let end = 0;
        while (end < left.length - start && end < right.length - start && left[left.length - 1 - end] === right[right.length - 1 - end]) end++;
        row.line = start + 1;
        row.before = left.slice(start, left.length - end);
        row.after = right.slice(start, right.length - end);
    }
    report.push(row);
}
console.log(JSON.stringify(report, null, 2));
