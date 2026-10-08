// Independently enumerate body units with stock TypeScript rather than the Go AST.
const fs = require('fs');
const path = require('path');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw new Error(`want TypeScript 6.0.3, got ${ts.version}`);
const [adapted, manifestPath, output] = process.argv.slice(2);
const result = {};
for (const entry of JSON.parse(fs.readFileSync(manifestPath, 'utf8'))) {
  const text = fs.readFileSync(path.join(adapted, entry.file), 'utf8');
  const file = ts.createSourceFile(entry.file, text, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS);
  const bytes = offset => Buffer.byteLength(text.slice(0, offset));
  const where = node => {
    const loc = file.getLineAndCharacterOfPosition(node.getStart(file));
    return `${entry.file}:${loc.line + 1}:${loc.character + 1}`;
  };
  const units = [];
  const visit = node => {
    if (ts.isFunctionDeclaration(node)) {
      let parent = node.parent;
      let depth = 0;
      let parentFunction;
      while (parent && parent !== file) {
        if (ts.isFunctionDeclaration(parent)) { depth++; parentFunction ||= parent; }
        parent = parent.parent;
      }
      units.push({where: where(node), name: node.name?.text || '', depth,
        parent_function: parentFunction ? where(parentFunction) : '',
        start: bytes(node.pos), end: bytes(node.end),
        body_start: node.body ? bytes(node.body.pos) : 0,
        body_end: node.body ? bytes(node.body.end) : 0});
    }
    ts.forEachChild(node, visit);
  };
  ts.forEachChild(file, visit);
  result[entry.file] = units;
}
fs.writeFileSync(output, JSON.stringify(result, null, 2) + '\n');
console.log(`stock TypeScript ${ts.version}: ${Object.keys(result).length} files, ${Object.values(result).reduce((n, units) => n + units.length, 0)} function declarations`);
