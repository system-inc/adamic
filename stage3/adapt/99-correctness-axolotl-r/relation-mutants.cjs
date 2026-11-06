'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),child=require('node:child_process');
const [tree,binary,evidence]=process.argv.slice(2);assert(tree&&binary&&evidence,'usage: node relation-mutants.cjs TREE BUILT_ADAMIC EVIDENCE');fs.mkdirSync(evidence,{recursive:true});
const original=fs.readFileSync(path.join(tree,'src/compiler/core.ts'),'utf8'),needle='emptyArray as readonly T[]';assert.equal(original.split(needle).length,2);
const cases=[['readonly-tuple-mutant','emptyArray as readonly []'],['branded-downcast-mutant','emptyArray as never[] & { readonly __probeBrand: true }'],['mutable-view-mutant','emptyArray as T[]']];
const rows=[];
for(const [name,replacement] of cases){
 const target=path.join(evidence,name),compiler=path.join(target,'src/compiler');fs.mkdirSync(compiler,{recursive:true});fs.writeFileSync(path.join(compiler,'core.ts'),original.replace(needle,replacement));
 const probes=path.join(target,'probes'),result=child.spawnSync(process.execPath,[path.join(__dirname,'relation-contract.cjs'),target,path.resolve(binary)],{encoding:'utf8',env:{...process.env,ADAMIC_RELATION_EVIDENCE:probes},maxBuffer:64*1024*1024});
 fs.writeFileSync(path.join(target,'stdout.log'),result.stdout);fs.writeFileSync(path.join(target,'stderr.log'),result.stderr);assert.ifError(result.error);
 assert.equal(result.status,1,name+' must turn the relation contract red');assert(result.stderr.includes('Adamic upcast relation rejected'),name+' failed outside the compiler relation check');
 const detail=JSON.parse(fs.readFileSync(path.join(probes,'result.json'),'utf8'));
 if(name==='mutable-view-mutant'){
  assert(detail.results.every(r=>r.commands.find(c=>c.command[1]==='types').exit===0),'mutable control must pass the checker');
  assert(detail.results.every(r=>r.commands.find(c=>c.command[1]==='c').exit===1&&r.commands.find(c=>c.command[1]==='c').stderr.includes('can be written')),'mutable control must fail Adamic beyond TypeScript assignability');
 }
 rows.push({name,replacement,exit:result.status,detail});
}
console.log(JSON.stringify({allRelationMutantsRed:true,ownerHashChecksNotInvoked:true,rows},null,2));
