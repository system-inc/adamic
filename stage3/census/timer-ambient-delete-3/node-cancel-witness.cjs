'use strict';
const assert=require('node:assert/strict'),path=require('node:path');
const ts=require(path.join(path.resolve(process.argv[2]),'built/local/typescript.js'));
assert.equal(typeof ts.sys.clearTimeout,'function');
(async()=>{
 let fires=0;
 const cancelled=setTimeout(()=>{fires++},20);
 ts.sys.clearTimeout(cancelled);
 const observations=[];
 for(const [kind,value] of [['string','not-a-timer-id'],['plain-object',{}],['undefined',undefined]]){
  const result=ts.sys.clearTimeout(value);
  assert.equal(result,undefined);
  observations.push({kind,result:'undefined',threw:false});
 }
 await new Promise(resolve=>setTimeout(resolve,50));
 assert.equal(fires,0,'real handle was not cancelled');
 console.log(JSON.stringify({realHandleCancelled:true,ignored:observations}));
})().catch(e=>{console.error(e);process.exitCode=1});
