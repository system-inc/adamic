'use strict';
const fs=require('node:fs'),assert=require('node:assert/strict'),Module=require('node:module'),path=require('node:path');
const base='/workspace/scratch';
const counter=path.join(base,'99d-main-catch.jsonl');
const raw=fs.readFileSync(counter,'utf8');
const pairs={};for(const line of raw.trim().split('\n').filter(Boolean)){const pair=JSON.stringify(JSON.parse(line));pairs[pair]=(pairs[pair]||0)+1;}
const emitted=path.join(base,'99d-main-oracle-tree/built/local/typescript.js');
const text=fs.readFileSync(emitted,'utf8');assert(text.includes(counter));
const runner=fs.readFileSync(path.join(base,'99d-main-oracle-tree/built/local/run.js'),'utf8');assert(runner.includes(counter));
const sanity=path.join(base,'99d-main-counter-sanity.jsonl');fs.writeFileSync(sanity,'');
const probe=new Module(emitted,module);probe.filename=emitted;probe.paths=Module._nodeModulePaths(path.dirname(emitted));
probe._compile(text.replaceAll(counter,sanity),emitted);
const result=probe.exports.tryReadFile('probe.json',()=>{throw new Error('counter sanity');});
assert.equal(result.messageText,"Cannot read file 'probe.json': counter sanity.");
assert.equal(fs.readFileSync(sanity,'utf8'),'["object","string"]\n');assert.equal(fs.readFileSync(counter,'utf8'),raw);
console.log(JSON.stringify({total:Object.values(pairs).reduce((a,b)=>a+b,0),pairs,emittedCompilerHasCounter:true,emittedRunnerHasCounter:true,separatePositiveControl:{pair:['object','string'],count:1,diagnostic:result.messageText},suiteCounterUnchanged:true},null,2));
