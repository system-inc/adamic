'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const root=process.argv[2],manifest=[];
const names=['attachFlowNodeDebugInfoWorker','attachNodeArrayDebugInfoWorker','attachDebugPrototypeIfDebug','enableDebugInfo'];
for(const file of fs.readdirSync(path.join(root,'built/local')).filter(file=>file.endsWith('.js'))){
 const p=path.join(root,'built/local',file),text=fs.readFileSync(p,'utf8'),sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true),edits=[],counts=Object.fromEntries(names.map(n=>[n,0]));
 function visit(n){
  if(ts.isFunctionDeclaration(n)&&names.some(name=>new RegExp('^'+name+'[0-9]*$').test(n.name?.text||''))){
   const name=names.find(name=>new RegExp('^'+name+'[0-9]*$').test(n.name.text));counts[name]++;
   const first=n.body.statements[0];
   let pos=n.body.getStart(sf)+1;
   if(name==='enableDebugInfo'){assert(ts.isIfStatement(first));assert(/^isDebugInfoEnabled[0-9]*$/.test(first.expression.getText(sf)));assert(ts.isReturnStatement(first.thenStatement));pos=first.end;}
   if(name==='attachDebugPrototypeIfDebug'){assert(ts.isIfStatement(first));assert(/^isDebugging[0-9]*$/.test(first.expression.getText(sf)));assert(ts.isBlock(first.thenStatement));pos=first.thenStatement.getStart(sf)+1;}
   edits.push({pos,value:`\n globalThis.__debugDisplayCount(${JSON.stringify(name)});\n`});
  }
  ts.forEachChild(n,visit);
 }
 visit(sf);if(!edits.length)continue;for(const name of names)assert(counts[name]<=1,file+' duplicate '+name);
 let after=text;for(const {pos,value}of edits.sort((a,b)=>b.pos-a.pos))after=after.slice(0,pos)+value+after.slice(pos);
 fs.writeFileSync(p,after);manifest.push({file,counts});
}
for(const file of ['_tsc.js','typescript.js','run.js'])assert(manifest.some(row=>row.file===file&&names.every(name=>row.counts[name]===1)),'missing complete family '+file);
fs.writeFileSync(path.join(root,'debug-display-2-instrumentation.json'),JSON.stringify(manifest,null,2)+'\n');
