'use strict';
const assert=require('node:assert/strict');
// Use the actual adapted compiler implementation; no production caller is edited.
const {filter}=require(process.argv[2]);
assert.equal(typeof filter,'function');
function countedFilter(site,argument,predicate){globalThis.__adamicFilterFreezeBegin(site);return globalThis.__adamicFilterFreeze(site,filter(argument,predicate),argument);}
const site='CONTROL (scratch only)';
const cases=[['push',a=>a.push(2)],['splice',a=>a.splice(0,1)],['sort',a=>a.sort()],['index',a=>{a[0]=9}],['length',a=>{a.length=0}]];
for(const [name,write] of cases){const argument=[3,1],result=countedFilter(site,argument,x=>typeof x==='number');assert.equal(result,argument);assert(Object.isFrozen(result));
 for(const [alias,array] of [['argument',argument],['result',result]]){let caught;try{write(array)}catch(e){caught=e}assert(caught instanceof TypeError,`${name}/${alias} must throw`);console.log(JSON.stringify({site,writer:name,alias,error:caught.stack}));}}
// Dropped-element path must remain mutable, undefined must count but is not an array freeze.
const a=[1,'x'];const b=countedFilter(site,a,x=>typeof x==='number');assert.notEqual(a,b);assert(!Object.isFrozen(a));assert(!Object.isFrozen(b));b.push(2);
countedFilter(site,undefined,()=>true);
console.log('CONTROL PASS: 7 calls, 6 identities (5 arrays + undefined), 10 caught TypeErrors; nonidentity mutable.');

assert.throws(()=>countedFilter('CONTROL throwing predicate',[1],()=>{throw Error('predicate control')}),/predicate control/);
console.log('ENTRY CONTROL PASS: throwing predicate counted as one call, zero completions.');
