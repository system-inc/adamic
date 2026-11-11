'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
let visits=0,calls=0;
function start(count){calls++;}
function record(type,objectFlagsCandidate,intrinsicCandidate){
 visits++;
 const original=type.objectFlags, intrinsic=type.intrinsicName;
 fs.appendFileSync(path.join(process.env.DUMP_COUNTER_DIR,process.pid+'.jsonl'),JSON.stringify({id:type.id,flags:type.flags,objectOwn:Object.hasOwn(type,'objectFlags'),objectValue:original===undefined?{undefined:true}:original,intrinsicOwn:Object.hasOwn(type,'intrinsicName'),intrinsicValue:intrinsic===undefined?{undefined:true}:intrinsic,objectCandidate:objectFlagsCandidate,objectRawEqual:original===objectFlagsCandidate,maskEqual:[16,4,1024,256,8].map(mask=>(original&mask)===(objectFlagsCandidate&mask)),nonObjectFlagsValue:!(type.flags&403963917)&&original!==undefined&&original!==0,intrinsicCandidate:intrinsicCandidate===undefined?{undefined:true}:intrinsicCandidate,intrinsicEqual:intrinsic===intrinsicCandidate,inReadEqual:intrinsic===('intrinsicName' in type?type.intrinsicName:undefined)})+'\n');
}
function intrinsic(type){
 const value=type.intrinsicName,candidate=type.flags&402431?type.intrinsicName:undefined;
 fs.appendFileSync(path.join(process.env.DUMP_COUNTER_DIR,process.pid+'.intrinsic.jsonl'),JSON.stringify({id:type.id,flags:type.flags,own:Object.hasOwn(type,'intrinsicName'),value:value===undefined?{undefined:true}:value,candidate:candidate===undefined?{undefined:true}:candidate,equal:value===candidate,inEqual:value===('intrinsicName' in type?type.intrinsicName:undefined)})+'\n');
 return value;
}
globalThis.__dumpIntrinsic=intrinsic;
globalThis.__dumpRecord=record;
globalThis.__dumpStart=start;
function inject(context){if(context&&typeof context==='object'){Object.defineProperty(context,'__dumpRecord',{value:record,configurable:true});Object.defineProperty(context,'__dumpStart',{value:start,configurable:true});Object.defineProperty(context,'__dumpIntrinsic',{value:intrinsic,configurable:true});}return context;}
const create=vm.createContext;vm.createContext=function(context={},...args){return Reflect.apply(create,this,[inject(context),...args]);};
const run=vm.runInNewContext;vm.runInNewContext=function(code,context={},...args){return Reflect.apply(run,this,[code,inject(context),...args]);};
const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(context={},...args){return Reflect.apply(script,this,[inject(context),...args]);};
process.on('exit',()=>fs.writeFileSync(path.join(process.env.DUMP_COUNTER_DIR,process.pid+'.json'),JSON.stringify({pid:process.pid,argv:process.argv,visits,calls})+'\n'));
