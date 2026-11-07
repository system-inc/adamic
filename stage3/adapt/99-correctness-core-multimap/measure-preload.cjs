'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),assert=require('node:assert/strict');
const maps=new WeakSet(),counts={},failures=[],is=Object.is;
const original=require(process.env.MM_BODIES);
const {legacyAdd,legacyRemove,freeAdd,freeRemove}=vm.runInThisContext('(function(){'+original.dependencies+';return {legacyAdd:('+original.add+'),legacyRemove:('+original.remove+'),freeAdd:('+original.freeAdd+'),freeRemove:('+original.freeRemove+')};})()');
function clone(map){const arrays=new Map();return new Map(Array.from(map,([k,v])=>{if(!arrays.has(v))arrays.set(v,v.slice());return [k,arrays.get(v)];}));}
function equalArray(a,b){return a===undefined?b===undefined:Array.isArray(b)&&a.length===b.length&&Array.from({length:a.length},(_,i)=>i).every(i=>(i in a)===(i in b)&&is(a[i],b[i]));}
function equalMap(a,b){if(a.size!==b.size)return false;const aa=[...a],bb=[...b];return aa.every(([k,v],i)=>is(k,bb[i][0])&&equalArray(v,bb[i][1]));}
function check(op,map,key,value){
 const existing=map.get(key);if(existing!==undefined&&!Array.isArray(existing))throw Error('Unexpected non-array MultiMap value');const kind=op+':'+(existing===undefined?'absent':existing.includes(value)?'present-value':'absent-value')+':'+(existing?.length===0?'empty':existing===undefined?'missing':existing.length===1?'singleton':'multiple');
 counts[kind]=(counts[kind]||0)+1;
 const a=clone(map),b=clone(map);let ar,br,ae,be;
 try{ar=Reflect.apply(op==='add'?legacyAdd:legacyRemove,a,[key,value]);}catch(e){ae=e;}
 try{br=(op==='add'?freeAdd:freeRemove)(b,key,value);}catch(e){be=e;}
 const equal=!ae&&!be&&equalArray(ar,br)&&equalMap(a,b);
 if(!equal){failures.push({op,kind,mapSize:map.size,legacyError:ae?.name,freeError:be?.name});throw Error('MultiMap measurement counterexample');}
 counts.calls=(counts.calls||0)+1;counts.sparseArrays=(counts.sparseArrays||0)+(existing?Object.keys(existing).filter(k=>/^\d+$/.test(k)).length!==existing.length:0);
}
function register(map){maps.add(map);counts.created=(counts.created||0)+1;return map;}
function enumeration(value){if(value&&maps.has(value)){counts.enumerations=(counts.enumerations||0)+1;throw Error('STOP: MultiMap own-key enumeration');}return value;}
function ownCall(receiver,fn,name){return function(...args){for(const arg of name==='assign'?args.slice(1):args.slice(0,1))enumeration(arg);return Reflect.apply(fn,receiver,args);};}
const callbacks={__mmCheck:check,__mmRegister:register,__mmEnumeration:enumeration,__mmOwnCall:ownCall};Object.assign(globalThis,callbacks);
function inject(context){if(context&&typeof context==='object')for(const [key,value]of Object.entries(callbacks))Object.defineProperty(context,key,{value,configurable:true});return context;}
const create=vm.createContext;vm.createContext=function(context={},...rest){return Reflect.apply(create,this,[inject(context),...rest]);};
const run=vm.runInNewContext;vm.runInNewContext=function(code,context={},...rest){return Reflect.apply(run,this,[code,inject(context),...rest]);};
const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(context={},...rest){return Reflect.apply(script,this,[inject(context),...rest]);};
process.on('exit',()=>{if(!process.env.MM_COUNTS)return;fs.mkdirSync(process.env.MM_COUNTS,{recursive:true});fs.writeFileSync(path.join(process.env.MM_COUNTS,(process.env.MM_SHARD||'prepare')+'.'+process.pid+'.json'),JSON.stringify({counts,failures}));});
