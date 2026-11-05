'use strict';
const fs=require('node:fs'),os=require('node:os'),path=require('node:path'),assert=require('node:assert/strict');
const adapter=require('./adapt.cjs');
const raw=fs.readFileSync(path.join(process.argv[2],'src/compiler/core.ts'),'utf8');
const dir=fs.mkdtempSync(path.join(os.tmpdir(),'sorted-brand-contract-'));
fs.mkdirSync(path.join(dir,'src/compiler'),{recursive:true});
const file=path.join(dir,'src/compiler/core.ts'),originalWrite=fs.writeFileSync;
let writes=0;fs.writeFileSync=(...args)=>{writes++;return originalWrite(...args);};
function seed(text){originalWrite(file,text);writes=0;}
const edits=adapter.owners.reduce((n,p)=>n+p.siteEdits,0);
seed(raw);assert.deepEqual(adapter.adapt(dir),{files:1,edits,javascriptChanging:false});assert.equal(writes,1);
const after=fs.readFileSync(file,'utf8'),stat=fs.statSync(file,{bigint:true});writes=0;
assert.deepEqual(adapter.adapt(dir),{files:0,edits:0,javascriptChanging:false});assert.equal(writes,0);assert.equal(fs.readFileSync(file,'utf8'),after);assert.equal(fs.statSync(file,{bigint:true}).mtimeNs,stat.mtimeNs);
const eol=raw.includes('\r\n')?'\r\n':'\n',spell=s=>s.replace(/\n/g,eol),mutations=[];
for(const pin of adapter.owners){
 const before=spell(pin.before),adapted=spell(pin.after);assert(raw.includes(before));assert(after.includes(adapted));
 const changed=s=>s.replace('{','{ /* foreign owner change */');
 mutations.push([pin.owner+': changed owner',raw.replace(before,changed(before))]);
 mutations.push([pin.owner+': changed adapted owner',after.replace(adapted,changed(adapted))]);
 mutations.push([pin.owner+': missing owner',raw.replace(before,'')]);
 mutations.push([pin.owner+': duplicate owner',raw+eol+before]);
 mutations.push([pin.owner+': nested duplicate owner',raw+eol+'function foreignContainer() {'+eol+before+eol+'}']);
 mutations.push([pin.owner+': moved owner',raw.replace(before,'function foreignContainer() {'+eol+before+eol+'}')]);
 mutations.push([pin.owner+': half-adapted tree',raw.replace(before,adapted)]);
}
const dedup=adapter.owners.find(p=>p.owner==='deduplicateSorted');
const dedupBefore=spell(dedup.before),dedupAfter=spell(dedup.after);
mutations.push(['deduplicateSorted: first mint only',raw.replace(dedupBefore,dedupBefore.replace('emptyArray as any as','emptyArray as unknown as'))]);
mutations.push(['deduplicateSorted: second mint only',raw.replace(dedupBefore,dedupBefore.replace('deduplicated as any as','deduplicated as unknown as'))]);
mutations.push(['parse error',raw+'\nfunction {']);
for(const [name,text] of mutations){assert.notEqual(text,raw,name+' failed to mutate');seed(text);assert.throws(()=>adapter.adapt(dir),undefined,name);assert.equal(writes,0,name);assert.equal(fs.readFileSync(file,'utf8'),text,name);}
const foreign=raw.replace('function selectIndex(_: unknown, i: number)', 'function selectIndex(_foreign: unknown, i: number)');assert.notEqual(foreign,raw);seed(foreign);assert.equal(adapter.adapt(dir).edits,edits);assert(fs.readFileSync(file,'utf8').includes('selectIndex(_foreign: unknown'));
for(const lineEnding of ['\n','\r\n']){seed(raw.replace(/\r\n/g,'\n').replace(/\n/g,lineEnding));assert.equal(adapter.adapt(dir).edits,edits);const result=fs.readFileSync(file,'utf8');if(lineEnding==='\n')assert(!result.includes('\r'));else assert(!result.replace(/\r\n/g,'').includes('\n'));}
console.log(JSON.stringify({firstRun:{files:1,edits,writes:1},secondRun:{files:0,edits:0,writes:0,bytesAndMtimePreserved:true},refusals:mutations.map(([name])=>name),foreignEditPreserved:true,crlfAndLfPreserved:true},null,2));
