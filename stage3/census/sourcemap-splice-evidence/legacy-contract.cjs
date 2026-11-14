'use strict';
const fs=require('node:fs'),os=require('node:os'),path=require('node:path'),cp=require('node:child_process'),assert=require('node:assert/strict');
const {owners,adapt}=require('/workspace/scratch/sourcemap-splice-1/stage3/adapt/99-correctness-axolotl-s/adapt.cjs'),tree=process.argv[2],root=fs.mkdtempSync(path.join(os.tmpdir(),'sourcemap-legacy-'));fs.mkdirSync(root+'/src/compiler',{recursive:true});
const before={};for(const file of ['core.ts','sourcemap.ts'])before[file]=cp.execFileSync('git',['-C',tree,'show','refs/stage3/99-correctness-axolotl-s-before:src/compiler/'+file],{encoding:'utf8'}).replace(/\r\n/g,'\n');
const after={...before};for(const p of owners)after[p.file]=after[p.file].replace(p.before,p.after);
const clear=owners.find(p=>p.owner==='clear'),queue=owners.find(p=>p.owner==='createQueue');const rows=[];
try{for(const [name,core,edits] of [['last-99s',after['core.ts'],1],['previous-clear',after['core.ts'].replace(clear.after,clear.previousAfter),2],['previous-queue-and-clear',after['core.ts'].replace(clear.after,clear.previousAfter).replace(queue.after,queue.previousAfter),3]]){
 fs.writeFileSync(root+'/src/compiler/core.ts',core);fs.writeFileSync(root+'/src/compiler/sourcemap.ts',before['sourcemap.ts']);assert.equal(adapt(root).edits,edits);for(const file of Object.keys(after))assert.equal(fs.readFileSync(root+'/src/compiler/'+file,'utf8'),after[file]);assert.equal(adapt(root).edits,0);rows.push({name,edits,replayEdits:0});
}console.log(JSON.stringify({status:'PASS',rows},null,2));}finally{fs.rmSync(root,{recursive:true,force:true});}
