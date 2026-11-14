'use strict';
const fs=require('fs'),path=require('path'),crypto=require('crypto'),assert=require('assert/strict');
const s='/workspace/scratch/diag-read-view-3',root=s+'/replay-tree';
const files=['resolutionCache.ts','program.ts','utilitiesPublic.ts','utilities.ts','types.ts'];
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
fs.mkdirSync(root+'/src/compiler',{recursive:true});
for(const name of files)fs.copyFileSync('/workspace/scratch/diag-read-view-2/applied-tree/src/compiler/'+name,root+'/src/compiler/'+name);
const snapshot=()=>Object.fromEntries(files.map(name=>[name,hash(fs.readFileSync(root+'/src/compiler/'+name))]));
const before=snapshot(),write=fs.writeFileSync;
let writes=0;fs.writeFileSync=function(...args){writes++;return write.apply(this,args)};
let result;try{result=require(s+'/repo/stage3/adapt/99-correctness-binding-views/adapt.cjs').adapt(root)}finally{fs.writeFileSync=write}
assert.equal(writes,0);assert.equal(result.files,0);assert.deepEqual(snapshot(),before);
const report={writes,result,sourceUnchanged:true,files:before};
fs.writeFileSync(s+'/raw/replay.json',JSON.stringify(report,null,2)+'\n');
console.log(JSON.stringify(report,null,2));
