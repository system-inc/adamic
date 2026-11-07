'use strict';
const fs=require('fs'),path=require('path'),os=require('os'),assert=require('assert/strict'),a=require('./adapt.cjs');
const root=fs.mkdtempSync(path.join(os.tmpdir(),'core-extend-contract-')),source=fs.readFileSync(path.join(process.argv[2],'src/compiler/core.ts'),'utf8').replace(/\r\n/g,'\n');
function tree(name,text=source){const d=path.join(root,name);fs.mkdirSync(path.join(d,'src/compiler'),{recursive:true});fs.writeFileSync(path.join(d,'src/compiler/core.ts'),text);return d;}
function snap(d){const p=path.join(d,'src/compiler/core.ts');return [fs.readFileSync(p,'utf8'),String(fs.statSync(p,{bigint:true}).mtimeNs)];}
const tests=[];
function refused(name,text){const d=tree(name,text),before=snap(d),write=fs.writeFileSync;let writes=0;fs.writeFileSync=(...args)=>{writes++;return write(...args)};try{assert.throws(()=>a.adapt(d));assert.equal(writes,0);assert.deepEqual(snap(d),before);}finally{fs.writeFileSync=write;}tests.push(name);}
for(const pin of a.owners){
 refused('mixed-'+pin.owner,source.replace(pin.before,pin.after));
 refused('missing-'+pin.owner,source.replace(pin.before,''));
 refused('duplicate-'+pin.owner,source+'\n'+pin.before);
 refused('moved-'+pin.owner,source.replace(pin.before,'namespace moved {\n'+pin.before+'\n}'));
 refused('drift-'+pin.owner,source.replace(pin.before,pin.before.replace('(', '(/* drift */')));
}
refused('half-cast',source.replace('const result: T1 & T2 = {} as any;','const result: T1 & T2 = {};'));
refused('parse-error',source+'\nfunction {');
for(const newline of ['\n','\r\n']){
 const foreign='\nfunction unrelatedCoreExtend() { return 1; }\n',d=tree(newline==='\n'?'lf':'crlf',(source+foreign).replace(/\n/g,newline));
 const first=a.adapt(d),state=snap(d),write=fs.writeFileSync;let writes=0;fs.writeFileSync=(...args)=>{writes++;return write(...args)};
 try{assert.equal(a.adapt(d).edits,0);assert.equal(writes,0);assert.deepEqual(snap(d),state);}finally{fs.writeFileSync=write;}
 assert(state[0].endsWith(foreign.replace(/\n/g,newline)));assert.equal(first.edits,2);if(newline==='\r\n')assert(!/(?<!\r)\n/.test(state[0]));
 tests.push(newline==='\n'?'LF, foreign edit, zero-write replay':'CRLF, foreign edit, zero-write replay');
}
fs.rmSync(root,{recursive:true,force:true});console.log(JSON.stringify({pass:true,tests},null,2));
