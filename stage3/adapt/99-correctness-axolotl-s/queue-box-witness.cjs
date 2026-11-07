'use strict';
const fs=require('fs'),vm=require('vm'),assert=require('assert/strict'),ts=require('typescript');
const {owners}=require('./adapt.cjs');
function load(source){const context={};vm.createContext(context);vm.runInContext(ts.transpileModule(source.replace('export function','function'),{compilerOptions:{target:ts.ScriptTarget.ES2024}}).outputText+'\nglobalThis.fn=createQueue;',context);return context.fn;}
const stock=load(owners.find(p=>p.owner==='createQueue').before),boxed=load(fs.readFileSync(__dirname+'/queue-box-proposed.ts','utf8'));
const object={id:1},cases=[];
for(const [name,items] of [['no input',undefined],['empty',[]],['undefined payload',[undefined]],['sparse initial array',new Array(1)],['mixed holes',[1,,undefined,object]],['dense strings',['first','second']],['dense objects',[object]],['compaction',Array.from({length:405},(_,i)=>i)]]){
 const a=stock(items),b=boxed(items),results=[],adaptedResults=[];
 const initialEmpty={before:a.isEmpty(),after:b.isEmpty()};assert.equal(initialEmpty.before,initialEmpty.after);
 for(const value of [undefined,object,1]){assert.equal(a.enqueue(value),b.enqueue(value));}
 while(!a.isEmpty()){assert.equal(a.isEmpty(),b.isEmpty());const x=a.dequeue(),y=b.dequeue();assert(Object.is(x,y));results.push(x===undefined?'undefined':x);adaptedResults.push(y===undefined?'undefined':y);}
 assert.equal(a.isEmpty(),b.isEmpty());for(const q of [a,b])assert.throws(()=>q.dequeue(),/Queue is empty/);
 cases.push({name,dequeues:results.length,initialEmpty,before:results,after:adaptedResults,equal:true,emptyErrorEqual:true});
}
console.log(JSON.stringify({node:process.version,cases,counterexamples:0},null,2));
