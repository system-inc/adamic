'use strict';
const fs=require('fs'),path=require('path'),os=require('os'),assert=require('assert/strict'),cp=require('child_process');
const adapter=require('./adapt.cjs'),{owners}=adapter;
const tree=process.argv[2],root=fs.mkdtempSync(path.join(os.tmpdir(),'clear-contract-'));fs.mkdirSync(root+'/src/compiler',{recursive:true});
const file=root+'/src/compiler/core.ts';
const snapshot=cp.spawnSync('git',['-C',tree,'show','refs/stage3/99-correctness-axolotl-s-before:src/compiler/core.ts'],{encoding:'utf8'});
console.log('snapshot_child_exit='+snapshot.status);if(snapshot.status!==0){console.error(snapshot.stderr);process.exit(snapshot.status);}const seed=snapshot.stdout;fs.writeFileSync(file,seed);
try {
 const r=cp.spawnSync('node',[__dirname+'/contract.cjs',root],{encoding:'utf8',env:process.env});
 console.log('original_contract_child_exit='+r.status);console.log(r.stdout);console.error(r.stderr);if(r.status!==0)process.exit(r.status);
 let text=seed.toString().replace(/\r\n/g,'\n');for(const p of owners)text=text.replace(p.before,p.after);
 const clear=owners.find(p=>p.owner==='clear'),queue=owners.find(p=>p.owner==='createQueue');
 const rows=[];
 for(const [name,raw,edits] of [['old-current',text.replace(clear.after,clear.previousAfter),1],['old-queue',text.replace(clear.after,clear.previousAfter).replace(queue.after,queue.previousAfter),2]]){
  fs.writeFileSync(file,raw);assert.equal(adapter.adapt(root).edits,edits);assert.equal(fs.readFileSync(file,'utf8'),text);
  const stat=fs.statSync(file,{bigint:true});assert.equal(adapter.adapt(root).edits,0);assert.equal(fs.statSync(file,{bigint:true}).mtimeNs,stat.mtimeNs);rows.push({name,edits,replayWrites:0});
 }
 const mutants=[text.replace(clear.after,clear.after.replace('splice(0)','length = 0')),text.replace(clear.after,clear.previousAfter).replace(queue.after,queue.before),text.replace(queue.after,queue.previousAfter)];
 for(const raw of mutants){fs.writeFileSync(file,raw);assert.throws(()=>adapter.adapt(root));assert.equal(fs.readFileSync(file,'utf8'),raw);}
 assert.deepEqual(owners,JSON.parse(fs.readFileSync(__dirname+'/owners.json','utf8')));
 console.log(JSON.stringify({migrations:rows,additionalNoWriteRefusals:mutants.length,ownersMatch:true},null,2));
} finally {fs.rmSync(root,{recursive:true,force:true});}
