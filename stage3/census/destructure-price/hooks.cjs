"use strict";
const fs=require('fs'),path=require('path'),{threadId}=require('worker_threads');
const counts={enter:{},calls:{}};
globalThis.__dpEnter=id=>{counts.enter[id]=(counts.enter[id]||0)+1};
globalThis.__dpCount=(id,value)=>{
 if(typeof value!=='function')return value;
 const counted=function(...args){counts.calls[id]=(counts.calls[id]||0)+1;return Reflect.apply(value,this,args)};
 Object.defineProperty(counted,'name',{value:value.name,configurable:true});
 Object.defineProperty(counted,'length',{value:value.length,configurable:true});
 return counted;
};
globalThis.__dpSnapshot=()=>JSON.parse(JSON.stringify(counts));
process.once('exit',()=>{if(process.env.DP_DIR){fs.mkdirSync(process.env.DP_DIR,{recursive:true});fs.writeFileSync(path.join(process.env.DP_DIR,process.pid+'-'+threadId+'.json'),JSON.stringify({pid:process.pid,threadId,argv:process.argv,...counts}));}});
