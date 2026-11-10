'use strict';
const fs=require('node:fs'),vm=require('node:vm');
const counts={attachFlowNodeDebugInfoWorker:0,attachNodeArrayDebugInfoWorker:0,attachDebugPrototypeIfDebug:0,enableDebugInfo:0};
globalThis.__debugDisplayCount=name=>{if(!(name in counts))throw Error('unreviewed counter '+name);counts[name]++;};
function inject(context){if(context&&typeof context==='object')Object.defineProperty(context,'__debugDisplayCount',{value:globalThis.__debugDisplayCount,configurable:true});return context;}
const create=vm.createContext;vm.createContext=function(context={},...rest){return Reflect.apply(create,this,[inject(context),...rest]);};
const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(context={},...rest){return Reflect.apply(script,this,[inject(context),...rest]);};
for(const name of ['runInNewContext','runInContext']){const original=vm[name];vm[name]=function(code,context,...rest){context=inject(context||{});return Reflect.apply(original,this,[code,context,...rest]);};}
process.on('exit',()=>{if(process.env.DEBUG_DISPLAY_COUNTS)fs.appendFileSync(process.env.DEBUG_DISPLAY_COUNTS,JSON.stringify({pid:process.pid,counts})+'\n');});
