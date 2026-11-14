'use strict';
const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),cp=require('node:child_process'),assert=require('node:assert/strict');
const adapter=require('./adapt.cjs');
const tree=process.argv[2],root=fs.mkdtempSync(path.join(os.tmpdir(),'sourcemap-contract-'));
fs.mkdirSync(path.join(root,'src/compiler'),{recursive:true});
const files=[...new Set(adapter.owners.map(p=>p.file))];
const before={};for(const file of files){const r=cp.spawnSync('git',['-C',tree,'show','refs/stage3/99-correctness-axolotl-s-before:src/compiler/'+file],{encoding:'utf8'});assert.equal(r.status,0,r.stderr);before[file]=r.stdout.replace(/\r\n/g,'\n');}
const after={...before};for(const pin of adapter.owners){assert(before[pin.file].includes(pin.before));after[pin.file]=after[pin.file].replace(pin.before,pin.after);}
const originalWrite=fs.writeFileSync;let writes=0;fs.writeFileSync=(...args)=>{writes++;return originalWrite(...args);};
function seed(data,eol='\n'){for(const file of files)originalWrite(path.join(root,'src/compiler',file),data[file].replace(/\n/g,eol));writes=0;}
function check(data){for(const file of files)assert.equal(fs.readFileSync(path.join(root,'src/compiler',file),'utf8'),data[file]);}
try{
 seed(before);assert.deepEqual(adapter.adapt(root),{files:2,edits:5,javascriptChanging:true});assert.equal(writes,2);check(after);
 const stats=files.map(f=>fs.statSync(path.join(root,'src/compiler',f),{bigint:true}).mtimeNs);writes=0;assert.equal(adapter.adapt(root).edits,0);assert.equal(writes,0);check(after);files.forEach((f,i)=>assert.equal(fs.statSync(path.join(root,'src/compiler',f),{bigint:true}).mtimeNs,stats[i]));
 const mutations=[];
 for(const p of adapter.owners){
  for(const [label,data,text] of [['before',before,p.before],['after',after,p.after]]){
   for(const [kind,replacement] of [['changed',text.replace('{','{ /* drift */')],['missing',''],['duplicate',text+'\n'+text],['moved','function foreignContainer() {\n'+text+'\n}']])mutations.push([p.owner+':'+label+':'+kind,{...data,[p.file]:data[p.file].replace(text,replacement)}]);
  }
 }
 // Each core mixed state and the new-owner-only state are refused; complete previous 99s is a migration.
 for(const p of adapter.owners.filter(p=>p.file==='core.ts'))mutations.push([p.owner+':mixed',{...before,[p.file]:before[p.file].replace(p.before,p.after)}]);
 mutations.push(['parse error', {...before, 'sourcemap.ts': before['sourcemap.ts']+'\nfunction {'}]);
 const sm=adapter.owners.at(-1);mutations.push(['sourcemap-only',{...before,[sm.file]:after[sm.file]}]);
 for(const [name,data] of mutations){seed(data);assert.throws(()=>adapter.adapt(root),undefined,name);assert.equal(writes,0,name);check(data);}
 seed({...after,[sm.file]:before[sm.file]});assert.equal(adapter.adapt(root).edits,1);assert.equal(writes,1);check(after);
 for(const eol of ['\n','\r\n']){seed(before,eol);adapter.adapt(root);for(const f of files)assert.equal(fs.readFileSync(path.join(root,'src/compiler',f),'utf8'),after[f].replace(/\n/g,eol));}
 const foreign={...before,'core.ts':before['core.ts'].replace('function selectIndex(_: unknown, i: number)', 'function selectIndex(_foreign: unknown, i: number)')};
 assert.notEqual(foreign['core.ts'],before['core.ts']);seed(foreign);adapter.adapt(root);assert(fs.readFileSync(path.join(root,'src/compiler/core.ts'),'utf8').includes('selectIndex(_foreign: unknown'));
 assert.deepEqual(adapter.owners,JSON.parse(fs.readFileSync(path.join(path.dirname(require.resolve('./adapt.cjs')),'owners.json'),'utf8')));
 console.log(JSON.stringify({status:'PASS',refusals:mutations.map(m=>m[0]),replayWrites:0,wholeOwnerMigration:true,crlfAndLf:true}));
}finally{fs.writeFileSync=originalWrite;fs.rmSync(root,{recursive:true,force:true});}
