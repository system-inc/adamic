"use strict";
const fs=require("node:fs"),path=require("node:path"),cp=require("node:child_process"),assert=require("node:assert/strict"),crypto=require("node:crypto");
const [input,output]=process.argv.slice(2).map(p=>path.resolve(p));
assert(input&&output,"usage: node maybebind-mutant.cjs PRE32 NEW_OUTPUT");
fs.mkdirSync(output,{recursive:false});
const rows=require("./maybebind-sites.json"), adapter=path.join(__dirname,"adapt.cjs");
function copy(name){const tree=path.join(output,name);fs.mkdirSync(tree);fs.cpSync(path.join(input,"src"),path.join(tree,"src"),{recursive:true});return tree;}
function snapshot(tree){return fs.readdirSync(path.join(tree,"src"),{recursive:true}).filter(f=>fs.statSync(path.join(tree,"src",f)).isFile()).sort().map(f=>{const p=path.join(tree,"src",f);return [f,crypto.createHash("sha256").update(fs.readFileSync(p)).digest("hex"),String(fs.statSync(p,{bigint:true}).mtimeNs)];});}
function run(tree){return cp.spawnSync(process.execPath,[adapter,tree],{encoding:"utf8"});}
function rewrite(tree,old,next){const file=path.join(tree,"src/compiler/watch.ts"),s=fs.readFileSync(file,"utf8");assert.equal(s.split(old).length,2);fs.writeFileSync(file,s.replace(old,next));}
const good=copy("accepted"),first=run(good);assert.equal(first.status,0,first.stderr);const state=snapshot(good),second=run(good);assert.equal(second.status,0,second.stderr);assert.deepEqual(snapshot(good),state);assert(JSON.parse(second.stdout).every(r=>!r.assertions&&!r.zeros&&!r.contracts&&!r.rewrites));
const mutants=[];
for(const name of ["mutant-half-owner","mutant-foreign-callee"]){
 const tree=copy(name);
 if(name==="mutant-half-owner")rewrite(tree,rows[0].before,rows[0].after);
 else rewrite(tree,"    maybeBind,","    noop as maybeBind,");
 const before=snapshot(tree),result=run(tree);assert.equal(result.status,1,name);assert.deepEqual(snapshot(tree),before,name+" wrote files");
 if(name==="mutant-half-owner")assert(result.stderr.includes("host producer invariant drift: createCompilerHostFromProgramHost"));
 else assert(result.stderr.includes("maybeBind callee symbol drift"),result.stderr);
 mutants.push({name,exit:result.status,refusal:result.stderr.trim(),noWrites:true});
}
console.log(JSON.stringify({pass:true,first:JSON.parse(first.stdout),secondRunNoWrites:true,mutants},null,2));
