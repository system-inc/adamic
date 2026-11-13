'use strict';
const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),assert=require('node:assert/strict');
const adapter=require('./adapt.cjs');
const baseline=path.resolve(process.argv[2]);
const root=fs.mkdtempSync(path.join(os.tmpdir(),'timer-label-contract-'));
const files=[...new Set(adapter.owners.map(p=>p.file))];
const original=new Map(files.map(file=>[file,fs.readFileSync(path.join(baseline,file),'utf8').replace(/\r\n/g,'\n')]));
function seed(ending='\n') { for(const [file,text]of original){const name=path.join(root,file);fs.mkdirSync(path.dirname(name),{recursive:true});fs.writeFileSync(name,text.replace(/\n/g,ending));} }
function snapshot(){return files.map(file=>{const name=path.join(root,file);return [fs.readFileSync(name,'utf8'),String(fs.statSync(name,{bigint:true}).mtimeNs)];});}
function noWrites(action){const write=fs.writeFileSync;let writes=0;fs.writeFileSync=(...args)=>{writes++;return write(...args);};try{action();assert.equal(writes,0);}finally{fs.writeFileSync=write;}}
const cases=[];
function refused(name,mutate,adapted=false){seed();if(adapted)adapter.adapt(root);mutate();const before=snapshot();noWrites(()=>assert.throws(()=>adapter.adapt(root),undefined,name));assert.deepEqual(snapshot(),before,name);cases.push(name);}
function replace(pin,before,after){const name=path.join(root,pin.file),text=fs.readFileSync(name,'utf8');assert(text.includes(before));fs.writeFileSync(name,text.replace(before,after));}
try {
 if(process.argv[3] === '--restore-slot') {
  seed();adapter.adapt(root);const pin=adapter.owners.find(p=>p.owner==='System');
  replace(pin,'setTimeout?(callback: (label?: unknown) => void, ms: number, label?: unknown): unknown;','setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;');
  console.error('case: restored-variadic-slot');adapter.adapt(root);throw Error('restored variadic slot escaped');
 }
 for(const ending of ['\n','\r\n']){seed(ending);assert.equal(adapter.adapt(root).edits,adapter.owners.length);const before=snapshot();noWrites(()=>assert.equal(adapter.adapt(root).edits,0));assert.deepEqual(snapshot(),before);}
 for(const pin of adapter.owners){
  refused('half-owner-'+pin.owner,()=>replace(pin,pin.before,pin.after));
  refused('drift-'+pin.owner,()=>replace(pin,pin.before,pin.before.replace('{','{ /* drift */')));
  refused('duplicate-'+pin.owner,()=>replace(pin,pin.before,pin.before+'\n'+pin.before));
  refused('moved-'+pin.owner,()=>replace(pin,pin.before,'namespace moved {\n'+pin.before+'\n}'));
 }
 seed();const pin=adapter.owners[0];const file=path.join(root,pin.file);fs.appendFileSync(file,'\nfunction unrelatedTimerWitness(){return 1;}\n');adapter.adapt(root);assert(fs.readFileSync(file,'utf8').endsWith('function unrelatedTimerWitness(){return 1;}\n'));
 seed('\r\n');fs.appendFileSync(file,'\nfunction mixedEndingWitness(){return 2;}\n');adapter.adapt(root);assert(fs.readFileSync(file,'utf8').endsWith('\nfunction mixedEndingWitness(){return 2;}\n'));
 console.log(JSON.stringify({mixedOutsidePreserved:true,owners:adapter.owners.length,files:files.length,noWriteReplay:true,bytesAndMtimePreserved:true,LF:true,CRLF:true,refusals:cases,unrelatedEditAccepted:true},null,2));
} finally { fs.rmSync(root,{recursive:true,force:true}); }
