'use strict';
const fs=require('node:fs'),os=require('node:os'),path=require('node:path'),assert=require('node:assert/strict');
function observe(tree,started){
 const file=path.resolve(tree,'built/local/typescript.js');delete require.cache[file];const ts=require(file);
 ts.Debug.setAssertionLevel(1);
 const directory=fs.mkdtempSync(path.join(os.tmpdir(),'tracing-assert-'));
 try{
  if(started)ts.tracingEnabled.startTracing('server',directory);
  try{ts.tracingEnabled.stopTracing();return {outcome:'returned',message:null};}
  catch(error){return {outcome:'threw',message:error.message};}
 }finally{fs.rmSync(directory,{recursive:true,force:true});}
}
const [before,after]=process.argv.slice(2);assert(before&&after);
for(const started of [true,false]){const b=JSON.stringify(observe(before,started)),a=JSON.stringify(observe(after,started));console.log(JSON.stringify({started,before:JSON.parse(b),after:JSON.parse(a),byteIdentical:b===a}));assert.equal(a,b);assert.equal(JSON.parse(a).outcome,started?'returned':'threw');if(!started)assert.equal(JSON.parse(a).message,'Debug Failure. False expression: Tracing is not in progress');}
