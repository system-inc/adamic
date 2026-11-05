'use strict';
const fs=require('fs'),path=require('path'),assert=require('assert/strict'),crypto=require('crypto'),ts=require('typescript');
const {owners,pins,plan}=require('./adapt.cjs');
const [before,after,measured]=process.argv.slice(2);assert(before&&after&&measured);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const all=root=>fs.readdirSync(path.join(root,'src'),{recursive:true}).filter(f=>fs.statSync(path.join(root,'src',f)).isFile()).sort();
const names=all(before);assert.deepEqual(names,all(after));assert.deepEqual(names,all(measured));const changed=[];
for(const name of names){const a=fs.readFileSync(path.join(before,'src',name)),b=fs.readFileSync(path.join(after,'src',name));assert(b.equals(fs.readFileSync(path.join(measured,'src',name))),name);if(!a.equals(b))changed.push(name);}
assert.deepEqual(changed,['compiler/transformers/declarations/diagnostics.ts','compiler/utilities.ts']);
const texts=new Map([...new Set(pins.map(p=>p.file))].map(f=>[f,fs.readFileSync(path.join(before,'src/compiler',f),'utf8')]));const expected=plan(texts).texts;for(const [file,text]of expected)assert.equal(text,fs.readFileSync(path.join(after,'src/compiler',file),'utf8'));
const declarations=['typescript.d.ts','tsserverlibrary.d.ts'].map(name=>{const a=fs.readFileSync(path.join(before,'built/local',name)),b=fs.readFileSync(path.join(after,'built/local',name));assert(a.equals(b),name);return {name,byteIdentical:true,bytes:a.length,sha256:hash(a)};});
const ownerRows=pins.map(pin=>{const file=path.join(after,'src/compiler',pin.file),o=owners(fs.readFileSync(file,'utf8'),pin.file).get(pin.owner);return {owner:pin.owner,file:pin.file,exported:pin.after.startsWith('export function '),beforeSha256:pin.ownerBeforeSha256,afterSha256:pin.ownerAfterSha256};});
console.log(JSON.stringify({sourceFiles:names.length,changedSourceFiles:changed,onlyPlannedOwnerAndImportEdits:true,measuredAndAppliedSourceByteIdentical:true,declarations,owners:ownerRows,unchangedFiles:['compiler/debug.ts','compiler/moduleSpecifiers.ts','compiler/expressionToTypeNode.ts'].map(file=>({file,sha256:hash(fs.readFileSync(path.join(after,'src',file))),byteIdentical:true}))},null,2));
