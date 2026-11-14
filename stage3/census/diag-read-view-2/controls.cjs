'use strict';
const fs=require('fs'),path=require('path'),assert=require('assert/strict'),crypto=require('crypto');
const s='/workspace/scratch/diag-read-view-2',r=s+'/repo',d=r+'/stage3/adapt/99-correctness-binding-views';
const {adapt,rewrite}=require(d+'/adapt.cjs');
const files=['resolutionCache.ts','program.ts','utilitiesPublic.ts','utilities.ts','types.ts'];
const sha=x=>crypto.createHash('sha256').update(x).digest('hex');
function copy(label,side='main'){const tree=s+'/controls/'+label;fs.mkdirSync(tree+'/src/compiler',{recursive:true});for(const f of files)fs.copyFileSync(s+'/'+side+'-tree/src/compiler/'+f,tree+'/src/compiler/'+f);return tree;}
function snap(tree){return files.map(f=>sha(fs.readFileSync(tree+'/src/compiler/'+f)));}
function measured(fn){const original=fs.writeFileSync;let writes=0;fs.writeFileSync=function(...args){writes++;return original.apply(this,args)};try{return {result:fn(),writes};}finally{fs.writeFileSync=original}}
const results={};
let tree=copy('positive');results.apply=measured(()=>adapt(tree));assert.equal(results.apply.writes,4);
results.replay=measured(()=>adapt(tree));assert.equal(results.replay.writes,0);
results.mutant=measured(()=>rewrite(tree,true));assert.equal(results.mutant.writes,5);
for(const file of files.filter(f=>f!=='resolutionCache.ts'))assert.equal(fs.readFileSync(tree+'/src/compiler/'+file,'utf8'),fs.readFileSync(s+'/main-tree/src/compiler/'+file,'utf8'));
for(const kind of ['drift','duplicate','mixed','half']){
 tree=copy(kind);const name=tree+'/src/compiler/types.ts';let text=fs.readFileSync(name,'utf8');
 if(kind==='half'){adapt(tree);fs.copyFileSync(s+'/main-tree/src/compiler/utilities.ts',tree+'/src/compiler/utilities.ts');}
 if(kind==='drift')fs.writeFileSync(name,text.replace('export interface DiagnosticWithLocation extends Diagnostic {','export interface DiagnosticWithLocation extends Diagnostic { /* drift */'));
 if(kind==='duplicate')fs.appendFileSync(name,'\nexport interface DiagnosticWithLocation extends Diagnostic {}\n');
 if(kind==='mixed')fs.writeFileSync(name,text.replace(/\r?\n/,'\n').replace(/\r?\n/,'\r\n')+'\n');
 const before=snap(tree);let error,writes;
 const original=fs.writeFileSync;writes=0;fs.writeFileSync=function(...args){writes++;return original.apply(this,args)};
 try{adapt(tree);}catch(e){error=e.message;}finally{fs.writeFileSync=original;}
 assert(error,kind);assert.equal(writes,0,kind);assert.deepEqual(snap(tree),before,kind);results[kind]={error,writes};
}
for(const mode of ['lf','crlf']){
 tree=copy(mode);for(const f of files){const n=tree+'/src/compiler/'+f;let t=fs.readFileSync(n,'utf8').replace(/\r\n/g,'\n');fs.writeFileSync(n,mode==='crlf'?t.replace(/\n/g,'\r\n'):t)}
 results[mode]=measured(()=>adapt(tree));assert.equal(measured(()=>adapt(tree)).writes,0);
}
console.log(JSON.stringify(results,null,2));
