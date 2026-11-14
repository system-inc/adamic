'use strict';
const fs=require('fs'),assert=require('assert/strict'),crypto=require('crypto');
const s='/workspace/scratch/diag-read-view-2',r=s+'/repo',ts=require(r+'/stage3/api/node_modules/typescript'),file='utilities.ts';
const path=r+'/stage3/adapt/99-correctness-binding-views/owners.json',pins=JSON.parse(fs.readFileSync(path,'utf8'));
const text=fs.readFileSync('/workspace/scratch/diag-read-view-1/main-tree/src/compiler/'+file,'utf8').replace(/\r\n/g,'\n'),sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
const names=['compareDiagnosticsSkipRelatedInformation','compareRelatedInformation','compareMessageText','getDiagnosticFilePath','getDiagnosticCode','getDiagnosticMessage'];
const hash=t=>crypto.createHash('sha256').update(t).digest('hex');
for(const name of names){
 assert(!pins.some(p=>p.owner===name));
 const nodes=sf.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text===name);assert.equal(nodes.length,1);
 const n=nodes[0],start=n.getStart(sf,true),before=text.slice(start,n.end),edits=n.parameters.filter(p=>p.type?.getText(sf)==='Diagnostic').map(p=>({start:p.type.getStart(sf)-start,end:p.type.end-start}));
 assert(edits.length);
 let after=before;for(const e of edits.sort((a,b)=>b.start-a.start))after=after.slice(0,e.start)+'DiagnosticReadView'+after.slice(e.end);
 pins.push({file,owner:name,names:[name],beforeCount:1,afterCount:1,before,after,ownerBeforeSha256:hash(before),ownerAfterSha256:hash(after)});
}
fs.writeFileSync(path,JSON.stringify(pins,null,2)+'\n');
console.log(JSON.stringify({added:names},null,2));
