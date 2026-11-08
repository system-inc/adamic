const fs=require('fs'),path=require('path'),ts=require('typescript');
const tree=process.argv[2], rows=[];
for(const file of ['utilitiesPublic.ts','core.ts']){
 const full=path.join(tree,'src/compiler',file),raw=fs.readFileSync(full,'utf8'),sf=ts.createSourceFile(full,raw,ts.ScriptTarget.Latest,true);
 for(const n of sf.statements){
  if(!ts.isFunctionDeclaration(n))continue;
  const p=sf.getLineAndCharacterOfPosition(n.getStart(sf));
  rows.push({file:full,name:n.name?.text,start:Buffer.byteLength(raw.slice(0,n.getStart(sf))),end:Buffer.byteLength(raw.slice(0,n.end)),where:full+':'+(p.line+1)+':'+(p.character+1),has_body:!!n.body,text:['findAncestor','orPredicate2','orPredicate3','or'].includes(n.name?.text)?n.getText(sf):undefined});
 }
}
console.log(JSON.stringify(rows,null,2));
