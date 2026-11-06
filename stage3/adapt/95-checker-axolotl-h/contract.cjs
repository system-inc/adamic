#!/usr/bin/env node
'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const {adapt}=require('./adapt.cjs'),inventory=require('./sites.json');
const [tree,out]=process.argv.slice(2);assert(tree&&out,'usage: contract.cjs BEFORE_95H NEW_OUTPUT');
fs.mkdirSync(out,{recursive:false});
// Copy the actual prefix tree, including generated diagnostics and configuration.
fs.cpSync(path.join(tree,'src'),path.join(out,'src'),{recursive:true});
const checkerFile=path.join(out,'src/compiler/checker.ts'),typesFile=path.join(out,'src/compiler/types.ts');
const raw=fs.readFileSync(checkerFile,'utf8'),rawTypes=fs.readFileSync(typesFile,'utf8');
const row=inventory.literals.find(row=>row.ownerAfterSha256),members=row.members.filter(member=>typeof member.initializer==='object');
const originalWrite=fs.writeFileSync;let writes=0;
fs.writeFileSync=(...args)=>{writes++;return originalWrite(...args);};
const first=adapt(path.resolve(out));assert.equal(first.files,2);assert.equal(first.declarations,4);assert.equal(first.rewrites,2);assert.equal(writes,2);
const after=fs.readFileSync(checkerFile,'utf8'),afterTypes=fs.readFileSync(typesFile,'utf8');
const stamp=fs.statSync(checkerFile,{bigint:true}).mtimeNs;writes=0;
const second=adapt(path.resolve(out));assert.equal(second.files,0);assert.equal(second.declarations,0);assert.equal(second.rewrites,0);assert.equal(writes,0);
assert.equal(fs.statSync(checkerFile,{bigint:true}).mtimeNs,stamp);assert.equal(fs.readFileSync(checkerFile,'utf8'),after);
assert.equal(adapt(path.resolve(out),true).files,0);assert.equal(writes,0);
const refusals=[];
function mutant(name,text,expected){
 const file=path.join(out,name+'-mutant-checker.ts');originalWrite(file,text);originalWrite(checkerFile,text);originalWrite(typesFile,rawTypes);writes=0;
 let error;try{adapt(path.resolve(out));}catch(e){error=e;}
 assert(error,name+' must refuse');assert.match(error.message,expected);assert.equal(writes,0);assert.equal(fs.readFileSync(checkerFile,'utf8'),text);assert.equal(fs.readFileSync(typesFile,'utf8'),rawTypes);
 refusals.push({name,file,message:error.message});
}
for(const member of members)mutant('half-'+member.name,raw.replace(member.initializer.before,member.initializer.after),/half-adapted checker owner/);
mutant('third-member',after.replace('getCurrentDirectory: () => host.getCurrentDirectory()','getCurrentDirectory: () => "foreign"'),/whole literal shape/);
mutant('wrong-receiver',raw.replace(members[0].initializer.before,'maybeBind(otherHost, host.getSymlinkCache)'),/folded initializer shape/);
assert(raw.includes('    maybeBind,'),'maybeBind import');
mutant('wrong-callee-symbol',raw.replace('    maybeBind,','    some as maybeBind,'),/core maybeBind symbol/);
originalWrite(checkerFile,after);originalWrite(typesFile,afterTypes);
console.log(JSON.stringify({first,second,replayNoWrites:true,refusals},null,2));
