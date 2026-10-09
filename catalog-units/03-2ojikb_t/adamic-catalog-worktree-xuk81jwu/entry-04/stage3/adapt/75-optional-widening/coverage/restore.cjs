#!/usr/bin/env node
'use strict';
// Restore only our scratch AST insertions to create a control source view.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');
const [tree, output, manifest] = process.argv.slice(2);
const sites = JSON.parse(fs.readFileSync(manifest));
let count = 0;
for (const file of new Set([...sites.map(r => r.File), 'core.ts'])) {
 const filename = path.join(tree, 'src/compiler', file);
 let text = fs.readFileSync(filename, 'utf8');
 const f = ts.createSourceFile(filename, text, 99, true);
 const edits = [];
 function visit(n) {
  if (ts.isParenthesizedExpression(n) && ts.isBinaryExpression(n.expression) && n.expression.operatorToken.kind === ts.SyntaxKind.CommaToken && ts.isCallExpression(n.expression.left) && n.expression.left.expression.getText(f) === 'optionalWideningCoverageHit') {
   edits.push({start: n.getStart(f), end: n.expression.right.getStart(f)});
   edits.push({start: n.expression.right.end, end: n.end});
   count++;
  }
  if (ts.isImportDeclaration(n) && n.getText(f).includes('optionalWideningCoverage')) edits.push({start: n.getStart(f), end: n.end + 1});
  ts.forEachChild(n, visit);
 }
 visit(f);
 edits.sort((a, b) => b.start - a.start);
 for (const e of edits) text = text.slice(0, e.start) + text.slice(e.end);
 if (file === 'core.ts') {
  const at = text.indexOf('\n// Scratch-only optional widening measurement.');
  assert(at >= 0); text = text.slice(0, at);
 }
 const target = path.join(output, 'src/compiler', file);
 fs.mkdirSync(path.dirname(target), {recursive: true});
 fs.writeFileSync(target, text);
}
assert.equal(count, 47);
console.log(JSON.stringify({restored: count}));
