// Resolve scan sites to lexical top-level units without rewriting raw scan owners.
const fs = require('fs');
const path = require('path');
const req = JSON.parse(fs.readFileSync(0, 'utf8'));
const ts = require(req.typescript);
const cache = new Map();
const result = {};
for (const where of req.locations) {
  const match = /^(.*):(\d+):(\d+)$/.exec(where);
  const file = match[1];
  if (!cache.has(file)) {
    cache.set(file, ts.createSourceFile(file, fs.readFileSync(path.join(req.tree, file), 'utf8'), ts.ScriptTarget.Latest, true));
  }
  const source = cache.get(file);
  const offset = source.getPositionOfLineAndCharacter(+match[2] - 1, +match[3] - 1);
  const node = source.statements.find(n => n.getStart(source) <= offset && offset < n.end);
  if (!node) throw new Error(`no top-level owner for ${where}`);
  const lc = source.getLineAndCharacterOfPosition(node.getStart(source));
  result[where] = `${file}:${lc.line+1}:${lc.character+1}`;
}
process.stdout.write(JSON.stringify(result));
