"use strict";
const fs=require("node:fs"),vm=require("node:vm"),assert=require("node:assert/strict"),ts=require("typescript");
const adapter=require("./adapt.cjs");
const product=require("/workspace/scratch/debug-messages/after/built/local/typescript.js");
function load(owner){const output=ts.transpileModule(owner,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.CommonJS}}).outputText;const exports={};vm.runInNewContext(output,{exports,Debug:product.Debug},{filename:"cast-witness.cjs"});return exports.cast;}
const before=load(adapter.pins.before),after=load(adapter.pins.after),rows=[];
for(const [kind,make] of [["undefined",()=>undefined],["null",()=>null],["number",()=>42],["string",()=>"text"],["object",()=>({})],["custom stringification",()=>({toString(){return "value kept";}})]]){
 const value=make();function invoke(fn){let predicateCalls=0;function reject(){predicateCalls++;return false;}let error;try{fn(value,reject);}catch(e){error=e;}assert(error instanceof Error);return {error,predicateCalls,throwPoint:error.stack.split("\n")[1]};}
 const a=invoke(before),b=invoke(after);assert.equal(a.error.constructor,b.error.constructor);assert.equal(a.predicateCalls,b.predicateCalls);assert.equal(a.error.message.replace(" 'reject'",""),b.error.message);assert.equal(a.throwPoint,b.throwPoint);assert.equal(a.predicateCalls,value===undefined?0:1);rows.push({kind,before:a.error.message,after:b.error.message,errorClass:b.error.constructor.name,predicateCalls:b.predicateCalls,throwPoint:b.throwPoint});
}
for(const value of [null,42,"text",{}]){let calls=0;assert.equal(before(value,()=>{calls++;return true;}),value);assert.equal(after(value,()=>{calls++;return true;}),value);assert.equal(calls,2);}
const result={passed:true,method:"Pinned owner source transpiled in scratch against the unchanged existing Debug.fail; not a completed adapted-tree build witness",failureClasses:rows,passingIdentityCases:4};fs.writeFileSync("/workspace/scratch/core-cast/raw/source-node-witness.json",JSON.stringify(result,null,2)+"\n");console.log("six failing value classes retain Error, predicate count, value text and throw point; four passing cases retain identity");
