'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
const maps=new Map();
function record(array,offset){
 const dir=process.env.EA_CALL_DIR;if(!dir)return;
 const length=array===undefined?null:array.length;
 const nonInteger=!Number.isInteger(offset);
 const classification=array===undefined?'array undefined':nonInteger?'non-integer':offset>=length||offset< -length?'out-of-range integer':offset<0?'in-range negative integer':'in-range non-negative integer';
 let caller=null,frame=null;
 if(classification!=='in-range non-negative integer'){
  const e={};Error.captureStackTrace(e,record);
  const frames=String(e.stack).split('\n').slice(1).map(s=>s.trim());
  for(const candidate of frames){
   if(candidate.includes('count-preload'))continue;
   const match=candidate.match(/(?:\(|at )((?:file:\/\/)?[^()]+):(\d+):(\d+)\)?$/);if(!match)continue;
   let source=match[1],line=Number(match[2]);
   const filename=source.replace(/^file:\/\//,'');
   const actual=fs.existsSync(filename+'.map')?filename+'.map':path.join('/workspace/scratch/elementat-b/counted/built/local',path.basename(filename)+'.map');
   if(fs.existsSync(actual)){
    const {TraceMap,originalPositionFor}=require('/workspace/scratch/elementat-b/modules/@jridgewell/trace-mapping');
    let map=maps.get(actual);if(!map){map=new TraceMap(JSON.parse(fs.readFileSync(actual,'utf8')));maps.set(actual,map);}
    const pos=originalPositionFor(map,{line,column:Number(match[3])-1});if(pos.source){source=pos.source;line=pos.line;}
   }
   source=source.replace(/^.*\/src\//,'src/');
   if(source==='src/compiler/core.ts'&&line>=1076&&line<=1090)continue;
   caller=source+':'+line;frame=candidate;break;
  }
 }
 const encoded=Number.isNaN(offset)?'NaN':offset===Infinity?'+Infinity':offset===-Infinity?'-Infinity':Object.is(offset,-0)?'-0':offset;
 fs.appendFileSync(path.join(dir,(process.env.EA_SHARD||'whole')+'.'+process.pid+'.jsonl'),JSON.stringify({recorderVersion:2,offset:encoded,length,classification,nonIntegerOffset:nonInteger,caller,frame,pid:process.pid})+'\n');
}
globalThis.__elementAtCensus=record;
function inject(context){if(context&&typeof context==='object')Object.defineProperty(context,'__elementAtCensus',{value:record,configurable:true});return context;}
const create=vm.createContext;vm.createContext=function(context={},...rest){return Reflect.apply(create,this,[inject(context),...rest]);};
const run=vm.runInNewContext;vm.runInNewContext=function(code,context={},...rest){return Reflect.apply(run,this,[code,inject(context),...rest]);};
const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(context={},...rest){return Reflect.apply(script,this,[inject(context),...rest]);};
