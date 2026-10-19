'use strict';
const fs=require('node:fs'),path=require('node:path'),ts=require('typescript');
const [tree,output]=process.argv.slice(2).map(x=>path.resolve(x));
function walk(dir){return fs.readdirSync(dir,{withFileTypes:true}).flatMap(e=>e.isDirectory()?walk(path.join(dir,e.name)):[path.join(dir,e.name)]);}
const rows=[];
for(const file of walk(path.join(tree,'src')).filter(f=>f.endsWith('.ts'))){
 const source=ts.createSourceFile(file,fs.readFileSync(file,'utf8'),ts.ScriptTarget.Latest,true);
 function visit(n){
  let expression;
  if(ts.isMethodDeclaration(n)&&n.name?.getText(source)==='exit')expression=n;
  if(ts.isFunctionDeclaration(n)&&n.name?.text==='exit')expression=n;
  if((ts.isPropertyDeclaration(n)||ts.isPropertyAssignment(n))&&n.name?.getText(source)==='exit'&&n.initializer)expression=n;
  if(ts.isBinaryExpression(n)&&n.operatorToken.kind===ts.SyntaxKind.EqualsToken&&ts.isPropertyAccessExpression(n.left)&&n.left.name.text==='exit')expression=n;
  if(expression)rows.push({file:path.relative(tree,file),line:source.getLineAndCharacterOfPosition(n.getStart(source)).line+1,kind:ts.SyntaxKind[n.kind],text:n.getText(source)});
  ts.forEachChild(n,visit);
 }
 visit(source);
}
fs.writeFileSync(output,JSON.stringify({scope:'All .ts under src, including harness, server, testRunner and library declarations',typescript:ts.version,implementations:rows},null,2)+'\n');
console.log(JSON.stringify({implementations:rows.length,output}));
