'use strict';
const fs=require('node:fs'),os=require('node:os'),path=require('node:path'),assert=require('node:assert/strict');
const adapter=require('./adapt.cjs'),raw=fs.readFileSync(path.join(process.argv[2],'src/compiler/core.ts'),'utf8').replace(/\r\n/g,'\n');
const dir=fs.mkdtempSync(path.join(os.tmpdir(),'combine-contract-')),file=path.join(dir,'src/compiler/core.ts');fs.mkdirSync(path.dirname(file),{recursive:true});
const originalWrite=fs.writeFileSync;let writes=0;fs.writeFileSync=(...args)=>{writes++;return originalWrite(...args);};
function seed(t){originalWrite(file,t);writes=0;}
seed(raw);assert.deepEqual(adapter.adapt(dir),{files:1,edits:2,javascriptChanging:false});assert.equal(writes,1);
const after=fs.readFileSync(file,'utf8'),stat=fs.statSync(file,{bigint:true});writes=0;assert.deepEqual(adapter.adapt(dir),{files:0,edits:0,javascriptChanging:false});assert.equal(writes,0);assert.equal(fs.statSync(file,{bigint:true}).mtimeNs,stat.mtimeNs);assert.equal(fs.readFileSync(file,'utf8'),after);
const p=adapter.owners[0],cases=[
 ['changed owner',raw.replace(p.before,p.before.replace('return [xs, ys]','return [ys, xs]'))],
 ['changed overload',raw.replace('xs: T[] | undefined, ys: T[] | undefined','xs: readonly T[] | undefined, ys: T[] | undefined')],
 ['changed accepted owner',after.replace('return [xs, ys]','return [ys, xs]')],
 ['missing owner',raw.replace(p.before,'')],
 ['duplicate owner',raw+'\n'+p.before],
 ['nested duplicate',raw+'\nfunction outer() {\n'+p.before+'\n}'],
 ['moved owner',raw.replace(p.before,'function outer() {\n'+p.before+'\n}')],
 ['only first call annotated',raw.replace('append(xs, ys)','append<T>(xs, ys)')],
 ['only second call annotated',raw.replace('append(ys, xs)','append<T>(ys, xs)')],
 ['parse error',raw+'\nfunction {'],
];
for(const [label,t]of cases){assert.notEqual(t,raw,label);seed(t);assert.throws(()=>adapter.adapt(dir),undefined,label);assert.equal(writes,0,label);assert.equal(fs.readFileSync(file,'utf8'),t,label);}
for(const eol of ['\n','\r\n']){seed((raw+'\n// unrelated edit\n').replace(/\n/g,eol));adapter.adapt(dir);const t=fs.readFileSync(file,'utf8');assert(t.endsWith('// unrelated edit'+eol));assert.equal(t.includes('\r'),eol==='\r\n');}
fs.writeFileSync=originalWrite;fs.rmSync(dir,{recursive:true});console.log(JSON.stringify({firstRun:{files:1,edits:2,writes:1},secondRun:{files:0,edits:0,writes:0,bytesAndMtimePreserved:true},noWriteRefusals:cases.map(x=>x[0]),unrelatedEditPreserved:true,crlfAndLfPreserved:true},null,2));
