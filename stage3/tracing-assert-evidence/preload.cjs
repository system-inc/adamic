'use strict';
const fs=require('node:fs'),path=require('node:path');
globalThis.__baobabTraceAssert=function(value){
 const boolean=Boolean(value),double=!!value,beforeFails=!value,afterFails=!double;
 fs.appendFileSync(path.join(process.env.BAOBAB_COUNTER_DIR,`${process.pid}.jsonl`),JSON.stringify({boolean,double,beforeFails,afterFails,difference:boolean!==double||beforeFails!==afterFails})+'\n');
};
const vm=require('node:vm'),hit=globalThis.__baobabTraceAssert;
let total=0,differences=0;
function count(value){total++;if(Boolean(value)!==!!value||(!value)!==(!(!!value)))differences++;return hit(value);}
globalThis.__baobabTraceAssert=count;
function inject(context){if(context&&typeof context==='object')Object.defineProperty(context,'__baobabTraceAssert',{value:count,configurable:true});return context;}
const create=vm.createContext;vm.createContext=function(context={},...args){return Reflect.apply(create,this,[inject(context),...args]);};
const run=vm.runInNewContext;vm.runInNewContext=function(code,context={},...args){return Reflect.apply(run,this,[code,inject(context),...args]);};
const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(context={},...args){return Reflect.apply(script,this,[inject(context),...args]);};
process.on('exit',()=>{if(process.env.BAOBAB_COUNTER_DIR)fs.writeFileSync(path.join(process.env.BAOBAB_COUNTER_DIR,process.pid+'.json'),JSON.stringify({pid:process.pid,argv:process.argv,total,differences})+'\n');});
