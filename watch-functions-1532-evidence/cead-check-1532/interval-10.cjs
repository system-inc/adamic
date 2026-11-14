'use strict';
const fs=require('node:fs'),assert=require('node:assert/strict');
const s='/tmp/watch-file-check-cead',main=require(s+'/baseline-typescript.cjs'),after=require(s+'/after-tree/built/local/typescript.js');
const dir=s+'/interval-10-fixture';fs.mkdirSync(dir);const file=dir+'/file.ts';fs.writeFileSync(file,'export const x = 10;\n');
try{
 function exercise(ts){const rows=[];for(const [kind,target,label] of [[ts.WatchFileKind.PriorityPollingInterval,file,'polling file'],[ts.WatchFileKind.UseFsEvents,file,'fsEvents file'],[ts.WatchFileKind.PriorityPollingInterval,dir,'polling directory']]){const watcher=ts.sys.watchFile(target,()=>{},10,{watchFile:kind});watcher.close();rows.push({case:label,interval:10,result:'registered and closed'});}return rows;}
 const a=exercise(main),b=exercise(after);assert.deepEqual(b,a);console.log(JSON.stringify({main:a,after:b,byteIdentical:JSON.stringify(a)===JSON.stringify(b)},null,2));
}finally{fs.rmSync(dir,{recursive:true,force:true});}
