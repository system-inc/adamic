// Stock TypeScript parser inventory. Generated sources are excluded, as in the census.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const root = path.resolve(process.argv[2]);
const rows = [];
function walk(dir) {
 for (const name of fs.readdirSync(dir).sort()) {
  const file = path.join(dir, name);
  if (fs.statSync(file).isDirectory()) { walk(file); continue; }
  if (!file.endsWith('.ts') || file.includes('.generated.')) continue;
  const sf = ts.createSourceFile(file, fs.readFileSync(file, 'utf8'), ts.ScriptTarget.Latest, true);
  function add(n, reason) {
   const p = sf.getLineAndCharacterOfPosition(n.getStart(sf));
   rows.push({file:path.relative(root, file), line:p.line+1, column:p.character+1, reason, text:sf.text.split(/\r?\n/)[p.line]});
  }
  function visit(n) {
   if (ts.isObjectLiteralExpression(n)) n.properties.forEach((p,i) => { if (i && ts.isSpreadAssignment(p)) add(p,'a spread after the first field'); });
   if ((ts.isPropertyDeclaration(n) || ts.isVariableDeclaration(n)) && n.exclamationToken) add(n,'a definite assignment assertion !');
   if (ts.isClassDeclaration(n) || ts.isClassExpression(n)) {
    add(n,'class declaration or expression');
    let a=n.parent; while(a && !ts.isFunctionLike(a)) a=a.parent;
    if(a) add(n,'a class inside a function');
   }
   if (ts.isBindingElement(n) && n.initializer) add(n,'destructuring with a default');
   if (ts.isBindingElement(n) && n.dotDotDotToken) add(n,'destructuring with rest');
   if (ts.isParameter(n) && (ts.isObjectBindingPattern(n.name) || ts.isArrayBindingPattern(n.name)) && n.parent.parameters.some(p=>p.initializer)) add(n,'a destructured parameter beside a parameter with a default');
   if (ts.isCallExpression(n) && (n.flags & ts.NodeFlags.OptionalChain)) add(n,'a call through ?. (an optional call)');
   if (ts.isCallExpression(n) && n.questionDotToken) add(n,'direct optional-call token');
   if (ts.isYieldExpression(n)) add(n,'yield (generators)');
   if (ts.isDeleteExpression(n)) add(n,'delete');
   if (ts.isDebuggerStatement(n)) add(n,'debugger');
   if (ts.isBinaryExpression(n) && n.operatorToken.kind === ts.SyntaxKind.InKeyword) add(n.operatorToken,'in');
   if (ts.isGetAccessorDeclaration(n)) add(n,'a getter');
   if (ts.isSetAccessorDeclaration(n)) add(n,'a setter');
   ts.forEachChild(n,visit);
  }
  visit(sf);
 }
}
walk(path.join(root,'src/compiler'));
// Resolve receivers with the stock checker; match the census's narrow direct-literal ledger.
const names = [];
function sources(dir) { for (const name of fs.readdirSync(dir).sort()) { const file=path.join(dir,name); if(fs.statSync(file).isDirectory()) sources(file); else if(file.endsWith('.ts') && !file.includes('.generated.')) names.push(file); } }
sources(path.join(root,'src/compiler'));
const program=ts.createProgram(names,{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext});
const checker=program.getTypeChecker();
const writes=[];
for(const sf of program.getSourceFiles().filter(f=>names.includes(f.fileName))) {
 function visit(n) {
  if(ts.isBinaryExpression(n) && n.operatorToken.kind===ts.SyntaxKind.EqualsToken && ts.isPropertyAccessExpression(n.left) && ts.isIdentifier(n.left.expression)) {
   const decl=checker.getSymbolAtLocation(n.left.expression)?.valueDeclaration;
   if(decl && ts.isVariableDeclaration(decl) && decl.initializer) {
    let initial=decl.initializer; while(ts.isAsExpression(initial) || ts.isParenthesizedExpression(initial)) initial=initial.expression;
    if((ts.isObjectLiteralExpression(initial) && !initial.properties.some(p=>p.name && p.name.getText(sf)===n.left.name.text)) || (ts.isArrayLiteralExpression(initial) && n.left.name.text!=='length')) {
     const pos=sf.getLineAndCharacterOfPosition(n.getStart(sf));
     writes.push({file:path.relative(root,sf.fileName),line:pos.line+1,column:pos.character+1,object:n.left.expression.text,property:n.left.name.text,initial_kind:ts.isArrayLiteralExpression(initial)?'array':'object',declared_property:!!checker.getPropertyOfType(checker.getTypeAtLocation(n.left.expression),n.left.name.text)});
    }
   }
  }
  ts.forEachChild(n,visit);
 }
 visit(sf);
}
const counts={}; for(const r of rows) counts[r.reason]=(counts[r.reason] || 0)+1;
fs.writeFileSync(process.argv[3],JSON.stringify({typescript:ts.version,counts,post_creation_writes:writes,sites:rows},null,2)+'\n');
console.log(JSON.stringify({counts,post_creation_writes:writes.length}));
