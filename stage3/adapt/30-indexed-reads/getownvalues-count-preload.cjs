'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
let sequence=0;
const own=Function.call.bind(Object.prototype.hasOwnProperty);
const index=k=>typeof k==='string'&&String(Number(k))===k&&Number.isInteger(Number(k))&&Number(k)>=0&&Number(k)<4294967295;
function record(value){
 const dir=process.env.GOV_CALL_DIR;if(!dir)return;
 let row={event:'call',pid:process.pid,sequence:++sequence,shard:process.env.GOV_SHARD||'seed',kind:'other'};
 if(Array.isArray(value)){
  const keys=Reflect.ownKeys(value),length=Object.getOwnPropertyDescriptor(value,'length').value;
  const indices=keys.filter(index);let presentUndefined=0,accessorIndices=0;
  for(const key of indices){const d=Object.getOwnPropertyDescriptor(value,key);if(own(d,'value'))presentUndefined+=d.value===undefined;else accessorIndices++;}
  const nonIndexProperties=keys.filter(k=>k!=='length'&&!index(k)).map(k=>typeof k==='symbol'?'symbol:'+String(k):k);
  row={...row,kind:'array',length,present:indices.length,holes:length-indices.length,sparse:indices.length<length,presentUndefined,accessorIndices,nonIndexProperties};
 }else if(value!==null&&typeof value==='object'){
  const proto=Object.getPrototypeOf(value),constructor=proto&&Object.getOwnPropertyDescriptor(proto,'constructor');
  const plain=proto===null||(Object.getPrototypeOf(proto)===null&&constructor&&typeof constructor.value==='function'&&Function.prototype.toString.call(constructor.value)==='function Object() { [native code] }');
  row={...row,kind:plain?'plain-record':'other',prototype:proto===null?'null':plain?'Object':constructor?.value?.name||'custom'};
 }else row={...row,valueType:value===null?'null':typeof value};
 fs.appendFileSync(path.join(dir,(process.env.GOV_SHARD||'seed')+'.'+process.pid+'.jsonl'),JSON.stringify(row)+'\n');
}
globalThis.__getOwnValuesCensus=record;
function inject(context){if(context&&typeof context==='object')Object.defineProperty(context,'__getOwnValuesCensus',{value:record,configurable:true});return context;}
const create=vm.createContext;vm.createContext=function(context={},...rest){return Reflect.apply(create,this,[inject(context),...rest]);};
const run=vm.runInNewContext;vm.runInNewContext=function(code,context={},...rest){return Reflect.apply(run,this,[code,inject(context),...rest]);};
const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(context={},...rest){return Reflect.apply(script,this,[inject(context),...rest]);};
