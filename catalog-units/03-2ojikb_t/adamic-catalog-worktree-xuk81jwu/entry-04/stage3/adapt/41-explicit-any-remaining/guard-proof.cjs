"use strict";
const fs=require("node:fs"),path=require("node:path"),assert=require("node:assert/strict"),cp=require("node:child_process");
const [beforeTree,afterTree,output]=process.argv.slice(2),rules=require("./rules.json");
const files=[...new Set(rules.map(r=>r.file))];
const copy=(tree,lf=false)=>{const dest=fs.mkdtempSync("/tmp/adaptation41-guards-");for(const file of files){const target=path.join(dest,file);fs.mkdirSync(path.dirname(target),{recursive:true});const text=fs.readFileSync(path.join(tree,file),"utf8");fs.writeFileSync(target,lf?text.replaceAll("\r\n","\n"):text);}return dest;};
const run=tree=>cp.spawnSync(process.execPath,[path.join(__dirname,"adapt.cjs"),tree],{encoding:"utf8",env:process.env});
const control=copy(afterTree),snapshot=Object.fromEntries(files.map(f=>[f,fs.readFileSync(path.join(control,f),"utf8")]));
let result=run(control);assert.equal(result.status,0,result.stderr);assert(result.stdout.includes('"sites":0'));
for(const f of files)assert.equal(fs.readFileSync(path.join(control,f),"utf8"),snapshot[f]);
const newline=copy(beforeTree,true);result=run(newline);assert.equal(result.status,0,result.stderr);
for(const f of files)assert.equal(fs.readFileSync(path.join(newline,f),"utf8"),fs.readFileSync(path.join(afterTree,f),"utf8").replaceAll("\r\n","\n"),"LF reconstruction "+f);
const mutations=[
 {name:"duplicate retained interface anchor",replace:s=>s+"\nexport interface ObjectAllocator { }\n",message:"missing or duplicate reviewed site: 26"},
 {name:"drift retained interface anchor",replace:s=>s.replace("export interface ObjectAllocator {","export interface ObjectAllocator /* drift */ {"),message:"missing or duplicate reviewed site: 26"}
];
const evidence={idempotent:true,LFReconstruction:true,mutants:[]};
for(const mutant of mutations){const tree=copy(afterTree),file=path.join(tree,"src/compiler/utilities.ts");fs.writeFileSync(file,mutant.replace(fs.readFileSync(file,"utf8")));result=run(tree);assert.notEqual(result.status,0);assert(result.stderr.includes(mutant.message),result.stderr);evidence.mutants.push({name:mutant.name,caughtBy:mutant.message});}
fs.writeFileSync(output,JSON.stringify(evidence,null,2)+"\n");console.log(JSON.stringify(evidence));
