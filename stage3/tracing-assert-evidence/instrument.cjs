'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const tree=process.argv[2],changes=[];
for(const name of fs.readdirSync(path.join(tree,'built/local')).filter(x=>x.endsWith('.js'))){
 const file=path.join(tree,'built/local',name),text=fs.readFileSync(file,'utf8');
 const re=/Debug(?:\d*)\.assert\((!!)?(tracing(?:\d*)?), "Tracing is not in progress"\)/g;
 let count=0;const after=text.replace(re,(all,double,value)=>{count++;return all.replace('('+(double||'')+value+',','( (globalThis.__baobabTraceAssert('+value+'), '+(double||'')+value+'),');});
 if(count){assert.equal(count,1,name);fs.writeFileSync(file,after);changes.push({name,count});}
}
assert(changes.some(x=>x.name==='run.js'));assert(changes.some(x=>x.name==='typescript.js'));console.log(JSON.stringify(changes));
