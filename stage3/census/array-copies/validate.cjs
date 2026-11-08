// Focused integration checks for classification, loop scopes and runtime transform.
const fs=require('fs'),os=require('os'),path=require('path'),cp=require('child_process'),assert=require('assert');
const api=path.resolve(process.argv[2]),ts=require(api),scratch=fs.mkdtempSync(path.join(os.tmpdir(),'array-copies-validation-')),src=path.join(scratch,'src/compiler'),here=__dirname;
fs.mkdirSync(src,{recursive:true});
fs.writeFileSync(path.join(src,'tsconfig.json'),JSON.stringify({compilerOptions:{strict:true,target:'ES2024',module:'CommonJS',types:[]},include:['*.ts']}));
fs.writeFileSync(path.join(src,'core.ts'),'export function sameMap<T>(array: T[], f: (x:T)=>T): T[] { const result=array.map(f); return result.every((x,i)=>x===array[i])?array:result; }');
const text=`import {sameMap} from './core';
const objects=[{value:1},{value:2}];
const branded=['x'] as (string & {brand:true})[];
branded.slice();
const mixed: (object | string)[]=[{},'x'];
mixed.slice();
function named(){return objects.slice();}
objects.forEach(named);
for(const item of objects){
 objects.slice();
 objects.map(()=>{objects.slice(); return item;});
 function nested(){return objects.slice();}
 nested();
 (()=>objects.slice())();
 sameMap(objects,x=>x);
 sameMap(objects,x=>({value:x.value}));
 [0,...objects,0];
 [...objects,...objects];
}
console.log(JSON.stringify(objects));
Array.from(new Map([[1,{value:1}]]));
Array.from({*[Symbol.iterator](){yield {value:1};}});
let evaluated=0;
function optional(input: typeof objects | undefined){for(const obj of objects){input?.map((evaluated++,x=>x));}}
optional(undefined);optional(objects);if(evaluated!==2)throw Error("optional arguments evaluated incorrectly");
function broad<T extends {}>(values:T[]){for(const value of values){values.slice();}}
objects.map(()=>{function callbackNested(){return objects.slice();}return callbackNested();});
`;
fs.writeFileSync(path.join(src,'fixture.ts'),text);
function run(name,args,env={}){const log=path.join(scratch,name+'.log'),fd=fs.openSync(log,'w');const r=cp.spawnSync(process.execPath,args,{stdio:['ignore',fd,fd],env:{...process.env,...env}});fs.closeSync(fd);console.log(`${name} child exit code: ${r.status}`);assert.strictEqual(r.status,0,fs.readFileSync(log,'utf8'));return fs.readFileSync(log,'utf8');}
const json=path.join(scratch,'sites.json');run('census',[path.join(here,'census.cjs'),api,scratch,json]);
const d=JSON.parse(fs.readFileSync(json));assert.strictEqual(d.diagnostics.length,0);
assert.strictEqual(d.sites.find(s=>s.expression==='values.slice()').classification,'unresolved');
assert.strictEqual(d.sites.find(s=>s.enclosingFunction==='callbackNested').uncertainOuterLoop,true);
const at=line=>d.sites.find(s=>s.file.endsWith('fixture.ts')&&s.line===line);
assert(!at(4),'branded string copy must be excluded');assert.strictEqual(at(6).classification,'mixed');assert.strictEqual(at(7).depth,1);assert.strictEqual(at(10).depth,1);
assert(d.sites.some(s=>s.line===11&&s.method==='.slice'&&s.depth===2));
assert.strictEqual(at(12).depth,0);assert.strictEqual(at(12).uncertainOuterLoop,true);
assert.strictEqual(at(14).depth,0);assert.strictEqual(at(14).uncertainOuterLoop,true);
assert.strictEqual(new Set(d.sites.map(s=>s.id)).size,d.sites.length);
assert.strictEqual(d.sites.filter(s=>s.method==='Array.from'&&s.classification==='object').length,2);
const baseline=run('baseline',['-e',ts.transpileModule(text.replace("import {sameMap} from './core';",fs.readFileSync(path.join(src,'core.ts'),'utf8').replace('export ','')),{compilerOptions:{module:ts.ModuleKind.CommonJS}}).outputText]);
run('transform',[path.join(here,'transform.cjs'),api,json,scratch]);
for(const name of ['core','fixture'])fs.writeFileSync(path.join(src,name+'.js'),ts.transpileModule(fs.readFileSync(path.join(src,name+'.ts'),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS}}).outputText);
const runtime=path.join(scratch,'runtime.json');const counted=run('counted',['-r',path.join(here,'counter.cjs'),path.join(src,'fixture.js')],{ARRAY_COPIES_SITES:json,ARRAY_COPIES_OUTPUT:runtime});
assert.strictEqual(counted,baseline);
const r=JSON.parse(fs.readFileSync(runtime));
const unchanged=at(15),changed=at(16);
assert.strictEqual(r.sites.find(s=>s.id===unchanged.id).calls,2);assert.strictEqual(r.sites.find(s=>s.id===unchanged.id).executions,0);
assert.strictEqual(r.sites.find(s=>s.id===changed.id).executions,2);assert.strictEqual(r.sites.find(s=>s.id===changed.id).elements,4);
const spread=d.sites.find(s=>s.method==='spread');assert(spread);assert.strictEqual(r.sites.find(s=>s.id===spread.id).calls,2);assert.strictEqual(r.sites.find(s=>s.id===spread.id).elements,8);
assert.strictEqual(d.sites.filter(s=>s.method==='spread').length,3);
for(const s of d.sites.filter(s=>s.method==='spread'))assert.strictEqual(r.sites.find(row=>row.id===s.id).elements,8);
const optional=d.sites.find(s=>s.expression.startsWith('input?.map'));assert(optional);assert.strictEqual(r.sites.find(s=>s.id===optional.id).calls,2);assert.strictEqual(r.totalThrows,0);
console.log(JSON.stringify({passed:true,scratch,sites:d.sites.length,totalExecutions:r.totalExecutions,totalElements:r.totalElements}));
