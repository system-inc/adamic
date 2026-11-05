'use strict';
const fs=require('fs'),vm=require('vm'),assert=require('assert/strict'),ts=require('typescript');
const {owners}=require('./adapt.cjs');
function load(text,name,instrument=false){
 text=text.replaceAll('export function','function');
 if(instrument)text=text.replace('        headIndex++;','        headIndex++;\n        globalThis.slots = elements;');
 const js=ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2024}}).outputText;
 const context={};vm.createContext(context);vm.runInContext(js+'\nglobalThis.fn = '+name,context);return context;
}
const records=[];
function snapshot(a){return {contents:Array.from(a),length:a.length};}
for(const owner of ['filterMutate','clear']){
 const p=owners.find(p=>p.owner===owner);
 for(const n of [0,2,4]){
  const initial=owner==='clear'?Array.from({length:n},(_,i)=>i+1):[1,2,3,4];
  const results=[];
  for(const text of [p.before,p.after]){const c=load(text,owner),a=initial.slice();const result=owner==='clear'?c.fn(a):c.fn(a,(_v,i)=>i<n);assert.equal(result,undefined);results.push(snapshot(a));}
  assert.deepEqual(results[0],results[1]);records.push({owner,n,initial:snapshot(initial),before:results[0],after:results[1],returnValue:'undefined'});
 }
}
const queue=owners.find(p=>p.owner==='createQueue');
// The exact copyWithin/truncation fragment, including no-op and zero length.
for(const n of [0,2,4]){
 const results=[];
 for(const text of [queue.before,queue.after]){const trunc=text.match(/elements\.(?:length = newLength|splice\(newLength\));/)[0];const c={elements:[1,2,3,4],newLength:n};vm.createContext(c);vm.runInContext('elements.copyWithin(0, elements.length-newLength);'+trunc,c);results.push(snapshot(c.elements));}
 assert.deepEqual(results[0],results[1]);records.push({owner:'createQueue.dequeue truncation fragment',n,initial:{contents:[1,2,3,4],length:4},before:results[0],after:results[1]});
}
// Full queue: undefined payload, object reference release, compaction and empty error.
for(const text of [queue.before,queue.after]){
 const c=load(text,'createQueue',true);const object={id:1};const q=c.fn([object,undefined,...Array.from({length:202},(_,i)=>i)]);
 assert.equal(q.dequeue(),object);assert.equal(c.slots[0],undefined);assert.equal(q.dequeue(),undefined);
 for(let i=0;i<202;i++)assert.equal(q.dequeue(),i);
 assert(q.isEmpty());assert.throws(()=>q.dequeue(),/Queue is empty/);
 const sparse=c.fn(new Array(1));assert.equal(sparse.dequeue(),undefined);
}
const skip=owners.find(p=>p.owner==='skipWhile');const input=[1,'x',2],outputs=[];
for(const text of [skip.before,skip.after])outputs.push(Array.from(load(text,'skipWhile').fn(input,x=>typeof x==='number')));
assert.deepEqual(outputs,[['x',2],['x',2]]);
console.log(JSON.stringify({node:process.version,records,queue:{fifo:true,undefinedPayload:true,consumedSlotUndefined:true,compaction:true,sparseInitialSlotReturnsUndefined:true,emptyThrows:true},skipWhile:{input,predicate:'typeof x === "number"',before:outputs[0],after:outputs[1],laterElementSatisfiesPredicate:true}},null,2));
