// Stock TypeScript supplies independent UTF-8 unit spans, including signatures.
const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw new Error(`expected TypeScript 6.0.3, got ${ts.version}`);
const [root, output] = process.argv.slice(2);
const result = {};
function visitDirectory(directory) {
  for (const entry of fs.readdirSync(directory, {withFileTypes: true})) {
    const name = path.join(directory, entry.name);
    if (entry.isDirectory()) { visitDirectory(name); continue; }
    if (!entry.isFile()) throw new Error(`unexpected non-file: ${name}`);
    const data = fs.readFileSync(name);
    const relative = path.relative(root, name).split(path.sep).join('/');
    const record = {bytes: data.length, sha256: crypto.createHash('sha256').update(data).digest('hex'), units: [], bodies: []};
    result[relative] = record;
    if (!/\.(ts|a)$/.test(name)) continue;
    const text = data.toString('utf8');
    if (!Buffer.from(text).equals(data)) throw new Error(`invalid UTF-8: ${name}`);
    const file = ts.createSourceFile(name, text, ts.ScriptTarget.Latest, true);
    const bytes = offset => Buffer.byteLength(text.slice(0, offset));
    function add(node) {
      const location = file.getLineAndCharacterOfPosition(node.getStart(file));
      record.units.push({where: `${relative}:${location.line + 1}:${location.character + 1}`,
        start: bytes(node.pos), end: bytes(node.end), function: ts.isFunctionDeclaration(node),
        body_start: node.body ? bytes(node.body.pos) : 0,
        body_end: node.body ? bytes(node.body.end) : 0});
    }
    for (const node of file.statements) if (!ts.isFunctionDeclaration(node)) add(node);
    function walk(node) {
      if (ts.isFunctionDeclaration(node)) add(node);
      if (node.body && (ts.isFunctionDeclaration(node) || ts.isFunctionExpression(node) ||
          ts.isArrowFunction(node) || ts.isMethodDeclaration(node) || ts.isConstructorDeclaration(node) ||
          ts.isGetAccessorDeclaration(node) || ts.isSetAccessorDeclaration(node))) {
        const location = file.getLineAndCharacterOfPosition(node.getStart(file));
        record.bodies.push({where: `${relative}:${location.line + 1}:${location.character + 1}`,
          start: bytes(node.body.pos), end: bytes(node.body.end)});
      }
      ts.forEachChild(node, walk);
    }
    ts.forEachChild(file, walk);
  }
}
visitDirectory(root);
fs.writeFileSync(output, JSON.stringify(result, null, 2) + '\n');
console.log(`stock TypeScript ${ts.version}: ${Object.keys(result).length} files`);
