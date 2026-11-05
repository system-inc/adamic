const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const {transpile}=require('./adapt.cjs');
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
function files(root) { return fs.readdirSync(root,{recursive:true}).filter(f=>fs.statSync(path.join(root,f)).isFile()).sort(); }
const [before,after,replay]=process.argv.slice(2);
assert(before && after && replay, 'usage: node compare.cjs BEFORE AFTER REPLAY');
const names=files(before+'/src'); assert.deepEqual(names,files(after+'/src'));assert.deepEqual(names,files(replay+'/src'));
let changed=[];
for(const name of names) { const a=fs.readFileSync(before+'/src/'+name),b=fs.readFileSync(after+'/src/'+name); assert(b.equals(fs.readFileSync(replay+'/src/'+name)),name+': replay drift'); if(!a.equals(b)) changed.push(name); }
assert.deepEqual(changed,['compiler/core.ts']);
const a=fs.readFileSync(before+'/src/compiler/core.ts','utf8'), b=fs.readFileSync(after+'/src/compiler/core.ts','utf8');
assert.equal(a.replace('const result = [];\r\n    for (let i = 0; i < array.length; i++) {\r\n        const v = array[i];','const result: T[] = [];\r\n    for (let i = 0; i < array.length; i++) {\r\n        const v = array[i];'),b);
const js=[];
for(const removeComments of [false,true]) { const x=transpile(a,removeComments),y=transpile(b,removeComments); assert.equal(x,y);js.push({removeComments,byteIdentical:true,sha256:hash(x),bytes:Buffer.byteLength(x)}); }
const declarations=['typescript.d.ts','tsserverlibrary.d.ts'].map(name=>{ const a=fs.readFileSync(before+'/built/local/'+name), b=fs.readFileSync(after+'/built/local/'+name);assert(a.equals(b)); return {name,byteIdentical:true,bytes:a.length,sha256:hash(a)}; });
const emittedNames=files(before+'/built/local').filter(name=>/\.(?:c|m)?js$/.test(name));
assert.deepEqual(emittedNames,files(after+'/built/local').filter(name=>/\.(?:c|m)?js$/.test(name)));
for (const name of emittedNames) assert(fs.readFileSync(before+'/built/local/'+name).equals(fs.readFileSync(after+'/built/local/'+name)), 'emitted JavaScript differs: '+name);
console.log(JSON.stringify({emittedJavascriptFiles:emittedNames.length,emittedJavascriptByteIdentical:true,sourceFiles:names.length,changedSources:changed,onlyEdit:'const result: T[] = [];',replaySourceBytesIdentical:true,coreJavascript:js,wholeAdaptedTreeJavascriptByteIdentical:true,wholeTreeComparison:'All src files other than core.ts are byte-identical. Both core.ts transpiles are byte-identical.',oracle:'skipped because adapted tree JavaScript is byte-identical',declarations},null,2));
