'use strict';
const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),assert=require('node:assert/strict'),adapter=require('./adapt.cjs');
const dir=fs.mkdtempSync(path.join(os.tmpdir(),'sorted-concatenate-contract-')),compiler=path.join(dir,'src/compiler');fs.mkdirSync(compiler,{recursive:true});
const originals=Object.fromEntries([...new Set(adapter.owners.map(p=>p.file))].map(file=>[file,fs.readFileSync(path.join(process.argv[2],'src/compiler',file),'utf8')]));
const originalWrite=fs.writeFileSync;let writes=0;fs.writeFileSync=(...args)=>{writes++;return originalWrite(...args);};
function seed(texts){for(const [file,text]of Object.entries(texts))originalWrite(path.join(compiler,file),text);writes=0;}
function snapshot(){return Object.fromEntries(Object.keys(originals).map(file=>[file,fs.readFileSync(path.join(compiler,file),'utf8')]));}
seed(originals);assert.equal(adapter.adapt(dir).files,2);assert.equal(writes,2);const after=snapshot();
const mtimes=Object.keys(originals).map(file=>fs.statSync(path.join(compiler,file),{bigint:true}).mtimeNs);writes=0;
assert.deepEqual(adapter.adapt(dir),{files:0,edits:0,javascriptChanging:true});assert.equal(writes,0);assert.deepEqual(snapshot(),after);
assert.deepEqual(Object.keys(originals).map(file=>fs.statSync(path.join(compiler,file),{bigint:true}).mtimeNs),mtimes);
const refusals=[];
function refuse(label,texts){seed(texts);assert.throws(()=>adapter.adapt(dir),undefined,label);assert.equal(writes,0,label);assert.deepEqual(snapshot(),texts,label);refusals.push(label);}
for(const pin of adapter.owners){
 const eol=originals[pin.file].includes('\r\n')?'\r\n':'\n',spell=s=>s.replace(/\n/g,eol),before=spell(pin.before),adapted=spell(pin.after);
 assert(originals[pin.file].includes(before));assert(after[pin.file].includes(adapted));
 const mutate=s=>s.replace('function '+pin.owner,'function /* changed owner */ '+pin.owner);
 refuse(pin.owner+': changed before',{...originals,[pin.file]:originals[pin.file].replace(before,mutate(before))});
 refuse(pin.owner+': changed after',{...after,[pin.file]:after[pin.file].replace(adapted,mutate(adapted))});
 refuse(pin.owner+': missing',{...originals,[pin.file]:originals[pin.file].replace(before,'')});
 refuse(pin.owner+': duplicate',{...originals,[pin.file]:originals[pin.file]+eol+before});
 refuse(pin.owner+': nested duplicate',{...originals,[pin.file]:originals[pin.file]+eol+'function foreignContainer(){'+eol+before+eol+'}'});
 refuse(pin.owner+': moved',{...originals,[pin.file]:originals[pin.file].replace(before,'function foreignContainer(){'+eol+before+eol+'}')});
 refuse(pin.owner+': moved into arrow',{...originals,[pin.file]:originals[pin.file].replace(before,'const foreignArrow=()=>{'+eol+before+eol+'};')});
 refuse(pin.owner+': moved into block',{...originals,[pin.file]:originals[pin.file].replace(before,'if (true) {'+eol+before+eol+'}')});
 refuse(pin.owner+': mixed state',{...originals,[pin.file]:originals[pin.file].replace(before,adapted)});
 if(pin.siteEdits>1){
  const a=pin.before.split('\n'),b=pin.after.split('\n');
  if(a.length===b.length){const i=a.findIndex((s,i)=>s!==b[i]);a[i]=b[i];refuse(pin.owner+': partial owner',{...originals,[pin.file]:originals[pin.file].replace(before,spell(a.join('\n')))});}
 }
}
refuse('parse error',{...originals,'core.ts':originals['core.ts']+'\nfunction {'});
refuse('mixed files',{...originals,'core.ts':after['core.ts']});
const foreign={...originals,'core.ts':originals['core.ts'].replace('function selectIndex(_: unknown, i: number)','function selectIndex(_foreign: unknown, i: number)')};assert.notEqual(foreign['core.ts'],originals['core.ts']);seed(foreign);assert.equal(adapter.adapt(dir).files,2);assert(snapshot()['core.ts'].includes('selectIndex(_foreign: unknown'));
for(const eol of ['\n','\r\n']) {seed(Object.fromEntries(Object.entries(originals).map(([f,t])=>[f,t.replace(/\r\n/g,'\n').replace(/\n/g,eol)])));assert.equal(adapter.adapt(dir).files,2);for(const t of Object.values(snapshot()))assert(eol==='\n'?!t.includes('\r'):!t.replace(/\r\n/g,'').includes('\n'));}
console.log(JSON.stringify({firstRunFiles:2,replay:{files:0,edits:0,writes:0,bytesAndMtimePreserved:true},refusals,foreignEditPreserved:true,lineEndingsPreserved:true},null,2));
