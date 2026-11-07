'use strict';const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const observe=globalThis.__sysRecordReplay;assert.equal(typeof observe,'function');
const input=[undefined,false,0,'',null,true,1,'yes',{},()=>{}];for(const v of input)assert.equal(!!observe(v,{recordreplay:v}),!!v);
let reads=0;const host={get recordreplay(){reads++;return true;}};assert.equal(!!observe(host.recordreplay,host),true);assert.equal(reads,1);
const before=reads;assert.equal(true||!!observe(host.recordreplay,host),true);assert.equal(reads,before);
for(const mode of ['create','run','script']){const context={host:{recordreplay:'vm'}};const code='!!globalThis.__sysRecordReplay(host.recordreplay, host)';const value=mode==='create'?vm.runInContext(code,vm.createContext(context)):mode==='run'?vm.runInNewContext(code,context):new vm.Script(code).runInNewContext(context);assert.equal(value,true);}
console.log('PASS: truthy and falsy values, one accessor read, short circuit, three VM paths');
