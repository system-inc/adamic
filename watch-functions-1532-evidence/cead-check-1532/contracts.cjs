'use strict';
const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process'),assert=require('node:assert/strict');
const s='/tmp/watch-file-check-cead',repo=s+'/work',tree=s+'/after-tree',ts=require(repo+'/stage3/api/node_modules/typescript');
function phase(name){const out=s+'/phase-'+name;fs.mkdirSync(out);const tar=cp.execFileSync('git',['-C',tree,'archive','refs/stage3/'+name+'-before','src'],{maxBuffer:128*1024*1024});cp.execFileSync('tar',['-x','-C',out],{input:tar});fs.symlinkSync(tree+'/node_modules',out+'/node_modules','dir');return out;}
const env={...process.env,NODE_PATH:repo+'/stage3/api/node_modules',TMPDIR:s+'/tmp'};
const results=[];
for(const name of ['99-correctness-core-multimap','99-correctness-single-watcher-forwarder','99-correctness-timer-closures']){
 const input=phase(name),args=[repo+'/stage3/adapt/'+name+'/contract.cjs',input];if(name==='99-correctness-core-multimap')args.push(s+'/core-contract-scratch',s+'/core-contract.json');
 const output=cp.execFileSync('node',args,{env,encoding:'utf8'});fs.writeFileSync(s+'/'+name+'-contract.json',output);results.push({name,exit:0});
}
for(const name of ['93-builtin-imports','99-wasm-crypto-import']){
 const input=phase(name),adapter=require(repo+'/stage3/adapt/'+name+'/adapt.cjs'),file=input+'/src/compiler/sys.ts';
 assert.equal(adapter.adapt(input).files,1);const raw=fs.readFileSync(file,'utf8'),stamp=String(fs.statSync(file,{bigint:true}).mtimeNs),write=fs.writeFileSync;
 const manifest=require(repo+'/stage3/adapt/'+name+'/proofs/targets.json')[0],source=ts.createSourceFile(file,raw,ts.ScriptTarget.Latest,true);let owner;function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='getNodeSystem'){assert(!owner);owner=n;}ts.forEachChild(n,visit);}visit(source);assert.equal(owner.getText(source),manifest.ownerAfterText);
 function noWrites(fn){let writes=0;fs.writeFileSync=(...args)=>{writes++;return write(...args);};try{fn();assert.equal(writes,0);}finally{fs.writeFileSync=write;}}
 noWrites(()=>assert.equal(adapter.adapt(input).files,0));assert.equal(fs.readFileSync(file,'utf8'),raw);assert.equal(String(fs.statSync(file,{bigint:true}).mtimeNs),stamp);
 const changes={interval:t=>t.replace('function fsWatchFileWorker(fileName: string, callback: FileWatcherCallback, pollingInterval: number | undefined)','function fsWatchFileWorker(fileName: string, callback: FileWatcherCallback, pollingInterval: number)'),duplicate:t=>t+'\n'+manifest.ownerAfterText,mixedInside:t=>t.replace('function getNodeSystem(): System {\r\n','function getNodeSystem(): System {\n'),import:t=>t.replace(name.startsWith('93')?'import * as _fs from "node:fs";':'import * as _crypto from "node:crypto";','')};
 for(const [label,change] of Object.entries(changes)){const mutated=change(raw);assert.notEqual(mutated,raw,label);fs.writeFileSync(file,mutated);const mtime=String(fs.statSync(file,{bigint:true}).mtimeNs);noWrites(()=>assert.throws(()=>adapter.adapt(input)));assert.equal(fs.readFileSync(file,'utf8'),mutated);assert.equal(String(fs.statSync(file,{bigint:true}).mtimeNs),mtime);fs.writeFileSync(file,raw);}
 results.push({name,fullAfterTextPinned:true,replayZeroWrites:true,bytesAndMtimePreserved:true,refusals:Object.keys(changes)});
}
fs.writeFileSync(s+'/contracts.json',JSON.stringify(results,null,2)+'\n');console.log(JSON.stringify(results,null,2));
