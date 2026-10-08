// Stock TypeScript audit of every declaration in the gathered scanner closure.
const fs = require('node:fs'), path = require('node:path');
const ts = require(process.env.SLICE_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw Error('TypeScript 6.0.3 required');
const tree = path.resolve(process.argv[2]);
const manifest = JSON.parse(fs.readFileSync(path.join(tree, 'slice.json')));
const files = [...new Set(manifest.evaluation.map(d => d.file))].sort();
const program = ts.createProgram(files.map(f => path.join(tree, f)), { target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.ESNext, moduleResolution: ts.ModuleResolutionKind.Bundler, strict: true, skipLibCheck: true, allowImportingTsExtensions: true });
const checker = program.getTypeChecker(), sites = [];
const upstream = process.argv[3] && path.resolve(process.argv[3]);
function record(site, node) {
 if (upstream) {
  const raw = fs.readFileSync(path.join(upstream, site.file), 'utf8');
  const offset = raw.indexOf(node.getText());
  if (offset < 0) throw Error('upstream span missing: ' + site.file + ':' + site.line);
  const original = ts.createSourceFile(site.file, raw, ts.ScriptTarget.Latest, true);
  const pos = original.getLineAndCharacterOfPosition(offset);
  site.upstream = {line:pos.line+1,column:pos.character+1};
 }
 sites.push(site);
}
for (const file of files) {
 const source = program.getSourceFile(path.join(tree, file));
 function visit(node) {
  const pos = source.getLineAndCharacterOfPosition(node.getStart(source));
  if (ts.isComputedPropertyName(node)) record({file, line:pos.line+1, column:pos.character+1, kind:'computed-name', text:node.getText(source), parent:ts.SyntaxKind[node.parent.kind]}, node);
  if (ts.isFunctionDeclaration(node) || ts.isMethodSignature(node) || ts.isCallSignatureDeclaration(node) || ts.isFunctionExpression(node) || ts.isArrowFunction(node)) {
   const signature = checker.getSignatureFromDeclaration(node);
   if (signature) {
    const type = checker.getReturnTypeOfSignature(signature), members = type.isUnion() ? type.types : [type];
    if (members.some(t => t.flags & ts.TypeFlags.TypeParameter)) record({file,line:pos.line+1,column:pos.character+1,kind:'generic-return',name:node.name?.getText(source) || '<anonymous>',returns:checker.typeToString(type), body:!!node.body}, node);
   }
  }
  ts.forEachChild(node,visit);
 }
 visit(source);
}
console.log(JSON.stringify({typescript:ts.version, upstream_commit:'050880ce59e30b356b686bd3144efe24f875ebc8', closure:manifest.summary, files, sites},null,2));
