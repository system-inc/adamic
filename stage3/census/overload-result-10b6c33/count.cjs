const fs=require('fs');
const ts=require('/workspace/overload-parent/stage3/api/node_modules/typescript');
const header=JSON.parse(fs.readFileSync('/workspace/overload-build-parent/latent.jsonl','utf8').split('\n')[0]);
let total=0,attempted=0,deferred=0,files=[];
for(const file of header.sources){
 const sf=ts.createSourceFile(file,fs.readFileSync(file,'utf8'),ts.ScriptTarget.Latest,true);
 let count=0,attempt=0;
 function nested(node,generic){
  total++;count++;
  const d=generic||Boolean(node.typeParameters?.length);
  if(d)deferred++;else{attempted++;attempt++;}
  
 }
 function walk(node,generic){ts.forEachChild(node,c=>{if(ts.isFunctionDeclaration(c))nested(c,generic);else walk(c,generic);});}
 for(const s of sf.statements){if(ts.isFunctionDeclaration(s))nested(s,false);else{total++;count++;attempted++;attempt++;}}
 files.push({file,units:count,attempted:attempt});
}
console.log(JSON.stringify({total,attempted,deferred,files},null,2));
