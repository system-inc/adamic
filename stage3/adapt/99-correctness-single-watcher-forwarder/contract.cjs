'use strict';
const fs=require('fs'),path=require('path'),os=require('os'),assert=require('assert/strict');
const adapter=require('./adapt.cjs'),before=path.resolve(process.argv[2]),root=fs.mkdtempSync(path.join(os.tmpdir(),'watcher-contract-'));
const name=dest=>path.join(dest,'src/compiler/sys.ts');
function tree(label){const dest=path.join(root,label);fs.mkdirSync(path.dirname(name(dest)),{recursive:true});fs.copyFileSync(name(before),name(dest));return dest;}
function snapshot(dest){return [adapter.digest(fs.readFileSync(name(dest))),String(fs.statSync(name(dest),{bigint:true}).mtimeNs)];}
const first=tree('first'),initial=adapter.adapt(first),after=snapshot(first),write=fs.writeFileSync;let writes=0;
function watch(fn){writes=0;fs.writeFileSync=function(...args){writes++;return write.apply(this,args);};try{return fn();}finally{fs.writeFileSync=write;}}
const second=watch(()=>adapter.adapt(first));assert.equal(writes,0);assert.equal(second.edits,0);assert.deepEqual(snapshot(first),after);
const refusals=[];
function refuse(label,change){const dest=tree(label);let text=fs.readFileSync(name(dest),'utf8').replace(/\r\n/g,'\n');fs.writeFileSync(name(dest),change(text));const snap=snapshot(dest);watch(()=>assert.throws(()=>adapter.adapt(dest)));assert.equal(writes,0);assert.deepEqual(snapshot(dest),snap);refusals.push(label);}
for(const [i,pin] of adapter.owners.entries()){
 refuse('drift-'+i,t=>t.replace(pin.before,pin.before.replace('(', '(/* drift */')));
 refuse('half-adapted-'+i,t=>t.replace(pin.before,pin.after));
 refuse('duplicate-'+i,t=>t+'\n'+pin.before);
 refuse('moved-'+i,t=>t.replace(pin.before,'namespace moved {\n'+pin.before+'\n}'));
}
refuse('half-helper-parameter',t=>t.replace('    createWatcher: (callback: T) => FileWatcher,','    createWatcher: (callback: T) => FileWatcher,\n    createForwarder: (getCallbacks: () => T[] | undefined) => T,'));
refuse('parse-error',t=>t+'\nfunction {');
for(const eol of ['\n','\r\n']){const dest=tree('eol-'+eol.length),extra=eol+'function unrelatedWatcherProof() { return 99; }'+eol;fs.writeFileSync(name(dest),fs.readFileSync(name(dest),'utf8').replace(/\r\n/g,'\n').replace(/\n/g,eol)+extra);adapter.adapt(dest);assert(fs.readFileSync(name(dest),'utf8').endsWith(extra));const snap=snapshot(dest);watch(()=>adapter.adapt(dest));assert.equal(writes,0);assert.deepEqual(snapshot(dest),snap);}
console.log(JSON.stringify({initial,second,noWriteReplay:true,refusals,unrelatedAndLFCRLF:true},null,2));
