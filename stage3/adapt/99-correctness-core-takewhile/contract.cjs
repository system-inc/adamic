'use strict';
const fs=require('node:fs'),os=require('node:os'),path=require('node:path'),assert=require('node:assert/strict');
const adapter=require('./adapt.cjs'),pin=adapter.owners[0];
const raw=fs.readFileSync(path.join(process.argv[2],'src/compiler/core.ts'),'utf8');
const dir=fs.mkdtempSync(path.join(os.tmpdir(),'takewhile-contract-'));fs.mkdirSync(path.join(dir,'src/compiler'),{recursive:true});
const file=path.join(dir,'src/compiler/core.ts'),originalWrite=fs.writeFileSync;let writes=0;
fs.writeFileSync=(...args)=>{writes++;return originalWrite(...args);};
function seed(text){originalWrite(file,text);writes=0;}
try {
    seed(raw);assert.deepEqual(adapter.adapt(dir),{files:1,edits:1,javascriptChanging:true});assert.equal(writes,1);
    const after=fs.readFileSync(file,'utf8'),stat=fs.statSync(file,{bigint:true});writes=0;
    assert.deepEqual(adapter.adapt(dir),{files:0,edits:0,javascriptChanging:true});assert.equal(writes,0);assert.equal(fs.readFileSync(file,'utf8'),after);assert.equal(fs.statSync(file,{bigint:true}).mtimeNs,stat.mtimeNs);
    const eol=raw.includes('\r\n')?'\r\n':'\n',spell=s=>s.replace(/\n/g,eol),before=spell(pin.before),adapted=spell(pin.after);
    const mutations=[
        ['changed owner',raw.replace(before,before.replace('{','{ /* drift */'))],
        ['changed adapted owner',after.replace(adapted,adapted.replace('{','{ /* drift */'))],
        ['missing owner',raw.replace(before,'')],
        ['duplicate owner',raw+eol+before],
        ['nested duplicate owner',raw+eol+'function foreignContainer() {'+eol+before+eol+'}'],
        ['moved owner',raw.replace(before,'function foreignContainer() {'+eol+before+eol+'}')],
        ['changed overload',raw.replace(before,before.replace('array: readonly T[]','array: T[]'))],
        ['half adapted result declaration',raw.replace(before,before.replace('const len = array.length;','const result: U[] = [];'+eol+'        const len = array.length;'))],
        ['half adapted return',raw.replace(before,before.replace('return array.slice(0, index) as U[];','return result;'))],
        ['parse error',raw+'\nfunction {'],
    ];
    for(const [name,text] of mutations){assert.notEqual(text,raw,name);seed(text);assert.throws(()=>adapter.adapt(dir),undefined,name);assert.equal(writes,0,name);assert.equal(fs.readFileSync(file,'utf8'),text,name);}
    const foreign=raw+'\nfunction unrelatedWitness() { return 17; }\n';seed(foreign);assert.equal(adapter.adapt(dir).edits,1);assert(fs.readFileSync(file,'utf8').endsWith('function unrelatedWitness() { return 17; }\n'));
    for(const ending of ['\n','\r\n']){seed(raw.replace(/\r\n/g,'\n').replace(/\n/g,ending));assert.equal(adapter.adapt(dir).edits,1);const result=fs.readFileSync(file,'utf8');if(ending==='\n')assert(!result.includes('\r'));else assert(!result.replace(/\r\n/g,'').includes('\n'));}
    console.log(JSON.stringify({firstRun:{files:1,edits:1,writes:1},secondRun:{files:0,edits:0,writes:0,bytesAndMtimePreserved:true},refusals:mutations.map(([name])=>name),foreignEditPreserved:true,crlfAndLfPreserved:true},null,2));
} finally { fs.writeFileSync=originalWrite;fs.rmSync(dir,{recursive:true,force:true}); }
