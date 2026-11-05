'use strict';
const ts=require('typescript');
function owners(text,file='source.ts'){
 const sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
 if(sf.parseDiagnostics.length)throw Error('parse error '+file);
 const found=new Map();
 function visit(n,ancestry){
 const scope=(ts.isModuleDeclaration(n)||ts.isClassDeclaration(n)||ts.isFunctionDeclaration(n)||(ts.isMethodDeclaration(n)&&ts.isClassDeclaration(n.parent)))&&n.name;
 const chain=scope?[...ancestry,n.name.getText(sf)]:ancestry;
 if((ts.isFunctionDeclaration(n)||(ts.isMethodDeclaration(n)&&ts.isClassDeclaration(n.parent)))&&n.body&&n.name){const key=chain.join('.');if(found.has(key))throw Error('duplicate owner '+key);found.set(key,{start:n.getStart(sf),end:n.end,text:n.getText(sf),node:n,sf});}
 ts.forEachChild(n,c=>visit(c,chain));
 }visit(sf,[]);return found;
}
module.exports={owners};
if(require.main===module){const fs=require('fs'),path=require('path');const root=process.argv[2];for(const f of ['debug.ts','utilities.ts','expressionToTypeNode.ts','moduleSpecifiers.ts','transformers/declarations/diagnostics.ts']){const text=fs.readFileSync(path.join(root,'src/compiler',f),'utf8');for(const[key,o]of owners(text,f))if(/Debug\.type<|type<TypeMapper>|function hasAntecedents/.test(o.text)&&![...owners(text,f).values()].some(v=>v.start>o.start&&v.end<o.end&&/Debug\.type<|type<TypeMapper>/.test(v.text)))console.log(f,key,o.text.slice(0,o.text.indexOf('{')+1));}}
