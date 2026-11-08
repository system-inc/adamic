'use strict';
const fs=require('node:fs');
const inputs=[['number',2147483646],['string','not-a-timer'],['object',{}],['undefined',undefined]];
const results=inputs.map(([kind,value])=>{try{const result=clearTimeout(value);return {kind,threw:false,returnType:typeof result};}catch(error){return {kind,threw:true,error:String(error)};}});
fs.writeFileSync(process.argv[2],JSON.stringify({node:process.version,results},null,2)+'\n');
if(results.some(x=>x.threw))process.exitCode=1;
console.log(JSON.stringify(results));
