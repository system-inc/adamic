'use strict';
const fs=require('node:fs'),vm=require('node:vm'),assert=require('node:assert/strict');
const dir='/workspace/scratch/elementat-declaration/excluded-seeds';fs.mkdirSync(dir,{recursive:true});process.env.EA_CALL_DIR=dir;process.env.EA_SHARD='seed';require('./preload.cjs');
const object={id:1};assert.equal(globalThis.__eaObserve([object],0,object,object),object);
assert.throws(()=>globalThis.__eaObserve([object],0,object,{id:1}),/result difference/);
assert.throws(()=>globalThis.__eaObserve([1],0,1,2),/result difference/);
assert.throws(()=>vm.runInNewContext('__eaRead(() => fn,"seed:tdz","call",false); const fn=()=>{};',{}),e=>e.name==='ReferenceError');
const rows=fs.readdirSync(dir).filter(n=>n.endsWith('.reads.jsonl')).flatMap(n=>fs.readFileSync(dir+'/'+n,'utf8').trim().split('\n').map(JSON.parse));assert(rows.some(r=>r.tdz)); console.log('changed primitive, substituted reference, and cross-realm dead zone detected');

globalThis.__eaRead(()=>undefined,'seed:bundled-var','call',false);
const latest=fs.readdirSync(dir).filter(n=>n.endsWith('.reads.jsonl')).flatMap(n=>fs.readFileSync(dir+'/'+n,'utf8').trim().split('\n').map(JSON.parse)); assert(latest.some(r=>r.site==='seed:bundled-var'&&r.initialized===false));

const exported=()=>1; assert.equal(globalThis.__eaExportRead(()=>exported,'seed:export'),exported); globalThis.__eaExportRead(()=>undefined,'seed:export-var');
assert.throws(()=>vm.runInNewContext('__eaExportRead(()=>fn,"seed:export-tdz"); const fn=()=>{};',{}),e=>e.name==='ReferenceError');
const exportRows=fs.readdirSync(dir).filter(n=>n.endsWith('.exports.jsonl')).flatMap(n=>fs.readFileSync(dir+'/'+n,'utf8').trim().split('\n').map(JSON.parse)); assert(exportRows.some(r=>r.site==='seed:export-tdz'&&r.tdz)); assert(exportRows.some(r=>r.site==='seed:export-var'&&!r.initialized));
