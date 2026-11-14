'use strict';
const fs=require('node:fs'),assert=require('node:assert/strict'),path=require('node:path');
const root='/workspace/scratch/binding-views-resolution',dir=root+'/repo/stage3/adapt/99-correctness-binding-views';
const {adapt,rewrite,owners}=require(dir+'/adapt.cjs');assert.equal(owners.length,1);assert.equal(owners[0].file,'resolutionCache.ts');const pin=owners[0];
const tree=root+'/contracts';fs.mkdirSync(tree+'/src/compiler',{recursive:true});const file=tree+'/src/compiler/resolutionCache.ts',program=tree+'/src/compiler/program.ts';const sentinel='// program.ts must not be read or changed\n';fs.writeFileSync(program,sentinel);const rows=[];
function monitored(fn){let writes=[];const original=fs.writeFileSync;fs.writeFileSync=(name,...args)=>{writes.push(name);return original(name,...args);};try{return {value:fn(),writes};}finally{fs.writeFileSync=original;assert.equal(fs.readFileSync(program,'utf8'),sentinel);}}
function refused(name,text){fs.writeFileSync(file,text);const result=monitored(()=>assert.throws(()=>adapt(tree)));assert.equal(result.writes.length,0);rows.push({name,refused:true,writes:0});}
fs.writeFileSync(file,pin.before);let x=monitored(()=>adapt(tree));assert.equal(x.value.files,1);assert.deepEqual(x.writes,[file]);assert.equal(fs.readFileSync(file,'utf8'),pin.after);rows.push({name:'apply writes only resolutionCache.ts',writes:1});
x=monitored(()=>adapt(tree));assert.equal(x.value.files,0);assert.equal(x.writes.length,0);rows.push({name:'adapter replay',writes:0});
refused('owner drift',pin.before.replace('interface ResolutionCache','interface ResolutionCacheDrift'));
refused('outside-owner drift',pin.before+'\n');
refused('duplicate owner','interface ResolutionReadView {}\n'+pin.after);
refused('half-adapted constraints',pin.before.replace('R extends ResolutionWithResolvedFileName','R extends ResolutionReadView'));
refused('half-adapted inserted owner',pin.after.replace('export interface ResolutionReadView','interface ResolutionReadView'));
refused('mixed line endings',pin.before.replace('\n','\r\n'));
fs.writeFileSync(file,pin.before.replace(/\n/g,'\r\n'));x=monitored(()=>adapt(tree));assert.equal(x.value.files,1);assert.equal(fs.readFileSync(file,'utf8'),pin.after.replace(/\n/g,'\r\n'));rows.push({name:'CRLF preserved',pass:true});
x=monitored(()=>rewrite(tree,true));assert.equal(x.value.files,1);assert.deepEqual(x.writes,[file]);assert.equal(fs.readFileSync(file,'utf8'),pin.before.replace(/\n/g,'\r\n'));rows.push({name:'mutant restores only resolutionCache.ts',writes:1});
x=monitored(()=>rewrite(tree,true));assert.equal(x.value.files,0);assert.equal(x.writes.length,0);rows.push({name:'mutant replay',writes:0});
// Both commands work with the untouched program sentinel removed: no dependency.
fs.unlinkSync(program);fs.writeFileSync(file,pin.before);assert.equal(adapt(tree).files,1);assert.equal(rewrite(tree,true).files,1);rows.push({name:'no program.ts dependency',pass:true});
fs.writeFileSync(root+'/evidence/contracts.json',JSON.stringify({status:'pass',rows},null,2)+'\n');console.log(JSON.stringify(rows));
