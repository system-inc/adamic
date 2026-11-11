'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const tree=process.argv[2],changes=[];
for(const name of fs.readdirSync(path.join(tree,'built/local')).filter(x=>x.endsWith('.js'))){
 const file=path.join(tree,'built/local',name),text=fs.readFileSync(file,'utf8');let after=text,sites=[];
 for(const [label,re] of [
 ['popCall',/(function pop\d*\(results\) \{\n)(\s*)(Debug\d*\.assert\((eventStack\d*)\.length > 0\);)/g],
 ['popAllCall',/(function popAll\d*\(\) \{\n)(\s*)(const endTime = (?:1000|1e3) \* timestamp\d*\(\);)/g]
 ]){
 let count=0;after=after.replace(re,(all,open,indent,stmt,stack)=>{count++;if(label==='popAllCall'){const m=after.match(/(?:var|const|let) (eventStack\d*) = \[\];/);assert(m);stack=m[1];}return open+indent+'globalThis.__traceLength('+JSON.stringify(label)+', '+stack+'.length);\n'+indent+stmt;});if(count){assert.equal(count,1,name+label);sites.push(label);}
 }
 for(const [label,re] of [['popWrite',/(eventStack\d*)\.(?:length--|splice\(\1\.length - 1\));/g],['popAllWrite',/(eventStack\d*)\.(?:length = 0|splice\(0\));/g]]){
 let count=0;after=after.replace(re,(all,stack)=>{count++;return 'globalThis.__traceLength('+JSON.stringify(label)+', '+stack+'.length); '+all;});if(count){assert.equal(count,1,name+label);sites.push(label);}
 }
 if(sites.length){assert.equal(sites.length,4,name);fs.writeFileSync(file,after);changes.push({name,sites});}
}
assert(changes.some(x=>x.name==='run.js'));assert(changes.some(x=>x.name==='typescript.js'));console.log(JSON.stringify(changes));
