'use strict';
const fs=require('fs'),path=require('path'),os=require('os'),assert=require('assert/strict');
const adapter=require('./adapt.cjs'),pins=require('./pins.json');
const beforeRoot=process.argv[2],files=[...new Set(pins.map(p=>p.file))];
const before=new Map(files.map(f=>[f,fs.readFileSync(path.join(beforeRoot,'src/compiler',f),'utf8')]));
const planned=adapter.plan(before).texts;
const scratch=fs.mkdtempSync(path.join(os.tmpdir(),'99t-contract-'));
function restore(texts){for(const[f,text]of texts){const p=path.join(scratch,'src/compiler',f);fs.mkdirSync(path.dirname(p),{recursive:true});fs.writeFileSync(p,text)}}
function snapshot(){return files.map(f=>{const p=path.join(scratch,'src/compiler',f);return{file:f,bytes:fs.readFileSync(p).toString('base64'),mtime:fs.statSync(p).mtimeMs}})}
const rows=[];
restore(before);const first=adapter.adapt(scratch);assert(first.edits>0);rows.push({case:'first apply',...first});
let saved=snapshot(),writes=0;const write=fs.writeFileSync;
fs.writeFileSync=function(...a){writes++;return Reflect.apply(write,this,a)};
try{assert.deepEqual(adapter.adapt(scratch),{files:0,edits:0,javascriptChanging:true})}finally{fs.writeFileSync=write}
assert.equal(writes,0);assert.deepEqual(snapshot(),saved);rows.push({case:'replay',writes,unchangedBytesAndMtime:true});
function refusal(label,texts){restore(texts);const saved=snapshot();let writes=0,error;fs.writeFileSync=function(...a){writes++;return Reflect.apply(write,this,a)};try{adapter.adapt(scratch)}catch(e){error=e.message}finally{fs.writeFileSync=write}assert(error,label+' accepted');assert.equal(writes,0);assert.deepEqual(snapshot(),saved);rows.push({case:label,error,writes});}
for(const pin of pins){
 for(const[state,original]of [['before',before],['after',planned]]){
  const texts=new Map(original),owner=state==='before'?pin.before:pin.after;
  texts.set(pin.file,texts.get(pin.file).replace(owner,owner.replace('{','{ /* changed owner */')));
  refusal(`${pin.owner} changed ${state}`,texts);
 }
 const mixed=new Map(before);mixed.set(pin.file,mixed.get(pin.file).replace(pin.before,pin.after));refusal(`${pin.owner} mixed tree`,mixed);
 const partial=new Map(before);const changed=pin.before.replace(pin.exactSites[0],pin.exactSites[0]+' /* partial */');partial.set(pin.file,partial.get(pin.file).replace(pin.before,changed));refusal(`${pin.owner} partial owner`,partial);
}
for(const[state,texts]of [['before',before],['after',planned]]){const foreign=new Map([...texts].map(([f,t])=>[f,t+'\r\n// Unrelated foreign edit.\r\n']));restore(foreign);adapter.adapt(scratch);for(const f of files)assert(fs.readFileSync(path.join(scratch,'src/compiler',f),'utf8').endsWith('// Unrelated foreign edit.\r\n'));rows.push({case:'foreign '+state,status:'accepted'})}
console.log(JSON.stringify({status:'pass',rows},null,2));
