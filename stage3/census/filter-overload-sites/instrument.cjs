'use strict';
// Rewrite a private adapted tree, using an already verified list of source calls.
const fs = require('node:fs'), path = require('node:path');
const ts = require('../../api/node_modules/typescript');
const [root, input] = process.argv.slice(2), rows = JSON.parse(fs.readFileSync(input, 'utf8'));
const byFile = new Map();
for (const row of rows) {
  const file = row.where.replace(/:\d+:\d+$/, '');
  if (!byFile.has(file)) byFile.set(file, []);
  byFile.get(file).push(row);
}
for (const [file, entries] of byFile) {
  const name = path.join(root, file), original = fs.readFileSync(name, 'utf8');
  const source = ts.createSourceFile(file, original, ts.ScriptTarget.Latest, true), edits = [];
  function visit(node) {
    if (ts.isCallExpression(node)) {
      const lc = source.getLineAndCharacterOfPosition(node.getStart());
      const id = file + ':' + (lc.line + 1) + ':' + (lc.character + 1);
      const row = entries.find(row => row.where === id);
      if (row) {
        if (node.getText(source) !== row.call) throw new Error('source mismatch: ' + id);
        const f = node.expression.getText(source);
        // Curry retains the original overload set and all inferred result types.
        edits.push({ start: node.expression.getStart(), end: node.expression.end,
          text: `((globalThis as typeof globalThis & { __adamicFilterSite: (id: string, original: typeof ${f}) => typeof ${f} }).__adamicFilterSite(${JSON.stringify(id)}, ${f}))` });
      }
    }
    ts.forEachChild(node, visit);
  }
  visit(source);
  if (edits.length !== entries.length) throw new Error('missing calls: ' + file);
  let text = original;
  for (const edit of edits.sort((a, b) => b.start - a.start)) text = text.slice(0, edit.start) + edit.text + text.slice(edit.end);
  fs.writeFileSync(name, text);
}
console.log('instrumented ' + rows.length + ' calls');
