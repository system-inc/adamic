// Inventory the pinned TypeScript compiler source AST; do not count comments.
// Usage: node inventory.cjs <typescript-module> <compiler-source-directory> <output-json>
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.argv[2]);
const directory = process.argv[3];
const names = new Set(['Error', 'TypeError', 'RangeError', 'SyntaxError', 'ReferenceError', 'EvalError', 'URIError', 'AggregateError']);
const rows = [];
for (const file of fs.readdirSync(directory).filter(file => file.endsWith('.ts')).sort()) {
  const source = ts.createSourceFile(file, fs.readFileSync(path.join(directory, file), 'utf8'), ts.ScriptTarget.Latest, true);
  function record(node, kind) {
    const start = source.getLineAndCharacterOfPosition(node.getStart(source));
    const end = source.getLineAndCharacterOfPosition(node.getEnd());
    rows.push({file: 'src/compiler/' + file, line: start.line + 1, column: start.character + 1,
      endLine: end.line + 1, kind, source: node.getText(source).replace(/\r\n/g, '\n')});
  }
  function visit(node) {
    if ((ts.isNewExpression(node) || ts.isCallExpression(node)) && ts.isIdentifier(node.expression) && names.has(node.expression.text)) {
      record(node, ts.isNewExpression(node) ? 'construction' : 'call-construction');
    }
    if (ts.isCatchClause(node)) record(node, 'catch');
    if (ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression) && node.expression.name.text === 'captureStackTrace') {
      let receiver = node.expression.expression;
      while (ts.isParenthesizedExpression(receiver) || ts.isAsExpression(receiver)) receiver = receiver.expression;
      if (ts.isIdentifier(receiver) && receiver.text === 'Error') record(node, 'capture');
    }
    ts.forEachChild(node, visit);
  }
  visit(source);
}
fs.writeFileSync(process.argv[4], JSON.stringify({typescript: ts.version, rows}, null, 2) + '\n');
