// Stock TypeScript independently records declaration spans and UTF-16 locations.
const fs = require('fs');
const path = require('path');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw new Error(`want TypeScript 6.0.3, got ${ts.version}`);
const [adapted, manifestFile, output] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, 'utf8'));
const names = ['FunctionDeclaration', 'FunctionExpression', 'MethodDeclaration', 'VariableDeclaration', 'Parameter', 'PropertyDeclaration', 'PropertySignature', 'MethodSignature', 'TypeAliasDeclaration', 'InterfaceDeclaration', 'ClassDeclaration', 'EnumDeclaration', 'ModuleDeclaration', 'GetAccessor', 'SetAccessor'];
const kinds = new Map(names.map(name => [ts.SyntaxKind[name], 'Kind' + name]));
const records = {};
for (const entry of manifest) {
  const text = fs.readFileSync(path.join(adapted, entry.file), 'utf8');
  const file = ts.createSourceFile(entry.file, text, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS);
  const declarations = [];
  function visit(node) {
    if (kinds.has(node.kind) && node.name) {
      const {line, character} = file.getLineAndCharacterOfPosition(node.getStart(file));
      declarations.push({where:`${entry.file}:${line+1}:${character+1}`,name:text.slice(node.name.pos,node.name.end),kind:kinds.get(node.kind),start:Buffer.byteLength(text.slice(0,node.pos)),end:Buffer.byteLength(text.slice(0,node.end))});
    }
    ts.forEachChild(node, visit);
  }
  ts.forEachChild(file,visit);
  records[entry.file] = declarations;
}
fs.writeFileSync(output, JSON.stringify(records));
console.log(`stock TypeScript declaration spans: ${manifest.length} files`);
