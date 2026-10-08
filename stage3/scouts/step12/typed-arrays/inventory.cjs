// Stock TypeScript checker inventory of the gathered scanner declarations.
const fs = require('node:fs'), path = require('node:path');
const ts = require(process.env.SLICE_TYPESCRIPT || 'typescript');
const [tree, manifestFile] = process.argv.slice(2);
if (!tree || !manifestFile || ts.version !== '6.0.3') throw Error('usage: inventory.cjs TREE SLICE_MANIFEST (stock TS 6.0.3)');
const manifest = JSON.parse(fs.readFileSync(manifestFile));
const files = [...new Set(manifest.declarations.map(d => d.file))];
const program = ts.createProgram(files.map(f => path.resolve(tree, f)), {target:ts.ScriptTarget.ESNext, module:ts.ModuleKind.ESNext, moduleResolution:ts.ModuleResolutionKind.Bundler, strict:true, noUncheckedIndexedAccess:true});
const checker = program.getTypeChecker();
const kinds = new Set(['Int8Array','Uint8Array','Uint8ClampedArray','Int16Array','Uint16Array','Int32Array','Uint32Array','Float32Array','Float64Array','BigInt64Array','BigUint64Array']);
function typed(node) {
 const t = checker.getTypeAtLocation(node);
 const parts = t.isUnion() ? t.types : [t];
 return parts.map(p => p.getSymbol()?.getName()).filter(n => kinds.has(n));
}
const sites = [], mentions = [];
for (const file of files) {
 const sf = program.getSourceFile(path.resolve(tree,file));
 const selected = manifest.declarations.filter(d => d.file === file).flatMap(d => d.names);
 function loc(n) {const p=sf.getLineAndCharacterOfPosition(n.getStart(sf));return `${file}:${p.line+1}:${p.character+1}`;}
 function visit(n) {
  let operations = [], receiver;
  if (ts.isNewExpression(n) && typed(n).length) {operations=['constructor'];receiver=n;}
  if (ts.isElementAccessExpression(n) && typed(n.expression).length) {
   receiver=n.expression;
   const p=n.parent;
   if (ts.isBinaryExpression(p) && p.left===n && p.operatorToken.kind>=ts.SyntaxKind.FirstAssignment && p.operatorToken.kind<=ts.SyntaxKind.LastAssignment) operations=p.operatorToken.kind===ts.SyntaxKind.EqualsToken?['write']:['read','write'];
   else if (ts.isPrefixUnaryExpression(p)||ts.isPostfixUnaryExpression(p)) operations=['read','write'];
   else operations=['read'];
  }
  if (ts.isPropertyAccessExpression(n) && typed(n.expression).length) {operations=[n.name.text];receiver=n.expression;}
  if (operations.length) sites.push({location:loc(n),operations,kinds:typed(receiver),expression:n.getText(sf),statement:ts.isElementAccessExpression(n)&&ts.isBinaryExpression(n.parent)?n.parent.getText(sf):n.getText(sf)});
  if (ts.isIdentifier(n) && kinds.has(n.text)) mentions.push({location:loc(n),text:n.text});
  ts.forEachChild(n,visit);
 }
 for (const s of sf.statements) {
  let names=[];
  if (ts.isVariableStatement(s)) names=s.declarationList.declarations.filter(d=>ts.isIdentifier(d.name)).map(d=>d.name.text);
  else if (s.name) names=[s.name.text];
  if (names.some(n=>selected.includes(n))) visit(s);
 }
}
console.log(JSON.stringify({typescript:ts.version,files,scope:'Whole retained top-level declarations; namespace wrappers conservatively scanned whole',sites,mentions,subarray:sites.filter(s=>s.operations.includes('subarray'))},null,2));
