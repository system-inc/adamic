// Batched stock-AST source identities. Census columns are one-based UTF-16.
const fs = require('fs'), path = require('path'), crypto = require('crypto');
const request = JSON.parse(fs.readFileSync(0, 'utf8'));
const ts = require(request.typescript);
if (ts.version !== '6.0.3') throw Error(`stock TypeScript 6.0.3 required, got ${ts.version}`);
const result = {};
for (const [file, locations] of Object.entries(request.files)) {
  const bytes = fs.readFileSync(path.join(request.tree, file));
  const text = bytes.toString('utf8');
  const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
  const sha = crypto.createHash('sha256').update(bytes).digest('hex');
  const access = n => ts.isPropertyAccessExpression(n) || ts.isElementAccessExpression(n);
  for (const [where, line, column] of locations) {
    const starts = source.getLineStarts();
    if (line < 1 || line > starts.length || column < 1) throw Error(`invalid location: ${where}`);
    const offset = starts[line - 1] + column - 1;
    if (offset >= (starts[line] ?? text.length + 1)) throw Error(`invalid column: ${where}`);
    // Descend to the smallest node containing this position, then retain the
    // outermost access ancestor. Do not merge separate accesses on one line.
    let node = source;
    function descend(n) {
      ts.forEachChild(n, child => {
        if (child.getStart(source) <= offset && offset < child.end) {
          node = child; descend(child); return true;
        }
      });
    }
    descend(source);
    let normalized = offset;
    let chain = access(node) ? node : null;
    for (let n = node.parent; n && access(n); n = n.parent) chain = n;
    if (chain) normalized = chain.getStart(source);
    const byteOffset = n => Buffer.byteLength(text.slice(0, n), 'utf8');
    result[where] = {file, sha256: sha, byte_offset: byteOffset(offset), read_offset: byteOffset(normalized)};
  }
}
process.stdout.write(JSON.stringify(result));
