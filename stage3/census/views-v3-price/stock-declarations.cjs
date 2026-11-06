// Independent stock TypeScript inventory of the full meter's declaration ownership.
const fs=require('fs'),path=require('path'),ts=require('../../api/node_modules/typescript');
if(ts.version!=='6.0.3')throw Error('TypeScript pin');
const [tree,manifest,out]=process.argv.slice(2),files=JSON.parse(fs.readFileSync(manifest,'utf8')).files;
const rows=[];
for(const name of files){
 const text=fs.readFileSync(path.join(tree,name),'utf8'),f=ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true),bytes=n=>Buffer.byteLength(text.slice(0,n));
 const where=n=>{const p=f.getLineAndCharacterOfPosition(n.getStart(f));return `${name}:${p.line+1}:${p.character+1}`};
 const row={file:name,tops:[],units:[]};
 const unit=(n,top,parent='',depth=0)=>({where:where(n),top,parent,depth,start:bytes(n.pos),end:bytes(n.end)});
 function nestedDeclarations(body,visit){if(!body)return;function walk(n){if(ts.isFunctionDeclaration(n)){visit(n);return}ts.forEachChild(n,walk)}ts.forEachChild(body,walk)}
 function nested(n,parent,top,depth){row.units.push(unit(n,top,parent,depth));nestedDeclarations(n.body,c=>nested(c,where(n),top,depth+1))}
 for(const n of f.statements){
  const top=where(n);row.tops.push(unit(n,top));
  if(ts.isFunctionDeclaration(n))nested(n,'',top,0);
  else{row.units.push(unit(n,top));nestedDeclarations(n,c=>nested(c,top,top,0))}
 }
 rows.push(row);
}
fs.writeFileSync(out,JSON.stringify(rows,null,2)+'\n');
console.log(JSON.stringify({files:rows.length,tops:rows.reduce((n,r)=>n+r.tops.length,0),units:rows.reduce((n,r)=>n+r.units.length,0)}));
