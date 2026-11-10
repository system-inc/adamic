const fs=require('fs'),path=require('path');
const ts=require(process.env.CENSUS_TYPESCRIPT || '/workspace/never-call-report/stage3/api/node_modules/typescript');
const root='/workspace/never-call-adapted', result={};
for(const file of ['debug.ts','semver.ts']){
 const name=path.join(root,'src/compiler',file),source=ts.createSourceFile(name,fs.readFileSync(name,'utf8'),ts.ScriptTarget.Latest,true);
 const declarations=[],calls=[];
 const loc=n=>{const p=source.getLineAndCharacterOfPosition(n.getStart(source));return `${file}:${p.line+1}:${p.character+1}`;};
 function walk(n,owners=[]){
  const text=n.name?.getText(source);
  const declaration=(ts.isFunctionDeclaration(n)||ts.isClassDeclaration(n)||ts.isMethodDeclaration(n)||ts.isConstructorDeclaration(n)||ts.isPropertyDeclaration(n)||ts.isGetAccessorDeclaration(n)||ts.isSetAccessorDeclaration(n));
  const next=declaration?[...owners,text||'constructor']:owners;
  if(declaration){const a=source.getLineAndCharacterOfPosition(n.getStart(source)),b=source.getLineAndCharacterOfPosition(n.end);declarations.push({name:text||'constructor',owner:owners.join('.'),where:loc(n),start_line:a.line+1,end_line:b.line+1,start:n.getStart(source),end:n.end});}
  if(ts.isCallExpression(n)&&['assertIsDefined','assertEachIsDefined'].includes(n.expression.getText(source).split('.').at(-1)))calls.push({where:loc(n),owner:owners.join('.'),expression:n.getText(source)});
  ts.forEachChild(n,c=>walk(c,next));
 }
 walk(source);result[file]={declarations,calls};
}
fs.writeFileSync(process.argv[2],JSON.stringify(result,null,2)+'\n');
