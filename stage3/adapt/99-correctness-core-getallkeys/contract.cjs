'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const {adapt,pins}=require('./adapt.cjs');
const [source,scratch,output]=process.argv.slice(2),files=[pins.core.file,pins.harness.file];
fs.mkdirSync(scratch,{recursive:true});
const original=files.map(f=>fs.readFileSync(path.join(source,f),'utf8'));
const snapshot=()=>files.map(f=>({text:fs.readFileSync(path.join(scratch,f),'utf8'),mtime:fs.statSync(path.join(scratch,f)).mtimeMs}));
function write(texts){files.forEach((f,i)=>{fs.mkdirSync(path.dirname(path.join(scratch,f)),{recursive:true});fs.writeFileSync(path.join(scratch,f),texts[i]);});}
const checks=[];
function refuse(label,texts){write(texts);const before=snapshot();assert.throws(()=>adapt(scratch));assert.deepEqual(snapshot(),before,label+': refusal wrote files');checks.push({check:label,pass:true});}
for(const nl of ['\n','\r\n']) {
 const before=original.map(s=>s.replace(/\r\n/g,'\n').replace(/\n/g,nl));write(before);
 assert.equal(adapt(scratch).files,2);const after=snapshot().map(r=>r.text),snap=snapshot();
 assert.equal(adapt(scratch).files,0);assert.deepEqual(snapshot(),snap);checks.push({check:'apply and no-write replay '+JSON.stringify(nl),pass:true});
 refuse('core only moved '+JSON.stringify(nl),[after[0],before[1]]);
 refuse('harness only moved '+JSON.stringify(nl),[before[0],after[1]]);
 for(const state of [before,after]) {
  const changed=state.slice();
  changed[1]=changed[1].replace(state===before?pins.harness.before:pins.harness.after,(state===before?pins.harness.before:pins.harness.after).replace('const keys','const  keys'));
  refuse('whole statement drift '+(state===before?'before':'after')+' '+JSON.stringify(nl),changed);
 }
 const changed=before.slice();changed[0]=changed[0].replace('while (obj = Object.getPrototypeOf(obj));','while (obj = Object.getPrototypeOf(obj)) /* drift */;');refuse('whole core body drift '+JSON.stringify(nl),changed);
 const local=after.slice();local[1]=local[1].replace('while (obj = Object.getPrototypeOf(obj));','while (obj = Object.getPrototypeOf(obj)) /* drift */;');refuse('whole local body drift '+JSON.stringify(nl),local);
 const comment=before.slice();comment[0]=comment[0].replace('/** @internal */'+nl+'export function getAllKeys','/** changed */'+nl+'export function getAllKeys');refuse('core doc drift '+JSON.stringify(nl),comment);
 const partial=before.slice();partial[1]=partial[1].replace(pins.harness.before,pins.harness.after);refuse('call only '+JSON.stringify(nl),partial);
 const importOnly=before.slice();importOnly[1]=importOnly[1].replace(pins.namespace,pins.namespace+nl+pins.import.after);refuse('import only '+JSON.stringify(nl),importOnly);
 const orphan=after.slice();orphan[0]=orphan[0].replace('/** @internal */'+nl+'export function getOwnValues','/** @internal */'+nl+nl+'/** @internal */'+nl+'export function getOwnValues');refuse('orphaned core comment '+JSON.stringify(nl),orphan);
 const duplicate=after.slice();duplicate[1]+=nl+pins.local.after.replace(/\n/g,nl)+nl;refuse('duplicate local '+JSON.stringify(nl),duplicate);
 write(before.map(s=>s+nl+'// unrelated preserved'+nl));adapt(scratch);assert(snapshot().every(r=>r.text.endsWith(nl+'// unrelated preserved'+nl)||r.text.includes(nl+'// unrelated preserved'+nl)));checks.push({check:'foreign edit preserved '+JSON.stringify(nl),pass:true});
}
fs.writeFileSync(output,JSON.stringify({passed:true,checks},null,2)+'\n');console.log(JSON.stringify({passed:true,checks:checks.length}));
