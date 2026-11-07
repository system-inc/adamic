'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const tree=process.argv[2],records=[];
function files(d){return fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?files(path.join(d,e.name)):/\.(?:js|mjs|cjs)$/.test(e.name)?[path.join(d,e.name)]:[]);}
for(const file of files(tree+'/built/local')){
 const text=fs.readFileSync(file,'utf8');let count=0;
 const next=text.replace(/elementAt\s*:\s*\(\)\s*=>\s*elementAt\b/g,(match,offset)=>{count++;const tag=path.relative(tree,file)+':'+(text.slice(0,offset).split('\n').length);return 'elementAt: () => globalThis.__eaExportRead(() => elementAt, '+JSON.stringify(tag)+')';});
 if(count){fs.writeFileSync(file,next);records.push({file:path.relative(tree,file),getters:count});}
}
assert(records.reduce((n,r)=>n+r.getters,0)>=5);console.log(JSON.stringify({measurementOnly:true,returnsOriginalFunctionIdentity:true,records},null,2));
