// Extract original declarations and reads using the independently pinned checker.
const fs = require('fs'), path = require('path'), child = require('child_process'), crypto = require('crypto');
const root = process.argv[2];
const ts = require(path.join(root, 'lib/typescript.js'));
const pin = '050880ce59e30b356b686bd3144efe24f875ebc8';
if (child.execFileSync('git', ['-C', root, 'rev-parse', 'HEAD'], {encoding:'utf8'}).trim() !== pin) throw Error('wrong original pin');
const config = path.join(root, 'src/compiler/tsconfig.json');
const parsed = ts.parseJsonConfigFileContent(ts.readConfigFile(config, ts.sys.readFile).config, ts.sys, path.dirname(config), undefined, config);
const program = ts.createProgram(parsed.fileNames, parsed.options), checker = program.getTypeChecker();
const ledger = JSON.parse(fs.readFileSync(path.join(__dirname, '../unknown-callable-pairs-ranked.json')));
const nodes = new Map();
for (const source of program.getSourceFiles()) {
 if (!source.fileName.startsWith(root)) continue;
 function visit(n) {
  if (ts.isInterfaceDeclaration(n) || ts.isTypeAliasDeclaration(n) || ts.isEnumDeclaration(n)) nodes.set(n.name.text, n);
  ts.forEachChild(n, visit);
 }
 visit(source);
}
const hash = file => crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');
const evidence = [];
for (const pair of ledger.filter(p => p.rank % 3 === 2 && !p.certified)) {
 const source = program.getSourceFile(path.join(root, pair.witness.file));
 let read;
 function visit(n) {
  if ((ts.isPropertyAccessExpression(n) || ts.isElementAccessExpression(n) || ts.isBindingElement(n))) {
   const loc = source.getLineAndCharacterOfPosition(n.getStart(source));
   const field = ts.isPropertyAccessExpression(n) ? n.name.text : ts.isBindingElement(n) ? (n.propertyName || n.name).getText(source) : pair.field;
   if (loc.line+1 === pair.witness.line && loc.character+1 === pair.witness.column && field === pair.field && (!read || n.end > read.end)) read = n;
  }
  ts.forEachChild(n, visit);
 }
 visit(source);
 if (!read) throw Error('missing read ' + pair.rank);
 const symbol = ts.isPropertyAccessExpression(read) ? checker.getSymbolAtLocation(read.name) : ts.isBindingElement(read) ? checker.getPropertyOfType(checker.getTypeAtLocation(read.parent), pair.field) : checker.getSymbolAtLocation(read);
 const declarations = (symbol?.declarations || []).map(d => ({text:(ts.isFunctionDeclaration(d) || ts.isMethodDeclaration(d)) && d.body ? d.getSourceFile().text.slice(d.getStart(), d.body.getStart()).trim()+';' : d.getText(), file:path.relative(root,d.getSourceFile().fileName), sha256:hash(d.getSourceFile().fileName)}));
 evidence.push({...pair, read:read.getText(source), start:read.getStart(source), end:read.end, fileSha256:hash(source.fileName), declarations});
}
fs.writeFileSync(path.join(__dirname,'original-members.json'), JSON.stringify({pin, members:evidence}, null, 2)+'\n');
// Only declaration text, never upstream implementation code, is emitted.
const declarations = {};
for (const [name,n] of nodes) declarations[name] = {kind:ts.SyntaxKind[n.kind], text:n.getText().replace(/^export\s+/, ''), kindValue:undefined};
for (const [name,n] of nodes) if (ts.isEnumDeclaration(n)) declarations[name].values = Object.fromEntries(n.members.map(m => [m.name.getText(), checker.getConstantValue(m)]));
for (const [name,n] of nodes) if (ts.isInterfaceDeclaration(n)) {
 const kind = checker.getPropertyOfType(checker.getTypeAtLocation(n), 'kind');
 if (kind) {
  const type = checker.getTypeOfSymbolAtLocation(kind,n);
  if (typeof type.value === 'number') declarations[name].kindValue = type.value;
 }
}
fs.writeFileSync('/tmp/lane5-c-carrier-declarations.json', JSON.stringify(declarations,null,2)+'\n');
console.log('Extracted '+evidence.length+' remaining share-c members with original read spans and declarations.');
