'use strict';
const fs=require('node:fs'),os=require('node:os'),path=require('node:path'),assert=require('node:assert/strict');
const adapter=require('./adapt.cjs');
const raw=fs.readFileSync(path.join(process.argv[2],'src/compiler/core.ts'),'utf8');
const dir=fs.mkdtempSync(path.join(os.tmpdir(),'flatmap-contract-'));
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
 const before=spell(pin.before),adapted=spell(pin.after);
 assert(raw.includes(before));assert(after.includes(adapted));
 mutations.push([pin.owner+': changed owner',raw.replace(before,before.replace('let result:','var result:').replace('const result:','var result:'))]);
 mutations.push([pin.owner+': changed adapted owner',after.replace(adapted,adapted.replace('let result:','var result:').replace('const result:','var result:'))]);
 mutations.push([pin.owner+': partly adapted type argument',raw.replace(before,before.replace('append(result, v)', 'append<{}>(result, v)'))]);
 mutations.push([pin.owner+': missing owner',raw.replace(before,'')]);
 mutations.push([pin.owner+': duplicate owner',raw+eol+before]);
 mutations.push([pin.owner+': moved owner',raw.replace(before,'function outer() {'+eol+before+eol+'}')]);
 // Remove one explicit type argument from the planned owner. With two edits this is half adapted.
 const partly=adapted.replace(/(addRange|append)<U>/,'$1');
 if(partly!==before)mutations.push([pin.owner+': half-adapted owner',raw.replace(before,partly)]);
 if(adapter.owners.length>1)mutations.push([pin.owner+': half-adapted tree',raw.replace(before,adapted)]);
}
mutations.push(['parse error',raw+'\nfunction {']);
for(const [name,text] of mutations){assert.notEqual(text,raw,name+' failed to mutate');seed(text);assert.throws(()=>adapter.adapt(dir),undefined,name);assert.equal(writes,0,name);assert.equal(fs.readFileSync(file,'utf8'),text,name);}
seed(raw+eol+'// unrelated outside owner'+eol);assert.equal(adapter.adapt(dir).edits,edits);assert(fs.readFileSync(file,'utf8').endsWith('// unrelated outside owner'+eol));
for(const lineEnding of ['\n','\r\n']){seed(raw.replace(/\r\n/g,'\n').replace(/\n/g,lineEnding));assert.equal(adapter.adapt(dir).edits,edits);const result=fs.readFileSync(file,'utf8');if(lineEnding==='\n')assert(!result.includes('\r'));else assert(!result.replace(/\r\n/g,'').includes('\n'));}
console.log(JSON.stringify({firstRun:{files:1,edits,writes:1},secondRun:{files:0,edits:0,writes:0,bytesAndMtimePreserved:true},refusals:mutations.map(([name])=>name),unrelatedEditPreserved:true,crlfAndLfPreserved:true},null,2));
