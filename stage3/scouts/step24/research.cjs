// Syntax sites in the declaration-level parser delta. These are syntactic
// observations, separate from the latent compiler's independent-unit findings.
const fs=require('node:fs');
const path=require('node:path');
const ts=require(process.env.PARSER_TYPESCRIPT);
const tree=process.argv[2], output=process.argv[3];
const closure=JSON.parse(fs.readFileSync(path.join(__dirname,'evidence/value-closure.json'),'utf8'));
const key=d=>`${d.file}:${d.start}:${d.end}`;
const scanner=new Set(closure.scanner.declarations.map(key));
const delta=closure.parser.declarations.filter(d=>!scanner.has(key(d)));
const grouped=new Map();
for(const d of delta) {if(!grouped.has(d.file))grouped.set(d.file,[]);grouped.get(d.file).push(d);}
const report=[];
for(const [file,spans] of grouped) {
 const text=fs.readFileSync(path.join(tree,file),'utf8');
 const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
 const constructs={};
 function add(category,n) {
  const start=n.getStart(source);
  if(!spans.some(s=>s.start<=start&&start<s.end))return;
  const position=source.getLineAndCharacterOfPosition(start);
  const list=constructs[category] ||= [];
  list.push({line:position.line+1,column:position.character+1,text:n.getText(source).replace(/\s+/g,' ').slice(0,180)});
 }
 function visit(n) {
  if(ts.isFunctionLike(n)) {
   if(n.typeParameters?.length)add('generic function',n);
   if(!n.body)add('bodyless function signature',n);
   if(n.body && !ts.isSourceFile(n.parent))add('nested function or callback',n);
  }
  if(ts.isAsExpression(n)||ts.isTypeAssertionExpression(n))add('assertion',n);
  if(ts.isNonNullExpression(n))add('non-null assertion',n);
  if(ts.isTypePredicateNode(n))add('predicate',n);
  if(ts.isModuleDeclaration(n))add('namespace',n);
  if(ts.isEnumDeclaration(n))add('enum',n);
  if(ts.isNewExpression(n))add('construction',n);
  if(ts.isObjectLiteralExpression(n))add('object literal',n);
  if(ts.isBinaryExpression(n)&&n.operatorToken.kind>=ts.SyntaxKind.FirstAssignment&&n.operatorToken.kind<=ts.SyntaxKind.LastAssignment) {
   if(ts.isPropertyAccessExpression(n.left)||ts.isElementAccessExpression(n.left)) add('field or index write',n);
   if(ts.isPropertyAccessExpression(n.left)&&n.left.name.text==='length')add('array length write',n);
  }
  if(ts.isCallExpression(n)&&ts.isPropertyAccessExpression(n.expression)&&['lookAhead','tryScan','scanRange'].includes(n.expression.name.text))add('scanner speculation boundary',n);
  if(ts.isCallExpression(n)&&ts.isIdentifier(n.expression)&&['lookAhead','tryParse','speculationHelper'].includes(n.expression.text))add('parser speculation call',n);
  ts.forEachChild(n,visit);
 }
 visit(source);
 report.push({file,deltaDeclarations:spans.length,constructs:Object.fromEntries(Object.entries(constructs).map(([name,sites])=>[name,{count:sites.length,sites}]))});
}
const shapeSource=ts.createSourceFile('shapes.js','/** @param { } value bad {@link Foo.bar label} */\nfunction f(value) { return value?.name; }\nconst arrow = (x) => x + 1;',ts.ScriptTarget.Latest,false,ts.ScriptKind.JS);
const shapes=[], pending=[shapeSource], seen=new Set();
while(pending.length) {
 const node=pending.pop(), ownKeys=Object.keys(node), signature=ts.SyntaxKind[node.kind]+':'+ownKeys.join(',');
 if(!seen.has(signature)) {seen.add(signature);shapes.push({kind:ts.SyntaxKind[node.kind],ownKeys,prototype:Object.getPrototypeOf(node)?.constructor?.name,
  absentVsUndefined:ownKeys.filter(k=>node[k]===undefined)});}
 const children=[...(node.jsDoc||[])];ts.forEachChild(node,n=>{children.push(n);});pending.push(...children);
}
fs.writeFileSync(output,JSON.stringify({definition:'parser reachable top-level declaration spans minus scanner reachable declaration spans; namespaces kept whole; syntax counts are not latent findings',files:report,
 shapes,statementArrayOwnKeys:Object.keys(shapeSource.statements),jsDocDiagnostics:shapeSource.jsDocDiagnostics.map(d=>({code:d.code,start:d.start,length:d.length}))},null,2)+'\n');
console.log(JSON.stringify(report.map(r=>({file:r.file,declarations:r.deltaDeclarations,counts:Object.fromEntries(Object.entries(r.constructs).map(([k,v])=>[k,v.count]))})),null,2));
