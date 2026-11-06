'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const {transpile,parse}=require('./adapt.cjs');
const owners=require('./owners.json').filter(p=>p.owner==='toSorted').map(p=>({...p,before:p.after.replace('emptyArray as readonly T[]','emptyArray')}));
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
function files(root){return fs.readdirSync(root,{recursive:true}).filter(f=>fs.statSync(path.join(root,f)).isFile()).sort();}
const [before,after,census]=process.argv.slice(2);assert(before&&after&&census,'usage: node compare.cjs BEFORE AFTER CENSUS_TREE');
const names=files(before+'/src');for(const tree of [after,census])assert.deepEqual(names,files(tree+'/src'));
const changed=[];
for(const name of names){const a=fs.readFileSync(before+'/src/'+name),b=fs.readFileSync(after+'/src/'+name);assert(b.equals(fs.readFileSync(census+'/src/'+name)),'census snapshot: '+name);if(!a.equals(b))changed.push(name);}
assert.deepEqual(changed,['compiler/core.ts']);
const a=fs.readFileSync(before+'/src/compiler/core.ts','utf8'),b=fs.readFileSync(after+'/src/compiler/core.ts','utf8');let expected=a.replace(/\r\n/g,'\n');for(const pin of owners){assert(expected.includes(pin.before));expected=expected.replace(pin.before,pin.after);}assert.equal(expected,b.replace(/\r\n/g,'\n'));
const counts=s=>{const ts=require('typescript'),n={casts:0,any:0};function visit(node){if(ts.isAsExpression(node)||ts.isTypeAssertionExpression(node))n.casts++;if(node.kind===ts.SyntaxKind.AnyKeyword)n.any++;ts.forEachChild(node,visit);}visit(parse('core.ts',s));return n;};const ca=counts(a),cb=counts(b);assert.equal(cb.casts,ca.casts+1);assert.equal(cb.any,ca.any);
const js=[];for(const removeComments of [false,true]){const x=transpile(a,removeComments),y=transpile(b,removeComments);assert.equal(x,y);js.push({removeComments,byteIdentical:true,sha256:hash(x),bytes:Buffer.byteLength(x)});}
const declarations=['typescript.d.ts','tsserverlibrary.d.ts'].map(name=>{const a=fs.readFileSync(before+'/built/local/'+name),b=fs.readFileSync(after+'/built/local/'+name);assert(a.equals(b),name+' public declaration differs');return {name,byteIdentical:true,bytes:a.length,sha256:hash(a)};});
const emittedNames=files(before+'/built/local').filter(name=>/\.(?:c|m)?js$/.test(name));assert.deepEqual(emittedNames,files(after+'/built/local').filter(name=>/\.(?:c|m)?js$/.test(name)));
const emitted=emittedNames.map(name=>{const a=fs.readFileSync(before+'/built/local/'+name),b=fs.readFileSync(after+'/built/local/'+name);assert(a.equals(b),'JavaScript differs: '+name);return {name,bytes:a.length,sha256:hash(a)};});
console.log(JSON.stringify({sourceFiles:names.length,changedSources:changed,onlyPinnedOwnersChanged:true,censusSnapshotSourceBytesIdentical:true,syntaxBefore:ca,syntaxAfter:cb,coreJavascript:js,emittedJavascriptFiles:emitted.length,emittedJavascriptByteIdentical:true,emitted,declarations,oracle:'skipped under the type-only exception'},null,2));
