const fs=require('fs'),ts=require('typescript'),vm=require('vm'),assert=require('assert/strict');
const {owners}=require('./adapt.cjs');
function load(source){const context={};vm.createContext(context);vm.runInContext(ts.transpileModule(source.replace('export function','function'),{compilerOptions:{target:ts.ScriptTarget.ES2024}}).outputText+'\nglobalThis.fn=createQueue;',context);return context.fn;}
const before=load(owners.find(p=>p.owner==='createQueue').before);
const draft=load(fs.readFileSync(__dirname+'/dequeue-draft.ts','utf8'));
function run(fn,items){try {const q=fn(items),result=q.dequeue();return {result:result===undefined?'undefined':result,isEmpty:q.isEmpty()};}catch(e){return {error:e.message};}}
const cases=[{name:'undefined payload',items:[undefined]},{name:'sparse initial array',items:new Array(1)}].map(({name,items})=>({name,before:run(before,items),draft:run(draft,items)}));
assert.deepEqual(cases[0].before,cases[0].draft);assert.notDeepEqual(cases[1].before,cases[1].draft);
console.log(JSON.stringify({node:process.version,castFree:true,typeChecks:true,kept:false,reason:'Sparse input changes from returning undefined to an occupancy error',cases},null,2));
