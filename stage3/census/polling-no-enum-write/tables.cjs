'use strict';
const assert = require('node:assert/strict');
const ts = require('../../api/node_modules/typescript');
const pins = require('/workspace/adamic/stage3/adapt/99-correctness-polling-no-enum-write/owners.json')[0];
function make(owner) {
 const text = `enum PollingInterval { Low=250, Medium=500, High=2000 }
 type Levels = {Low:number; Medium:number; High:number};
 function createPollingIntervalBasedLevels(levels:Levels) { return {[PollingInterval.Low]:levels.Low,[PollingInterval.Medium]:levels.Medium,[PollingInterval.High]:levels.High}; }
 const defaultChunkLevels = {Low:32,Medium:64,High:256};
 let pollingChunkSize = createPollingIntervalBasedLevels(defaultChunkLevels);
 let unchangedPollThresholds = createPollingIntervalBasedLevels(defaultChunkLevels);
 ${owner}
 return {configure:setCustomPollingValues, tables:()=>[pollingChunkSize,unchangedPollThresholds]};`;
 return new Function(ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2022}}).outputText)();
}
let cases=0;
for (const interval of ['', '0','NaN','17']) for (const chunk of ['', '0','NaN','19']) for (const threshold of ['', '0','NaN','23']) {
 const env={};
 for(const level of ['LOW','MEDIUM','HIGH']) {
  env['TSC_WATCH_POLLINGINTERVAL_'+level]=interval;
  env['TSC_WATCH_POLLINGCHUNKSIZE_'+level]=chunk;
  env['TSC_WATCH_UNCHANGEDPOLLTHRESHOLDS_'+level]=threshold;
 }
 const b=make(pins.before),a=make(pins.after),bt=b.tables(),at=a.tables(),br=[],ar=[];
 assert.equal(b.configure({getEnvironmentVariable:k=>(br.push(k),env[k])}),undefined);
 assert.equal(a.configure({getEnvironmentVariable:k=>(ar.push(k),env[k])}),undefined);
 assert.deepEqual(ar,br);
 for(let i=0;i<2;i++) {
  // Compare level values: custom intervals intentionally change only old keys.
  const oldValues=Object.values(b.tables()[i]),newValues=Object.values(a.tables()[i]);
  if(interval==='17') { // all three levels collapse to the same old numeric key
   assert.equal(oldValues.at(-1),newValues.at(-1));
  } else assert.deepEqual(newValues,oldValues);
  assert.equal(a.tables()[i]===at[i],b.tables()[i]===bt[i]);
 }
 assert.equal(a.configure({}),undefined);
 cases++;
}
console.log(JSON.stringify({cases,environmentReadOrder:'identical',returnValue:'undefined',tableRebuildIdentity:'identical',numericCases:['empty','zero','NaN','positive']}));
