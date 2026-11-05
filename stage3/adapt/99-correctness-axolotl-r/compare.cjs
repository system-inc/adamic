'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const {owners,transpile,parse}=require('./adapt.cjs');
const [before,after]=process.argv.slice(2);assert(before&&after);
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
function files(root){return fs.readdirSync(root,{recursive:true}).filter(f=>fs.statSync(path.join(root,f)).isFile()).sort();}
const names=files(before+'/src');assert.deepEqual(names,files(after+'/src'));const changed=[];
for(const name of names)if(!fs.readFileSync(before+'/src/'+name).equals(fs.readFileSync(after+'/src/'+name)))changed.push(name);
assert.deepEqual(changed,['compiler/checker.ts','compiler/core.ts']);
for(const file of ['core.ts','checker.ts']) {
 let expected=fs.readFileSync(before+'/src/compiler/'+file,'utf8').replace(/\r\n/g,'\n');
 for(const pin of owners.filter(p=>p.file===file)){assert(expected.includes(pin.before),pin.owner);expected=expected.replace(pin.before,pin.after);}
 const actual=fs.readFileSync(after+'/src/compiler/'+file,'utf8').replace(/\r\n/g,'\n');assert.equal(actual,expected);parse(file,actual);
}
const core=Object.fromEntries(['before','after'].map((label,i)=>[label,fs.readFileSync([before,after][i]+'/src/compiler/core.ts','utf8')]));
const js=[false,true].map(removeComments=>{const a=transpile(core.before,removeComments),b=transpile(core.after,removeComments);assert.equal(a,b);return {removeComments,byteIdentical:true,sha256:hash(a),bytes:Buffer.byteLength(a)};});
const declarations=['typescript.d.ts','tsserverlibrary.d.ts'].map(name=>{const a=fs.readFileSync(before+'/built/local/'+name),b=fs.readFileSync(after+'/built/local/'+name);assert(a.equals(b),name);assert(!a.includes('relativeComplement'));return {name,byteIdentical:true,sha256:hash(a),bytes:a.length};});
console.log(JSON.stringify({sourceFiles:names.length,changedSources:changed,onlyPinnedOwnersChanged:true,stockTypescript:'6.0.3',coreJavascript:js,declarations},null,2));
