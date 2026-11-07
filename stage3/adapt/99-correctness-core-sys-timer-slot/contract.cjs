'use strict';
const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),assert=require('node:assert/strict');
const adapter=require('./adapt.cjs');
assert.equal(process.argv.length,3,'usage: contract.cjs HEAD1');
const root=fs.mkdtempSync(path.join(os.tmpdir(),'sys-timer-slot-')),before=path.resolve(process.argv[2]);
function tree(name){const dest=path.join(root,name);fs.mkdirSync(path.join(dest,'src/compiler'),{recursive:true});fs.copyFileSync(path.join(before,'src/compiler/sys.ts'),path.join(dest,'src/compiler/sys.ts'));return dest;}
const file=dest=>path.join(dest,'src/compiler/sys.ts');
const snapshot=dest=>[adapter.digest(fs.readFileSync(file(dest))),String(fs.statSync(file(dest),{bigint:true}).mtimeNs)];
function replace(dest,from,to){const raw=fs.readFileSync(file(dest),'utf8'),text=raw.replace(/\r\n/g,'\n');assert(text.includes(from));let result=text.replace(from,to);if(raw.includes('\r\n'))result=result.replace(/\n/g,'\r\n');fs.writeFileSync(file(dest),result);}
const refusals=[];
function refused(name,mutate){const dest=tree(name);mutate(dest);const original=snapshot(dest),write=fs.writeFileSync;let writes=0;fs.writeFileSync=function(...args){writes++;return write.apply(this,args);};try{assert.throws(()=>adapter.adapt(dest));assert.equal(writes,0);assert.deepEqual(snapshot(dest),original);}finally{fs.writeFileSync=write;}refusals.push(name);}
let initial,second;
for(const ending of ['LF','CRLF']){const dest=tree(ending);let text=fs.readFileSync(file(dest),'utf8').replace(/\r\n/g,'\n');fs.writeFileSync(file(dest),ending==='CRLF'?text.replace(/\n/g,'\r\n'):text);initial=adapter.adapt(dest);assert.equal(initial.edits,3);const after=snapshot(dest),write=fs.writeFileSync;let writes=0;fs.writeFileSync=function(...args){writes++;return write.apply(this,args);};try{second=adapter.adapt(dest);assert.equal(second.edits,0);assert.equal(writes,0);assert.deepEqual(snapshot(dest),after);}finally{fs.writeFileSync=write;}assert.equal(fs.readFileSync(file(dest),'utf8').includes('\r\n'),ending==='CRLF');}
for(const pin of adapter.owners){refused('half-owner-'+pin.owner,d=>replace(d,pin.before,pin.after));refused('drift-'+pin.owner,d=>replace(d,pin.before,pin.before.replace('{','{ /* drift */')));refused('duplicate-'+pin.owner,d=>replace(d,pin.before,pin.before+'\n'+pin.before));refused('moved-'+pin.owner,d=>replace(d,pin.before,'namespace moved {\n'+pin.before+'\n}'));}
refused('half-slot-setTimeout',d=>replace(d,'setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;','setTimeout?<A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A): unknown;'));
refused('half-slot-clearTimeout',d=>replace(d,'clearTimeout?(timeoutId: any): void;','clearTimeout?(timeoutId: unknown): void;'));
refused('parse-error',d=>fs.appendFileSync(file(d),'\nfunction {'));
const unrelated=tree('unrelated'),addition='\r\nfunction unrelatedTimerProof() { return 1; }\r\n';fs.appendFileSync(file(unrelated),addition);adapter.adapt(unrelated);assert(fs.readFileSync(file(unrelated),'utf8').endsWith(addition));
fs.rmSync(root,{recursive:true});
console.log(JSON.stringify({initial,second,noWriteReplay:true,LF:true,CRLF:true,refusals,unrelatedEditAccepted:true},null,2));
