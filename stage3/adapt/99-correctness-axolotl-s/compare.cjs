'use strict';
const fs=require('fs'),path=require('path'),assert=require('assert/strict'),crypto=require('crypto'),ts=require('typescript');
const {owners,transpile}=require('./adapt.cjs');const [before,after]=process.argv.slice(2);assert(before&&after);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
function files(root){return fs.readdirSync(root,{recursive:true}).filter(f=>fs.statSync(path.join(root,f)).isFile()).sort();}
const names=files(before+'/src');assert.deepEqual(files(after+'/src'),names);const changed=names.filter(n=>!fs.readFileSync(before+'/src/'+n).equals(fs.readFileSync(after+'/src/'+n)));assert.deepEqual(changed,['compiler/core.ts']);
let expected=fs.readFileSync(before+'/src/compiler/core.ts','utf8').replace(/\r\n/g,'\n');for(const p of owners){assert(expected.includes(p.before),p.owner);expected=expected.replace(p.before,p.after);}assert.equal(expected,fs.readFileSync(after+'/src/compiler/core.ts','utf8').replace(/\r\n/g,'\n'));
const skip=owners.find(p=>p.owner==='skipWhile');for(const comments of [false,true])assert.equal(transpile(skip.before,comments),transpile(skip.after,comments));
const declarations=[];
for(const name of ['typescript.d.ts','tsserverlibrary.d.ts'])for(const dir of ['lib','built/local']){const a=fs.readFileSync(before+'/'+dir+'/'+name),b=fs.readFileSync(after+'/'+dir+'/'+name);assert(a.equals(b),dir+'/'+name);const sf=ts.createSourceFile(name,b.toString(),99,true);assert.equal(sf.parseDiagnostics.length,0);let count=0;function visit(n){if(ts.isIdentifier(n)&&n.text==='skipWhile')count++;ts.forEachChild(n,visit);}visit(sf);assert.equal(count,0,'STOP public skipWhile');declarations.push({file:dir+'/'+name,byteIdentical:true,bytes:a.length,sha256:hash(a),skipWhileIdentifiers:count});}
console.log(JSON.stringify({sourceFiles:names.length,changedSources:changed,onlyPinnedOwnersChanged:true,skipWhileJavascriptIdentical:true,dequeueCastUnchanged:true,declarations},null,2));
