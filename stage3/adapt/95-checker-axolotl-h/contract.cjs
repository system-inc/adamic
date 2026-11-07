#!/usr/bin/env node
'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const multimap=require('./multimap.json'),hash=text=>crypto.createHash('sha256').update(text).digest('hex');
const {adapt}=require('./adapt.cjs'),inventory=require('./sites.json');
const [tree,out]=process.argv.slice(2);assert(tree&&out,'usage: contract.cjs BEFORE_95H NEW_OUTPUT');
fs.mkdirSync(out,{recursive:false});fs.cpSync(path.join(tree,'src'),path.join(out,'src'),{recursive:true});
const relative=['src/compiler/types.ts','src/compiler/checker.ts','src/compiler/program.ts'],raw=new Map(relative.map(name=>[name,fs.readFileSync(path.join(out,name),'utf8')]));
const originalWrite=fs.writeFileSync;let writes=0;fs.writeFileSync=(...args)=>{writes++;return originalWrite(...args);};
const seed=map=>{for(const [name,text] of map)originalWrite(path.join(out,name),text);writes=0;};
const first=adapt(path.resolve(out));assert.equal(first.files,3);assert.equal(first.declarations,4);assert.equal(first.rewrites,6);assert.equal(first.importsRemoved,1);assert.equal(writes,3);
const after=new Map(relative.map(name=>[name,fs.readFileSync(path.join(out,name),'utf8')])),stamps=new Map(relative.map(name=>[name,fs.statSync(path.join(out,name),{bigint:true}).mtimeNs]));writes=0;
const second=adapt(path.resolve(out));assert.equal(second.files,0);assert.equal(second.rewrites,0);assert.equal(second.importsRemoved,0);assert.equal(writes,0);
assert.equal(adapt(path.resolve(out),true).files,0);assert.equal(writes,0);
for(const name of relative){assert.equal(fs.readFileSync(path.join(out,name),'utf8'),after.get(name));assert.equal(fs.statSync(path.join(out,name),{bigint:true}).mtimeNs,stamps.get(name));}
const checkerRow=inventory.literals[0],checkerSites=checkerRow.members.filter(m=>typeof m.initializer==='object'),programSites=inventory.literals.filter(row=>row.file==='src/compiler/program.ts').flatMap(row=>row.members.filter(m=>typeof m.initializer==='object').map(member=>({row,member})));
const refusals=[];
function mutant(name,base,relativeName,text,expected){
 assert(expected instanceof RegExp,'missing refusal expectation: '+name);
 const state=new Map(base);assert.notEqual(text,state.get(relativeName),name+' mutates');state.set(relativeName,text);seed(state);
 const artifact=path.join(out,name+'-mutant-'+path.basename(relativeName));originalWrite(artifact,text);writes=0;
 const beforeSha256=hash(fs.readFileSync(path.join(out,relativeName)));
 let error;try{adapt(path.resolve(out));}catch(e){error=e;}
 assert(error,name+' must refuse');assert.match(error.message,expected);assert.equal(writes,0,'no write on refusal');for(const [file,value] of state)assert.equal(fs.readFileSync(path.join(out,file),'utf8'),value);
 const afterSha256=hash(fs.readFileSync(path.join(out,relativeName)));assert.equal(afterSha256,beforeSha256,'refusal file hash');
 refusals.push({name,artifact,message:error.message,beforeSha256,afterSha256,writes});console.log(JSON.stringify({mutant:name,refused:true,message:error.message,beforeSha256,afterSha256,writes}));
}
const checker='src/compiler/checker.ts',program='src/compiler/program.ts';
const gap=multimap.import,importBefore=gap.previous+gap.before+gap.next,importAfter=gap.previous+gap.after+gap.next;
mutant('multimap-half-owner',raw,program,raw.get(program).replace(importBefore,importAfter).replace(multimap.calls[0].before,multimap.calls[0].after),/^half-adapted MultiMap owner$/);
let multiMapCalls=raw.get(program).replace(importBefore,importAfter);for(const call of multimap.calls)multiMapCalls=multiMapCalls.replace(call.before,call.after);
mutant('multimap-half-calls',raw,program,multiMapCalls,/^half-adapted MultiMap calls/);
// Keep the helper bound elsewhere in the same import so the pinned gap,
// rather than the earlier free-helper symbol check, diagnoses this mutant.
mutant('multimap-import-left',after,program,after.get(program).replace(importAfter,importBefore).replace('    maybeBind,','    maybeBind,\r\n    multiMapAddValue,'),/^pinned MultiMap import gap/);
for(const m of checkerSites)mutant('checker-half-'+m.name,raw,checker,raw.get(checker).replace(m.initializer.before,m.initializer.after),/half-adapted owner/);
for(let mask=1;mask<15;mask++){
 let text=raw.get(program);for(let i=0;i<4;i++)if(mask&(1<<i))text=text.replace(programSites[i].member.initializer.before,programSites[i].member.initializer.after);
 mutant('program-half-'+mask.toString(2).padStart(4,'0'),raw,program,text,/half-adapted owner: createProgram/);
}
let callsOnly=raw.get(checker);for(const m of checkerSites)callsOnly=callsOnly.replace(m.initializer.before,m.initializer.after);
mutant('checker-import-left',raw,checker,callsOnly,/half-adapted checker import/);
mutant('checker-import-removed-early',raw,checker,raw.get(checker).replace(inventory.importRemoval.before,inventory.importRemoval.after),/half-adapted checker import/);
mutant('checker-third-member',after,checker,after.get(checker).replace('getCurrentDirectory: () => host.getCurrentDirectory()','getCurrentDirectory: () => "foreign"'),/whole literal shape/);
mutant('program-third-member',after,program,after.get(program).replace('useCaseSensitiveFileNames: () => host.useCaseSensitiveFileNames()','useCaseSensitiveFileNames: () => true'),/writer owner shape: createProgram/);
mutant('emit-host-third-member',after,program,after.get(program).replace('getBuildInfo: () => program.getBuildInfo?.()','getBuildInfo: () => undefined'),/writer owner shape: createProgram/);
mutant('program-callee-symbol',raw,program,raw.get(program).replace('    maybeBind,','    some as maybeBind,'),/core maybeBind symbol/);
for(const readyFile of [checker,program]){
 const state=new Map(raw);state.set(readyFile,after.get(readyFile));seed(state);const result=adapt(path.resolve(out));assert.equal(result.rewrites,readyFile===checker?4:2);
 for(const [name,text] of after)assert.equal(fs.readFileSync(path.join(out,name),'utf8'),text);
}
for(const [name,reason] of [['multimap-half-owner','half-adapted MultiMap owner'],['multimap-half-calls','half-adapted MultiMap calls'],['multimap-import-left','pinned MultiMap import gap']])assert(refusals.some(row=>row.name===name&&row.message.split('\n')[0]===reason),'missing MultiMap expectation: '+name);
seed(after);console.log(JSON.stringify({first,second,replayNoWrites:true,refusals,independentOwnerStatesAccepted:true,programImportUntouched:true},null,2));
