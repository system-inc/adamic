'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
function observe(array,offset,original,candidate){
 const dir=process.env.EA_CALL_DIR;if(!dir)return original;
 const length=array===undefined?null:array.length;
 const equal=Object.is(original,candidate);
 fs.appendFileSync(path.join(dir,(process.env.EA_SHARD||'preparation')+'.'+process.pid+'.jsonl'),JSON.stringify({offset:Number.isFinite(offset)?(Object.is(offset,-0)?'-0':offset):String(offset),length,equal,kind:original===null?'null':typeof original})+'\n');
 if(!equal)throw Error('elementAt declaration result difference');
 return original;
}
function read(get,site,kind,tsObject){
 const dir=process.env.EA_CALL_DIR;
 try { const value=get(); if(dir)fs.appendFileSync(path.join(dir,(process.env.EA_SHARD||'preparation')+'.'+process.pid+'.reads.jsonl'),JSON.stringify({site,kind,tsObject,tdz:false,initialized:typeof value==='function'})+'\n'); return value; }
 catch(error){ if(dir)fs.appendFileSync(path.join(dir,(process.env.EA_SHARD||'preparation')+'.'+process.pid+'.reads.jsonl'),JSON.stringify({site,kind,tsObject,tdz:error?.name==='ReferenceError',initialized:false,error:String(error)})+'\n'); throw error; }
}
function exportRead(get,site){
 const dir=process.env.EA_CALL_DIR;
 try { const value=get(); if(dir){const e={};Error.captureStackTrace(e,exportRead);fs.appendFileSync(path.join(dir,(process.env.EA_SHARD||'preparation')+'.'+process.pid+'.exports.jsonl'),JSON.stringify({site,tdz:false,initialized:typeof value==='function',stack:String(e.stack)})+'\n');} return value; }
 catch(error){if(dir)fs.appendFileSync(path.join(dir,(process.env.EA_SHARD||'preparation')+'.'+process.pid+'.exports.jsonl'),JSON.stringify({site,tdz:error?.name==='ReferenceError',initialized:false,error:String(error)})+'\n');throw error;}
}
globalThis.__eaObserve=observe; globalThis.__eaRead=read; globalThis.__eaExportRead=exportRead;
function inject(context){if(context&&typeof context==='object'){Object.defineProperty(context,'__eaObserve',{value:observe,configurable:true}); Object.defineProperty(context,'__eaRead',{value:read,configurable:true}); Object.defineProperty(context,'__eaExportRead',{value:exportRead,configurable:true});}return context;}
const create=vm.createContext;vm.createContext=function(context={},...rest){return Reflect.apply(create,this,[inject(context),...rest]);};
const run=vm.runInNewContext;vm.runInNewContext=function(code,context={},...rest){return Reflect.apply(run,this,[code,inject(context),...rest]);};
const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(context={},...rest){return Reflect.apply(script,this,[inject(context),...rest]);};
