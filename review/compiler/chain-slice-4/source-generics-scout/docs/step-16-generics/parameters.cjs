// TypeScript's own AST supplies binder names and their lexical witness scopes.
const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const ts = require(process.argv[2]);
const root = process.argv[3];
const names = new Set();
const scopes = {};
const genericUses = {};
const hashes = {};
function walk(dir) {
  for (const entry of fs.readdirSync(dir, {withFileTypes:true})) {
    const filename = path.join(dir, entry.name);
    if (entry.isDirectory()) { walk(filename); continue; }
    if (!filename.endsWith('.ts') && !filename.endsWith('.a')) continue;
    const bytes = fs.readFileSync(filename);
    const text = bytes.toString('utf8');
    hashes[path.relative(root, filename)] = crypto.createHash('sha256').update(bytes).digest('hex');
    const source = ts.createSourceFile(filename, text, ts.ScriptTarget.Latest, true);
    const ranges = [];
    const uses = [];
    function visit(node) {
      if (node.typeParameters) {
        const bound = node.typeParameters.map(p => p.name.text);
        bound.forEach(n => names.add(n));
        const start = source.getLineAndCharacterOfPosition(node.getStart(source));
        const end = source.getLineAndCharacterOfPosition(node.end);
        ranges.push({start:[start.line+1,start.character+1], end:[end.line+1,end.character+1], names:bound});
      }
      if (ts.isAsExpression(node) || ts.isTypePredicateNode(node)) {
        const bound = new Set();
        for (let owner = node; owner; owner = owner.parent) {
          if (owner.typeParameters) owner.typeParameters.forEach(p => bound.add(p.name.text));
        }
        let generic = false;
        function inType(type) {
          if (ts.isIdentifier(type) && bound.has(type.text)) generic = true;
          ts.forEachChild(type, inType);
        }
        if (node.type) inType(node.type);
        if (generic) {
          const start = source.getLineAndCharacterOfPosition(node.getStart(source));
          const end = source.getLineAndCharacterOfPosition(node.end);
          uses.push({start:[start.line+1,start.character+1], end:[end.line+1,end.character+1]});
        }
      }
      ts.forEachChild(node, visit);
    }
    visit(source);
    scopes[path.relative(root, filename)] = ranges;
    genericUses[path.relative(root, filename)] = uses;
  }
}
walk(root);
process.stdout.write(JSON.stringify({names:[...names].sort(), scopes, genericUses, hashes}));
