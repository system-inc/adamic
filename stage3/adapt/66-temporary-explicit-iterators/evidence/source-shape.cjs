const fs=require('fs'),crypto=require('crypto'),assert=require('assert/strict');
const {ts,nodes}=require('../census.cjs');
const beforeTree=process.argv[2], afterTree=process.argv[3];
assert(beforeTree&&afterTree,'usage: source-shape.cjs BEFORE AFTER');
const report=[];
for(const file of ['core.ts','checker.ts']) {
 const before=ts.createSourceFile(file,fs.readFileSync(beforeTree+'/src/compiler/'+file,'utf8'),99,true);
 const after=ts.createSourceFile(file,fs.readFileSync(afterTree+'/src/compiler/'+file,'utf8'),99,true);
 const signatures=[];
 for(const g of nodes(before,n=>ts.isFunctionLike(n)&&n.asteriskToken)) {
   const name=g.name?.getText(before)||'<anonymous>';
   const candidates=nodes(after,n=>ts.isFunctionLike(n)&&((n.name?.getText(after)||'<anonymous>')===name)&&n.kind===g.kind);
   const signature=g.getText(before).slice(0,g.body.getStart(before)-g.getStart(before)).replace('*','');
   assert(candidates.some(n=>n.getText(after).slice(0,n.body.getStart(after)-n.getStart(after))===signature),name);
   signatures.push(name);
 }
 const anyCount=s=>nodes(s,n=>n.kind===ts.SyntaxKind.AnyKeyword).length;
 assert.equal(anyCount(before),anyCount(after));
 report.push({file,explicitAnyBefore:anyCount(before),explicitAnyAfter:anyCount(after),signaturesUnchanged:signatures,sha256:crypto.createHash('sha256').update(fs.readFileSync(afterTree+'/src/compiler/'+file)).digest('hex')});
 if(file==='core.ts') {
  const find=s=>nodes(s,n=>ts.isFunctionDeclaration(n)&&n.name?.text==='sameMap'&&n.body)[0];
  assert.equal(find(before).getText(before),find(after).getText(after));
  const cast=s=>nodes(find(s),ts.isAsExpression).find(n=>n.getText(s)==='array.slice(0, i) as unknown[] as U[]');
  const loc=s=>{const p=s.getLineAndCharacterOfPosition(cast(s).getStart(s));return {line:p.line+1,column:p.character+1};};
  report.push({sameMapUnchanged:true,before:loc(before),after:loc(after)});
 }
 let declaredResults=0;
 for(const call of nodes(after,n=>ts.isCallExpression(n)&&n.expression.getText(after)==='temporaryExplicitIterator')) {
  for(const i of [0,2]) {const arg=call.arguments[i];if(arg&&ts.isArrowFunction(arg)){assert(arg.type, file+': result callback must declare a type');declaredResults++;}}
 }
 for(const method of nodes(after,n=>ts.isMethodDeclaration(n)&&['next','return','throw'].includes(n.name.getText(after)))) {
  if(method.parent.getText(after).includes('Generator is already running')){assert(method.type);declaredResults++;}
 }
 report.push({file,declaredResultCallbacksAndMethods:declaredResults});
}

console.log(JSON.stringify(report));
