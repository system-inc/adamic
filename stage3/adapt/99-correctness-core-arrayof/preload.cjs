'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
function write(row){const dir=process.env.ARRAYOF_CALL_DIR;if(dir)fs.appendFileSync(path.join(dir,(process.env.ARRAYOF_UNIT||'whole')+'.'+process.pid+'.jsonl'),JSON.stringify(row)+'\n');}
function begin(count){const classification=Number.isNaN(count)?'NaN':!Number.isFinite(count)?'infinite':!Number.isInteger(count)?'fractional':count<0?'negative':'non-negative integer';write({event:'call',count:Number.isFinite(count)?count:String(count),classification});}
function end(count,result,shadow){let equal=result.length===shadow.length;for(let i=0;i<result.length;i++)equal=equal&&Object.hasOwn(result,i)===Object.hasOwn(shadow,i)&&Object.is(result[i],shadow[i]);write({event:'result',count,length:result.length,equal});if(!equal)throw Error('arrayOf counterexample');}
function thrown(count,index,error){write({event:'result',count,callbackThrewAt:index,equal:true,sameErrorRethrown:true});}
function inject(context){if(context&&typeof context==='object'){Object.defineProperty(context,'__arrayOfBegin',{value:begin,configurable:true});Object.defineProperty(context,'__arrayOfEnd',{value:end,configurable:true});Object.defineProperty(context,'__arrayOfThrow',{value:thrown,configurable:true});}return context;}
inject(globalThis);
const create=vm.createContext;vm.createContext=function(context={},...rest){return Reflect.apply(create,this,[inject(context),...rest]);};
const run=vm.runInNewContext;vm.runInNewContext=function(code,context={},...rest){return Reflect.apply(run,this,[code,inject(context),...rest]);};
const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(context={},...rest){return Reflect.apply(script,this,[inject(context),...rest]);};
