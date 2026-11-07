'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
function encode(v){return v===undefined?{type:'undefined'}:typeof v==='number'?(Number.isNaN(v)?{type:'number',value:'NaN'}:Object.is(v,-0)?{type:'number',value:'-0'}:!Number.isFinite(v)?{type:'number',value:String(v)}:{type:'number',value:v}):{type:typeof v,value:v};}
const callDirectory=process.env.COMPARE_CALL_DIR;
const recordPrefix=callDirectory?path.join(callDirectory,(process.env.COMPARE_SHARD||'whole')+'.'+process.pid):null;
let descriptor=null;const byte=Buffer.alloc(1);
function record(a,b,original,adapted){
 if(!recordPrefix)return;
 const ta=typeof a,tb=typeof b,ai=ta==='undefined'?0:ta==='number'?1:ta==='string'?2:-1,bi=tb==='undefined'?0:tb==='number'?1:tb==='string'?2:-1;
 const valid=ai>=0&&bi>=0&&Number.isInteger(original)&&original>=-1&&original<=1&&Number.isInteger(adapted)&&adapted>=-1&&adapted<=1;
 if(descriptor===null)descriptor=fs.openSync(recordPrefix+'.bin','a');
 byte[0]=valid?9*(3*ai+bi)+3*(original+1)+(adapted+1):255;
 if(fs.writeSync(descriptor,byte,0,1)!==1)throw Error('compare census short write');
 const mixed=(ta==='string'&&tb==='number')||(ta==='number'&&tb==='string');
 if(!valid||mixed||!Object.is(original,adapted)){
  fs.appendFileSync(recordPrefix+'.failure.txt',JSON.stringify({recorderVersion:4,a:encode(a),b:encode(b),typeofA:ta,typeofB:tb,original,adapted,mixed,same:Object.is(original,adapted)})+'\n');
  throw Error('compare census STOP: unexpected type, mixed operands, or result difference');
 }
}
globalThis.__compareCensus=record;
function inject(context){if(context&&typeof context==='object')Object.defineProperty(context,'__compareCensus',{value:record,configurable:true});return context;}
const create=vm.createContext;vm.createContext=function(context={},...rest){return Reflect.apply(create,this,[inject(context),...rest]);};
const run=vm.runInNewContext;vm.runInNewContext=function(code,context={},...rest){return Reflect.apply(run,this,[code,inject(context),...rest]);};
const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(context={},...rest){return Reflect.apply(script,this,[inject(context),...rest]);};
