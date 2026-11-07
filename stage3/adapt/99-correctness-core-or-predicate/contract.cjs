'use strict';
const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const {adapt,pins}=require('./adapt.cjs'),amendment=require('./import-amendment.json');
const root=process.argv[2],output=process.argv[3];assert(root&&output,'usage: node contract.cjs BASELINE OUTPUT');
const hash=s=>crypto.createHash('sha256').update(s).digest('hex'),files=[...new Set(pins.map(p=>p.file))],results=[];
function make(eol='\r\n'){
 const dir=fs.mkdtempSync(path.join(os.tmpdir(),'or-predicate-contract-'));
 for(const file of files){const target=path.join(dir,file);fs.mkdirSync(path.dirname(target),{recursive:true});let text=fs.readFileSync(path.join(root,file),'utf8').replace(/\r\n/g,'\n');if(file===amendment.file){assert(text.includes(amendment.oldAfter));text=text.replace(amendment.oldAfter,amendment.newAfter);}fs.writeFileSync(target,text.replace(/\n/g,eol));}return dir;
}
function snapshot(dir){return files.map(f=>[f,hash(fs.readFileSync(path.join(dir,f)))]);}
function change(dir,pin,text){const target=path.join(dir,pin.file),raw=fs.readFileSync(target,'utf8'),before=pin.before.replace(/\n/g,'\r\n');assert(raw.includes(before));fs.writeFileSync(target,raw.replace(before,text.replace(/\n/g,'\r\n')));}
function reject(name,mutate){const dir=make();try{mutate(dir);const before=snapshot(dir);let writes=0;const original=fs.writeFileSync;fs.writeFileSync=function(...args){writes++;return original.apply(this,args);};let message;try{adapt(dir);}catch(error){message=error.message;}finally{fs.writeFileSync=original;}assert(message,name+' must refuse');assert.equal(writes,0,name+' wrote before refusal');assert.deepEqual(snapshot(dir),before);results.push({name,refused:true,writes,message});}finally{fs.rmSync(dir,{recursive:true,force:true});}}
const core=pins.find(p=>p.kind==='coreGroup'),owner=pins.find(p=>p.kind==='function'),importPin=pins.find(p=>p.kind==='import');
reject('partially removed predicate overload',dir=>change(dir,core,core.before.slice(core.before.indexOf('\n')+1)));
reject('only one helper added',dir=>change(dir,core,core.after.split('\n/** @internal */\n')[0]+'\n'+core.before));
reject('one rewritten owner',dir=>change(dir,owner,owner.after));
reject('one rewritten import',dir=>change(dir,importPin,importPin.after));
reject('drifted owner',dir=>change(dir,owner,owner.before.replace('or(', 'or(isExportDeclaration, ')));
reject('renamed owner',dir=>change(dir,owner,owner.before.replace(owner.path.at(-1),owner.path.at(-1)+'Changed')));
reject('changed helper body',dir=>{adapt(dir);const file=path.join(dir,core.file),s=fs.readFileSync(file,'utf8');fs.writeFileSync(file,s.replace('f1(p) || f2(p)','f1(p) && f2(p)'));});
reject('one original owner restored',dir=>{adapt(dir);const file=path.join(dir,owner.file),s=fs.readFileSync(file,'utf8');fs.writeFileSync(file,s.replace(owner.after.replace(/\n/g,'\r\n'),owner.before.replace(/\n/g,'\r\n')));});
for(const eol of ['\n','\r\n']){
 const dir=make(eol);try{
  for(const file of files)fs.appendFileSync(path.join(dir,file),eol+'// foreign text survives'+eol);
  const first=adapt(dir);assert.equal(first.calls,19);assert.equal(first.files,11);const old=snapshot(dir),mtimes=files.map(f=>fs.statSync(path.join(dir,f)).mtimeMs);
  const second=adapt(dir);assert.deepEqual(second,{files:0,owners:0,calls:0,javascriptChanging:true});assert.deepEqual(snapshot(dir),old);assert.deepEqual(files.map(f=>fs.statSync(path.join(dir,f)).mtimeMs),mtimes);
  for(const file of files){const s=fs.readFileSync(path.join(dir,file),'utf8');assert(s.endsWith(eol+'// foreign text survives'+eol));if(eol==='\r\n')assert.equal(s.replace(/\r\n/g,'').includes('\n'),false);else assert(!s.includes('\r'));}
  results.push({name:eol==='\n'?'LF apply/replay/foreign preservation':'CRLF apply/replay/foreign preservation',first,second,bytesAndMtimePreservedOnReplay:true});
 }finally{fs.rmSync(dir,{recursive:true,force:true});}
}
fs.writeFileSync(output,JSON.stringify({checks:results.length,passed:true,results},null,2)+'\n');console.log(JSON.stringify({checks:results.length,passed:true}));
