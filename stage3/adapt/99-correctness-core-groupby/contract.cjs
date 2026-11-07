'use strict';
const fs=require('fs'),os=require('os'),path=require('path'),assert=require('assert/strict');
const adapter=require('./adapt.cjs'),pin=adapter.owners[0];
const raw=fs.readFileSync(process.argv[2]+'/src/compiler/core.ts','utf8'),dir=fs.mkdtempSync(path.join(os.tmpdir(),'core-contract-'));
fs.mkdirSync(dir+'/src/compiler',{recursive:true});const file=dir+'/src/compiler/core.ts',write=fs.writeFileSync;let writes=0;fs.writeFileSync=(...a)=>{writes++;return write(...a);};
const seed=t=>{write(file,t);writes=0;},spell=t=>raw.includes('\r\n')?t.replace(/\n/g,'\r\n'):t;
seed(raw);assert.deepEqual(adapter.adapt(dir),{files:1,edits:1,javascriptChanging:false});assert.equal(writes,1);
const after=fs.readFileSync(file,'utf8'),mtime=fs.statSync(file,{bigint:true}).mtimeNs;writes=0;assert.deepEqual(adapter.adapt(dir),{files:0,edits:0,javascriptChanging:false});assert.equal(writes,0);assert.equal(fs.statSync(file,{bigint:true}).mtimeNs,mtime);
const beforeOwner=spell(pin.before),afterOwner=spell(pin.after),mutations=[
 ['foreign owner drift',raw.replace(beforeOwner,beforeOwner.replace('{','{ /* drift */'))],
 ['adapted owner drift',after.replace(afterOwner,afterOwner.replace('{','{ /* drift */'))],
 ['missing owner',raw.replace(beforeOwner,'')],['duplicate owner',raw+'\n'+beforeOwner],
 ['nested duplicate',raw+'\nfunction foreignContainer() {\n'+beforeOwner+'\n}'],
 ['moved owner',raw.replace(beforeOwner,'function foreignContainer() {\n'+beforeOwner+'\n}')],
 ['parse error',raw+'\nfunction {']
];
if(pin.owner==='groupBy')mutations.push(['half-adapted return',raw.replace('return result as { [P in K as `${P}`]?: T[]; };','return result;')]);
for(const [label,t]of mutations){assert.notEqual(t,raw);seed(t);assert.throws(()=>adapter.adapt(dir),undefined,label);assert.equal(writes,0);assert.equal(fs.readFileSync(file,'utf8'),t);}
for(const eol of ['\n','\r\n']){seed(raw.replace(/\r\n/g,'\n').replace(/\n/g,eol));adapter.adapt(dir);const a=fs.readFileSync(file,'utf8');assert(eol==='\n'?!a.includes('\r'):!a.replace(/\r\n/g,'').includes('\n'));}
const suffix=spell('\n// outside owner witness\n');seed(raw+suffix);adapter.adapt(dir);assert(fs.readFileSync(file,'utf8').endsWith(suffix));
fs.writeFileSync=write;fs.rmSync(dir,{recursive:true});console.log(JSON.stringify({firstRun:{files:1,edits:1,writes:1},replay:{files:0,edits:0,writes:0,mtimePreserved:true},noWriteRefusals:mutations.map(x=>x[0]),lineEndingsPreserved:true,outsideOwnerPreserved:true},null,2));
