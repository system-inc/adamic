'use strict';
const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),assert=require('node:assert/strict');
const adapter=require('./adapt.cjs');
assert.equal(process.argv.length,3,'usage: contract.cjs BEFORE');
const root=fs.mkdtempSync(path.join(os.tmpdir(),'99k-contract-')),before=path.resolve(process.argv[2]),files=['core.ts','builder.ts'];
function tree(name){const dest=path.join(root,name);fs.mkdirSync(path.join(dest,'src/compiler'),{recursive:true});for(const file of files)fs.copyFileSync(path.join(before,'src/compiler',file),path.join(dest,'src/compiler',file));return dest;}
function snapshot(dest){return files.map(file=>{const p=path.join(dest,'src/compiler',file);return [file,adapter.digest(fs.readFileSync(p)),String(fs.statSync(p,{bigint:true}).mtimeNs)];});}
const first=tree('first'),initial=adapter.adapt(first),after=snapshot(first),write=fs.writeFileSync;let writes=0;
fs.writeFileSync=function(...args){writes++;return write.apply(this,args);};const second=adapter.adapt(first);assert.equal(writes,0);assert.equal(second.edits,0);assert.deepEqual(snapshot(first),after);fs.writeFileSync=write;
const refusals=[];
function refused(name,mutate){const dest=tree(name);mutate(dest);const original=snapshot(dest);writes=0;fs.writeFileSync=function(...args){writes++;return write.apply(this,args);};assert.throws(()=>adapter.adapt(dest));assert.equal(writes,0);assert.deepEqual(snapshot(dest),original);fs.writeFileSync=write;refusals.push(name);}
function replace(dest,pin,from,to){const p=path.join(dest,'src/compiler',pin.file);let text=fs.readFileSync(p,'utf8').replace(/\r\n/g,'\n');assert(text.includes(from));fs.writeFileSync(p,text.replace(from,to).replace(/\n/g,'\r\n'));}
for(const [index,pin]of adapter.owners.entries()){
 refused('half-owner-'+index,dest=>replace(dest,pin,pin.before,pin.after));
 refused('changed-owner-'+index,dest=>replace(dest,pin,pin.before,pin.before.replace('(', '(/* changed */')));
}
const corePin=adapter.owners.find(pin=>pin.owner==='sameMap');
refused('half-single-overload',dest=>replace(dest,corePin,'<T, U = T>','<T extends U, U = T>'));
refused('half-slice-cast',dest=>replace(dest,corePin,'array.slice(0, i) as unknown[] as U[]','array.slice(0, i)'));
refused('half-return-cast',dest=>replace(dest,corePin,'return array as unknown[] as U[];','return array;'));
for(const file of files)refused('mixed-file-'+file,dest=>fs.copyFileSync(path.join(first,'src/compiler',file),path.join(dest,'src/compiler',file)));
refused('parse-error',dest=>fs.appendFileSync(path.join(dest,'src/compiler/builder.ts'),'\nfunction {'));
refused('moved-owner',dest=>{const pin=adapter.owners.find(p=>p.owner==='convertOrRepopulateDiagnosticMessageChainArray');replace(dest,pin,pin.before,'namespace moved {\n'+pin.before+'\n}');});
const unrelated=tree('unrelated'),p=path.join(unrelated,'src/compiler/core.ts'),addition='\r\nfunction unrelated99k() { return 99; }\r\n';fs.appendFileSync(p,addition);adapter.adapt(unrelated);assert(fs.readFileSync(p,'utf8').endsWith(addition));const unrelatedState=snapshot(unrelated);adapter.adapt(unrelated);assert.deepEqual(snapshot(unrelated),unrelatedState);
console.log(JSON.stringify({initial,second,noWriteSecondRun:true,refusals,unrelatedEditAccepted:true},null,2));
