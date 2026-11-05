'use strict';
const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),assert=require('node:assert/strict');
const adapter=require('./adapt.cjs');
const source=process.argv[2];assert(source,'usage: node contract.cjs BASELINE');
const files=[...new Set(adapter.pins.map(p=>p.file))],original=new Map(files.map(f=>[f,fs.readFileSync(path.join(source,'src/compiler',f),'utf8')]));
const results=[];
function fixture(texts,action){const dir=fs.mkdtempSync(path.join(os.tmpdir(),'99u-contract-'));for(const [f,t]of texts){const name=path.join(dir,'src/compiler',f);fs.mkdirSync(path.dirname(name),{recursive:true});fs.writeFileSync(name,t);}return action(dir);}
function observeWrites(action){const write=fs.writeFileSync;let writes=0;fs.writeFileSync=function(...args){writes++;return Reflect.apply(write,this,args);};try{return {value:action(),writes};}finally{fs.writeFileSync=write;}}
function refused(name,texts){fixture(texts,dir=>{const before=new Map(files.map(f=>[f,fs.readFileSync(path.join(dir,'src/compiler',f))]));let threw=false;const seen=observeWrites(()=>{try{adapter.adapt(dir);}catch(e){threw=true;return e.message;}});assert(threw,name);assert.equal(seen.writes,0,name);for(const [f,t]of before)assert(t.equals(fs.readFileSync(path.join(dir,'src/compiler',f))),name);results.push({name,refused:true,writes:seen.writes,reason:seen.value});});}
const planned=adapter.plan(original).texts;
fixture(original,dir=>{const first=observeWrites(()=>adapter.adapt(dir));assert.equal(first.value.files,2);assert.equal(first.writes,2);const stamps=files.map(f=>fs.statSync(path.join(dir,'src/compiler',f)).mtimeMs);const replay=observeWrites(()=>adapter.adapt(dir));assert.deepEqual(replay.value,{files:0,edits:0,javascriptChanging:true});assert.equal(replay.writes,0);assert.deepEqual(stamps,files.map(f=>fs.statSync(path.join(dir,'src/compiler',f)).mtimeMs));results.push({name:'first-apply-and-replay',first:first.value,replay:replay.value,replayWrites:0});});
for(const pin of adapter.pins){
 for(const [state,texts]of [['before',original],['after',planned]]){
  const owner=adapter.owners(texts.get(pin.file),pin.file).get(pin.owner);
  for(const variant of ['changed','missing','duplicate']){
   const next=new Map(texts),text=next.get(pin.file);
   const replacement=variant==='changed'?owner.text.replace('{','{ /* changed */'):variant==='missing'?'':owner.text+'\n'+owner.text;
   next.set(pin.file,text.slice(0,owner.start)+replacement+text.slice(owner.end));refused(pin.owner+'-'+state+'-'+variant,next);
  }
 }
 const mixed=new Map(original),owner=adapter.owners(mixed.get(pin.file),pin.file).get(pin.owner),text=mixed.get(pin.file);mixed.set(pin.file,text.slice(0,owner.start)+pin.after.replace(/\n/g,'\r\n')+text.slice(owner.end));refused(pin.owner+'-mixed',mixed);
}
const missingImports=new Map(planned);missingImports.set(adapter.imports.file,missingImports.get(adapter.imports.file).replace('    '+adapter.imports.names[0]+',\r\n',''));refused('partial-imports',missingImports);
const beforeImports=new Map(original);beforeImports.set(adapter.imports.file,beforeImports.get(adapter.imports.file).replace('    isAsExpression,','    '+adapter.imports.names[0]+',\r\n    isAsExpression,'));refused('partial-imports-before',beforeImports);
for(const ending of ['\n','\r\n']){const texts=new Map([...original].map(([f,t])=>[f,t.replace(/\r\n/g,'\n').replace(/\n/g,ending)]));const after=adapter.plan(texts).texts;for(const [f,t]of after){if(ending==='\r\n')assert(!/(?<!\r)\n/.test(t));else assert(!t.includes('\r'));}assert.equal(adapter.plan(after).edits,0);results.push({name:ending==='\n'?'LF-preserved':'CRLF-preserved',passed:true});}
for(const [state,texts]of [['before',original],['after',planned]]){const foreign=new Map([...texts].map(([f,t])=>[f,'// unrelated outside pinned owners\n'+t]));const p=adapter.plan(foreign);for(const t of p.texts.values())assert(t.startsWith('// unrelated outside pinned owners\n'));results.push({name:'foreign-edit-'+state,passed:true});}
fixture(original,dir=>{const read=fs.readFileSync;let calls=0;fs.readFileSync=function(name,...args){const result=Reflect.apply(read,this,[name,...args]);if(typeof name==='string'&&name.endsWith(files[0])&&++calls===2)return result+'// changed concurrently\n';return result;};let threw=false;const seen=observeWrites(()=>{try{adapter.adapt(dir);}catch(e){threw=true;return e.message;}});fs.readFileSync=read;assert(threw);assert.equal(seen.writes,0);results.push({name:'concurrent-source-change',refused:true,writes:0,reason:seen.value});});
console.log(JSON.stringify({cases:results.length,results},null,2));
