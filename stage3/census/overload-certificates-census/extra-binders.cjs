'use strict';
const fs=require('node:fs'),path=require('node:path');const ts=require(process.env.TYPESCRIPT_API);
const tree=path.resolve(process.argv[2]),callers=JSON.parse(fs.readFileSync(process.argv[3])),entry=new Set(callers.programs.entry.sources),rows=[];
for(const file of ts.sys.readDirectory(tree+'/src',['.ts','.tsx'],undefined,['**/*'])){
 const sf=ts.createSourceFile(file,fs.readFileSync(file,'utf8'),ts.ScriptTarget.Latest,true),stack=[sf];
 while(stack.length){const n=stack.pop();
  if(ts.isFunctionDeclaration(n)&&n.body&&n.name&&n.typeParameters?.length){const siblings=n.parent.statements;if(siblings)for(const o of siblings)if(ts.isFunctionDeclaration(o)&&!o.body&&o.name?.text===n.name.text&&(o.typeParameters?.length??0)<n.typeParameters.length){const loc=sf.getLineAndCharacterOfPosition(o.getStart(sf));rows.push({file:path.relative(tree,file),line:loc.line+1,column:loc.character+1,name:n.name.text,entryReach:entry.has(path.relative(tree,file)),implementationBinders:n.typeParameters.map(p=>p.getText(sf)),overloadBinders:(o.typeParameters??[]).map(p=>p.getText(sf)),overload:o.getText(sf),implementationSignature:n.getText(sf).slice(0,n.body.getStart(sf)-n.getStart(sf))});}}
  ts.forEachChild(n,c=>{stack.push(c);});
 }
}
fs.writeFileSync(process.argv[4],JSON.stringify({scope:'syntactic sibling function declarations with more implementation binders; not an acceptance census',rows},null,2)+'\n');console.log(JSON.stringify({overloads:rows.length,entryOverloads:rows.filter(r=>r.entryReach).length,owners:[...new Set(rows.map(r=>r.file+':'+r.name))]}));
