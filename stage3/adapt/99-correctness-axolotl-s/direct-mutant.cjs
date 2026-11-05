'use strict';
const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const {owners}=require('./adapt.cjs');const binary=path.resolve(process.argv[2]);assert(fs.existsSync(binary));
const cases=[];
for(const name of ['filterMutate','clear']){
 const pin=owners.find(p=>p.owner===name);
 const concrete=pin.after.replace('array: unknown[]','array: number[]');
 const restored=concrete.replace(name==='clear'?'array.splice(0);':'array.splice(outIndex);',name==='clear'?'array.length = 0;':'array.length = outIndex;');
 assert.notEqual(concrete,restored);
 for(const [state,body] of [['splice',concrete],['length',restored]]){
  const file=path.join(__dirname,`${name}-mutant-${state}.ts`);
  const source=body+'\nconst values: number[] = [1, 2, 3];\n'+(name==='clear'?'clear(values);':'filterMutate(values, value => value > 1);')+'\n';
  fs.writeFileSync(file,source);
  const result=cp.spawnSync(binary,['c',file],{encoding:'utf8'});assert(!result.error);
  const log=file.replace(/\.ts$/,'.log');fs.writeFileSync(log,result.stdout+result.stderr);
  assert.equal(result.status,state==='splice'?0:1);
  if(state==='length')assert.match(result.stderr,/stage 0 can't lower assigning a field of a value yet/);
  cases.push({owner:name,state,file:path.basename(file),exit:result.status,classification:state==='splice'?'lowered':'NotYet',diagnostic:result.stderr.trim(),log:path.basename(log),sourceSha256:crypto.createHash('sha256').update(source).digest('hex')});
 }
}
console.log(JSON.stringify({status:'PASS',compiler:binary,compilerSha256:crypto.createHash('sha256').update(fs.readFileSync(binary)).digest('hex'),bodyScope:'Exact adapted bodies; filterMutate instantiated with number; clear parameter specialized to number[] to examine concrete storage. Each control changes only the truncation statement.',cases},null,2));
