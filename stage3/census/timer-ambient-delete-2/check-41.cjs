'use strict';
const fs=require('fs'),path=require('path'),cp=require('child_process'),assert=require('assert/strict'),os=require('os');
const repo=process.argv[2],tree=process.argv[3],dir=path.join(repo,'stage3/adapt/41-explicit-any-remaining');
const rules=require(path.join(dir,'rules.json')),files=[...new Set(rules.map(r=>r.file))];
const before={};for(const file of files)before[file]=cp.execFileSync('git',['-C',tree,'show','refs/stage3/41-explicit-any-remaining-before:'+file],{maxBuffer:32*1024*1024});
function make(content){const root=fs.mkdtempSync(path.join(os.tmpdir(),'road5-41-'));for(const [file,text] of Object.entries(content)){fs.mkdirSync(path.dirname(path.join(root,file)),{recursive:true});fs.writeFileSync(path.join(root,file),text)}return root}
function snapshot(root){return Object.fromEntries(files.map(f=>[f,{text:fs.readFileSync(path.join(root,f)).toString('base64'),mtime:fs.statSync(path.join(root,f)).mtimeMs}]))}
function run(root,check=false){return cp.spawnSync(process.execPath,[path.join(dir,check?'timer-heads-check.cjs':'adapt.cjs'),root],{encoding:'utf8',env:process.env})}
const results=[];
for(const crlf of [false,true]){
 const input=Object.fromEntries(Object.entries(before).map(([f,b])=>[f,b.toString().replace(/\r\n/g,'\n').replace(/\n/g,crlf?'\r\n':'\n')]));
 const root=make(input),first=run(root);assert.equal(first.status,0,first.stderr);const snap=snapshot(root),replay=run(root);assert.equal(replay.status,0,replay.stderr);assert(replay.stdout.includes('"sites":0'));assert.deepEqual(snapshot(root),snap);assert.equal(run(root,true).status,0);results.push({name:crlf?'CRLF':'LF',apply:0,replay:0,sites:0,bytesAndMtimesUnchanged:true});
 const sys='src/compiler/sys.ts',original=input[sys],heads=rules.filter(r=>r.kind==='timers');
 for(const [name,mutate] of [
 ['half-pair',s=>s.replace(heads[0].before+(crlf?'\r\n':'\n'),'')],
 ['drift',s=>s.replace(heads[0].before,heads[0].before.replace('handler:','handler /* drift */:'))],
 ['duplicate',s=>s+'\n'+heads[0].before+'\n'],
 ['moved',s=>s.replace(heads[0].before,'namespace Moved { '+heads[0].before+' }')]
 ]){const mutant=make({...input,[sys]:mutate(original)}),snap=snapshot(mutant),r=run(mutant);assert.notEqual(r.status,0,name);assert.deepEqual(snapshot(mutant),snap);results.push({name:(crlf?'CRLF ':'LF ')+name,exit:r.status,noWrites:true,diagnostic:r.stderr.split('\n').find(l=>/Error:/.test(l))})}
 const restored=make(Object.fromEntries(files.map(f=>[f,fs.readFileSync(path.join(root,f))])));fs.appendFileSync(path.join(restored,sys),'\n'+heads.map(r=>r.before).join('\n')+'\n');const red=run(restored,true);assert.notEqual(red.status,0);assert(red.stderr.includes('41 timer-head check: ambient timer declaration restored'));results.push({name:'restored-ambient-pair '+(crlf?'CRLF':'LF'),exit:red.status,diagnostic:'41 timer-head check: ambient timer declaration restored'});
 const unrelated=make({...input,[sys]:original+'\n// unrelated source\n'});const r=run(unrelated);assert.equal(r.status,0,r.stderr);assert(fs.readFileSync(path.join(unrelated,sys),'utf8').endsWith('// unrelated source\n'));
}
console.log(JSON.stringify(results,null,2));
