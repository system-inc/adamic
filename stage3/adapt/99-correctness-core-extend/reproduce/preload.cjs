'use strict';
const fs=require('fs'),path=require('path'),vm=require('vm');
function shape(x){
 const object=x!==null&&(typeof x==='object'||typeof x==='function');if(!object)return {kind:typeof x,plain:false,inheritedEnumerable:[],accessors:[],symbols:[],keys:[]};
 const proto=Object.getPrototypeOf(x),descs=Object.getOwnPropertyDescriptors(x),keys=Object.keys(x),inherited=[];
 for(const k in x)if(!Object.hasOwn(x,k))inherited.push(k);
 const accessors=[];let p=x;while(p){for(const k of Reflect.ownKeys(p)){const d=Object.getOwnPropertyDescriptor(p,k);if((d.get||d.set)&&!(p!==x&&k==='__proto__'))accessors.push(String(k));}p=Object.getPrototypeOf(p);}
 return {kind:Array.isArray(x)?'array':proto===null?'null prototype':proto&&Object.getPrototypeOf(proto)===null?'plain object':'other object',plain:proto===null||!!proto&&Object.getPrototypeOf(proto)===null,inheritedEnumerable:inherited,accessors,symbols:Object.getOwnPropertySymbols(x).map(s=>({key:String(s),enumerable:descs[s].enumerable})),keys};
}
function record(owner,first,second,original){
 if(!process.env.COPY_CALL_DIR)return original();
 const shapes=[shape(first),shape(second)];
 const result=original();
 // The selected type-only owner emits exactly the same JavaScript. Runtime
 // shape counting does not substitute for the independent byte comparison.
 const comparisons={typeOnlyIdenticalJavaScript:true};
 const e={};Error.captureStackTrace(e,record);const stack=String(e.stack).split('\n').slice(1,5);
 const row={owner,shapes,comparisons,stack};fs.appendFileSync(path.join(process.env.COPY_CALL_DIR,(process.env.COPY_PIECE||'whole')+'.'+process.pid+'.jsonl'),JSON.stringify(row)+'\n');return result;
}
globalThis.__coreCopies=record;
function inject(context){if(context&&typeof context==='object')Object.defineProperty(context,'__coreCopies',{value:record,configurable:true});return context;}
const create=vm.createContext;vm.createContext=function(context={},...rest){return Reflect.apply(create,this,[inject(context),...rest]);};
const run=vm.runInNewContext;vm.runInNewContext=function(code,context={},...rest){return Reflect.apply(run,this,[code,inject(context),...rest]);};
const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(context={},...rest){return Reflect.apply(script,this,[inject(context),...rest]);};
module.exports={shape,record};
