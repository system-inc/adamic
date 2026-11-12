'use strict';
const fs=require('node:fs'),Module=require('node:module'),assert=require('node:assert/strict');
const output='/workspace/scratch/tracing-fs-import-evidence/catch2-positive-events.jsonl';process.env.TRACING_CATCH_EVENTS=output;
const events=[],original=Module._load;
Module._load=function(name,...args){if(name==='fs'||name==='node:fs')events.push({specifier:name,stack:new Error().stack});return original.call(this,name,...args);};
const ts=require('/workspace/scratch/tracing-control2-counted/built/local/typescript.js');
const loadsBeforeStart=events.slice();assert(loadsBeforeStart.some(e=>e.specifier==='fs'&&!/startTracing/.test(e.stack)));assert(!loadsBeforeStart.some(e=>/startTracing/.test(e.stack)));
Module._load=function(name,...args){if(name==='fs'&&/startTracing/.test(new Error().stack))throw new Error('counter control forced fs load failure');return original.call(this,name,...args);};
try{assert.throws(()=>ts.tracingEnabled.startTracing('server','/workspace/scratch/catch2-positive-trace'),/tracing requires having fs/);}finally{Module._load=original;}
const counts=fs.readFileSync(output,'utf8').trim().split('\n').map(JSON.parse);assert.equal(counts.filter(e=>e.kind==='catch').length,1);
fs.writeFileSync('/workspace/scratch/tracing-fs-import-evidence/catch2-positive.json',JSON.stringify({catchExecutions:1,forcedFailure:true,builtinLoadsBeforeStart:loadsBeforeStart,graph:['src/tsc/tsc.ts imports its namespace','src/tsc/_namespaces/ts.ts reexports compiler namespace and executeCommandLine','compiler namespace reexports sys.ts and tracing.ts','sys.ts imports node:fs at module scope','tsc.ts invokes executeCommandLine after dependencies finish evaluating','executeCommandLine may then invoke startTracing']},null,2)+'\n');
