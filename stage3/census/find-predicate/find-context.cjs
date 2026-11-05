// Classify predicate/overload diagnostics at a direct core find call with stock TypeScript.
const fs = require('fs');
const path = require('path');
const [tree, packagePath, stream] = process.argv.slice(2);
const ts = require(packagePath);
const rows = fs.readFileSync(stream, 'utf8').trim().split('\n').map(JSON.parse);
const files = new Map();
const results = new Map();
for (const row of rows.slice(1)) for (const finding of row.findings) {
  if (!/predicate|overload/.test(finding.reason) || results.has(finding.where)) continue;
  const [, file, line, column] = finding.where.match(/^(.*):(\d+):(\d+)$/);
  const relative = file.includes('/src/') ? 'src/' + file.split('/src/')[1] : file;
  if (!files.has(relative)) {
    const text = fs.readFileSync(path.join(tree, relative), 'utf8');
    files.set(relative, ts.createSourceFile(relative, text, ts.ScriptTarget.Latest, true));
  }
  const source = files.get(relative);
  const pos = source.getPositionOfLineAndCharacter(+line - 1, +column - 1);
  let node = source;
  function walk(parent) {
    ts.forEachChild(parent, child => {
      if (child.getStart(source) <= pos && pos < child.end) { node = child; walk(child); }
    });
  }
  walk(source);
  const ancestors = [];
  let directFind = false;
  for (let n = node; n; n = n.parent) {
    if (ts.isCallExpression(n)) {
      const callee = n.expression.getText(source);
      ancestors.push(callee);
      if (callee === 'find' || callee === 'ts.find') directFind = true;
    }
    if (ts.isFunctionLike(n) && n !== node) break;
  }
  results.set(finding.where, {where: finding.where, node: ts.SyntaxKind[node.kind], direct_find: directFind, calls: ancestors});
}
process.stdout.write(JSON.stringify([...results.values()], null, 2) + '\n');
