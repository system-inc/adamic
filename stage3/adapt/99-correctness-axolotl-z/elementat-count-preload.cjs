'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
function record(array,offset){
 const dir=process.env.EA_CALL_DIR;if(!dir)return;
 const length=array===undefined?null:array.length;
 const nonInteger=!Number.isInteger(offset);
 const classification=array===undefined?'array undefined':nonInteger?'non-integer':offset>=length||offset< -length?'out-of-range integer':offset<0?'in-range negative integer':'in-range non-negative integer';
 const e={};Error.captureStackTrace(e,record);
 const frames=String(e.stack).split('\n').slice(1).map(s=>s.trim());
 const frame=frames.find(s=>!s.includes('count-preload.cjs')&&!/core\.ts:10(?:7\d|8\d|90):/.test(s))||null;
 const match=frame?.match(/(?:\(|at )((?:file:\/\/)?[^()]+):(\d+):(\d+)\)?$/);
 const caller=match?match[1].replace(/^.*\/src\//,'src/')+':'+match[2]:frame;
 const encoded=Number.isNaN(offset)?'NaN':offset===Infinity?'+Infinity':offset===-Infinity?'-Infinity':Object.is(offset,-0)?'-0':offset;
 fs.appendFileSync(path.join(dir,(process.env.EA_SHARD||'whole')+'.'+process.pid+'.jsonl'),JSON.stringify({offset:encoded,length,classification,nonIntegerOffset:nonInteger,caller,frame,pid:process.pid})+'\n');
}
globalThis.__elementAtCensus=record;
function inject(context){if(context&&typeof context==='object')Object.defineProperty(context,'__elementAtCensus',{value:record,configurable:true});return context;}
const create=vm.createContext;vm.createContext=function(context={},...rest){return Reflect.apply(create,this,[inject(context),...rest]);};
const run=vm.runInNewContext;vm.runInNewContext=function(code,context={},...rest){return Reflect.apply(run,this,[code,inject(context),...rest]);};
const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(context={},...rest){return Reflect.apply(script,this,[inject(context),...rest]);};
