'use strict';
const fs=require('node:fs'),os=require('node:os'),path=require('node:path'),assert=require('node:assert/strict');
const adapter=require('/workspace/scratch/sourcefile-like-position-build/stage3/adapt/99-correctness-sourcefile-like-position/adapt.cjs');
const before=fs.readFileSync('/workspace/scratch/sourcefile-like-position-main-tree-2/src/compiler/types.ts','utf8');
const after=fs.readFileSync('/workspace/scratch/sourcefile-like-position-trial-tree/src/compiler/types.ts','utf8');
const tree=fs.mkdtempSync(path.join(os.tmpdir(),'position-contract-')),file=path.join(tree,'src/compiler/types.ts');
fs.mkdirSync(path.dirname(file),{recursive:true});
fs.writeFileSync(file,before);assert.equal(adapter.adapt(tree).edits,1);assert.equal(fs.readFileSync(file,'utf8'),after);
const time=fs.statSync(file).mtimeMs;assert.equal(adapter.adapt(tree).edits,0);assert.equal(fs.statSync(file).mtimeMs,time);
assert.equal(adapter.rewrite(tree,true).edits,1);assert.equal(fs.readFileSync(file,'utf8'),before);
const controls=[];
for(const [name,text] of [['owner-drift',before.replace('export interface SourceFileLike {','export interface SourceFileLike {\n    drift?: string;')],['duplicate-owner',before+'\n'+adapter.owners[0].before],['parse-error',before+'\ninterface {']]){
 fs.writeFileSync(file,text);assert.throws(()=>adapter.adapt(tree));assert.equal(fs.readFileSync(file,'utf8'),text);controls.push(name);
}
fs.writeFileSync(file,before+'\r\n// outside owner\r\n');assert.equal(adapter.adapt(tree).edits,1);assert.equal(fs.readFileSync(file,'utf8'),after+'\r\n// outside owner\r\n');
fs.rmSync(tree,{recursive:true});
console.log(JSON.stringify({status:'pass',exactCandidate:true,idempotent:true,restoringRoundTrip:true,noWriteControls:controls,outsideOwnerEditAccepted:true},null,2));
