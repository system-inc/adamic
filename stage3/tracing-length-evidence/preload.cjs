'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
const counts={popCall:{},popWrite:{},popAllCall:{},popAllWrite:{}};
function count(site,length){if(!Object.hasOwn(counts,site)||!Number.isInteger(length)||length<0)throw new Error('counter drift');counts[site][length]=(counts[site][length]||0)+1;}
globalThis.__traceLength=count;
function inject(context){if(context&&typeof context==='object')Object.defineProperty(context,'__traceLength',{value:count,configurable:true});return context;}
const create=vm.createContext;vm.createContext=function(context={},...args){return Reflect.apply(create,this,[inject(context),...args]);};
const run=vm.runInNewContext;vm.runInNewContext=function(code,context={},...args){return Reflect.apply(run,this,[code,inject(context),...args]);};
const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(context={},...args){return Reflect.apply(script,this,[inject(context),...args]);};
process.on('exit',()=>{if(process.env.TRACING_LENGTH_COUNTER_DIR)fs.writeFileSync(path.join(process.env.TRACING_LENGTH_COUNTER_DIR,process.pid+'.json'),JSON.stringify({pid:process.pid,argv:process.argv,counts})+'\n');});
