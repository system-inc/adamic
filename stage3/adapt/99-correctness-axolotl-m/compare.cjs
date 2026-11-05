'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const {transpile}=require('./adapt.cjs');
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
function files(root){return fs.readdirSync(root,{recursive:true}).filter(f=>fs.statSync(path.join(root,f)).isFile()).sort();}
const [before,after,replay,census]=process.argv.slice(2);assert(before&&after&&replay&&census,'usage: node compare.cjs BEFORE AFTER REPLAY CENSUS_TREE');
const names=files(before+'/src');for(const tree of [after,replay,census])assert.deepEqual(names,files(tree+'/src'));
const changed=[];
for(const name of names){const a=fs.readFileSync(before+'/src/'+name),b=fs.readFileSync(after+'/src/'+name);assert(b.equals(fs.readFileSync(replay+'/src/'+name)),'replay: '+name);assert(b.equals(fs.readFileSync(census+'/src/'+name)),'census snapshot: '+name);if(!a.equals(b))changed.push(name);}
assert.deepEqual(changed,['compiler/core.ts']);
const a=fs.readFileSync(before+'/src/compiler/core.ts','utf8'),b=fs.readFileSync(after+'/src/compiler/core.ts','utf8');assert.equal(a.replace('result = append(result, v);','result = append<U>(result, v);'),b);
const js=[];for(const removeComments of [false,true]){const x=transpile(a,removeComments),y=transpile(b,removeComments);assert.equal(x,y);js.push({removeComments,byteIdentical:true,sha256:hash(x),bytes:Buffer.byteLength(x)});}
const declarations=['typescript.d.ts','tsserverlibrary.d.ts'].map(name=>{const a=fs.readFileSync(before+'/built/local/'+name),b=fs.readFileSync(after+'/built/local/'+name);assert(a.equals(b));return {name,byteIdentical:true,bytes:a.length,sha256:hash(a)};});
const emittedNames=files(before+'/built/local').filter(name=>/\.(?:c|m)?js$/.test(name));assert.deepEqual(emittedNames,files(after+'/built/local').filter(name=>/\.(?:c|m)?js$/.test(name)));
const emitted=emittedNames.map(name=>{const a=fs.readFileSync(before+'/built/local/'+name),b=fs.readFileSync(after+'/built/local/'+name);assert(a.equals(b),'JavaScript differs: '+name);return {name,bytes:a.length,sha256:hash(a)};});
console.log(JSON.stringify({sourceFiles:names.length,changedSources:changed,onlyEdit:'append<U>(result, v)',replaySourceBytesIdentical:true,censusSnapshotSourceBytesIdentical:true,coreJavascript:js,emittedJavascriptFiles:emitted.length,emittedJavascriptByteIdentical:true,emitted,declarations,oracle:'skipped because all emitted JavaScript and the complete source tree JavaScript are byte-identical'},null,2));
