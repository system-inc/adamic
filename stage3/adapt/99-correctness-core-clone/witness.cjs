'use strict';
const fs=require('fs'),vm=require('vm'),assert=require('assert/strict'),ts=require('typescript');
const {owners,transpile}=require(process.argv[2]+'/adapt.cjs'),pin=owners[0];
function load(text){const js=transpile(text.replaceAll('export function','function'));return vm.runInNewContext(js+'\n'+pin.owner,{hasOwnProperty:Object.prototype.hasOwnProperty});}
const before=load(pin.before),after=load(pin.after);assert.equal(transpile(pin.before),transpile(pin.after));
const cases=pin.owner==='groupBy'?[
 ['undefined',()=>undefined,String],['empty',()=>[],String],['booleans',()=>[0,1,2,3],v=>v%2===0],['string keys',()=>['a','b','a'],v=>v],['nullable keys',()=>[null,undefined,false,1],v=>v],['sparse',()=>Array(2),String],['prototype key',()=>['__proto__'],v=>v],['constructor key',()=>['constructor'],v=>v]
]:[['object',()=>({a:1,b:'b'})],['inherited',()=>Object.assign(Object.create({inherited:3}),{own:2})],['array',()=>[1,2]],['sparse array',()=>Array(2)],['string',()=> 'abc'],['number',()=> 7],['null',()=>null],['undefined',()=>undefined],['symbol key',()=>({[Symbol.for('s')]:1,a:2})]];
function evaluate(fn,make,selector){try{const input=make(),result=fn(input,selector);return {status:'return',keys:Object.keys(result),json:JSON.stringify(result),prototype:Object.getPrototypeOf(result)===Object.prototype?'Object':Object.getPrototypeOf(result)?.constructor?.name};}catch(e){return {status:'throw',name:e.name,message:e.message};}}
const results=cases.map(([name,make,key])=>{const a=evaluate(before,make,key),b=evaluate(after,make,key);assert.deepEqual(a,b);return {name,before:a,after:b,equal:true};});
console.log(JSON.stringify({owner:pin.owner,javascriptByteIdentical:true,results},null,2));
