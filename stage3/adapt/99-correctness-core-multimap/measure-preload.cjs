'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),assert=require('node:assert/strict');
const maps=new WeakSet(),counts={},failures=[],is=Object.is;
const original=require(process.env.MM_BODIES);
const {legacyAdd,legacyRemove,freeAdd,freeRemove}=vm.runInThisContext('(function(){'+original.dependencies+';return {legacyAdd:('+original.add+'),legacyRemove:('+original.remove+'),freeAdd:('+original.freeAdd+'),freeRemove:('+original.freeRemove+')};})()');
function clone(map){const arrays=new Map();return new Map(Array.from(map,([k,v])=>{if(!arrays.has(v))arrays.set(v,v.slice());return [k,arrays.get(v)];}));}
function equalArray(a,b){return a===undefined?b===undefined:Array.isArray(b)&&a.length===b.length&&Array.from({length:a.length},(_,i)=>i).every(i=>(i in a)===(i in b)&&is(a[i],b[i]));}
function equalMap(a,b){if(a.size!==b.size)return false;const aa=[...a],bb=[...b];return aa.every(([k,v],i)=>is(k,bb[i][0])&&equalArray(v,bb[i][1]));}
const sameKey=(a,b)=>a===b||(a!==a&&b!==b);
class Snapshot extends Map {
 constructor(base,key){super();this.base=base;this.key=key;this.beforeHas=base.has(key);this.before=base.get(key);this.originalBucket=Array.isArray(this.before)?this.before.slice():this.before;this.bucket=this.originalBucket;this.present=this.beforeHas;this.operations=[];}
 keyOnly(key){assert(sameKey(key,this.key),'helper touched another key');}
 get(key){this.keyOnly(key);this.operations.push('get');return this.present?this.bucket:undefined;}
 set(key,value){this.keyOnly(key);this.operations.push('set');this.present=true;this.bucket=value;return this;}
 delete(key){this.keyOnly(key);this.operations.push('delete');const had=this.present;this.present=false;return had;}
 clear(){throw Error('helper unexpectedly cleared its Map');}
 get size(){return this.base.size+(this.present?1:0)-(this.beforeHas?1:0);}
 *entries(){for(const [key,value]of this.base){if(sameKey(key,this.key)){if(this.present)yield [key,this.bucket];}else yield [key,value===this.before?this.originalBucket:value];}if(!this.beforeHas&&this.present)yield [this.key===0?0:this.key,this.bucket];}
 [Symbol.iterator](){return this.entries();}
}
function equalSnapshots(a,b){return a.base===b.base&&sameKey(a.key,b.key)&&a.beforeHas===b.beforeHas&&a.present===b.present&&a.size===b.size&&JSON.stringify(a.operations)===JSON.stringify(b.operations)&&equalArray(a.originalBucket,b.originalBucket)&&equalArray(a.bucket,b.bucket);}
function evaluate(op,map,key,value,lazy){const a=lazy?new Snapshot(map,key):clone(map),b=lazy?new Snapshot(map,key):clone(map);let ar,br,ae,be;try{ar=Reflect.apply(op==='add'?legacyAdd:legacyRemove,a,[key,value]);}catch(e){ae=e;}try{br=(op==='add'?freeAdd:freeRemove)(b,key,value);}catch(e){be=e;}
 const equal=!ae&&!be&&equalArray(ar,br)&&(lazy?equalSnapshots(a,b):equalMap(a,b));if(lazy&&!ae&&!be){const valid=op==='add'?['get','get,set']:['get','get,delete'];assert(valid.includes(a.operations.join(','))&&valid.includes(b.operations.join(',')),'unexpected helper operation trace');}return {equal,a,b,ar,br,ae,be};}
function crossCheck(){let cases=0;const v={},w={},shared=[v,w,v],inputs=[new Map(),new Map([['k',[]]]),new Map([['k',[v]]]),new Map([['k',[v,w,v]]]),new Map([['k',shared],['alias',shared]]),new Map([['k',[,w]]]),new Map([[NaN,[NaN,v]]]),new Map([[0,[undefined]]]),new Map([[v,[w,v]]]),new Map(Array.from({length:1024},(_,i)=>[i,[i]]))];
 for(const map of inputs)for(const key of ['k',NaN,-0,v,500])for(const value of [v,w,undefined,NaN,500])for(const op of ['add','remove']){const lazy=evaluate(op,map,key,value,true),full=evaluate(op,map,key,value,false);assert(lazy.equal&&full.equal,'comparator witness mismatch');assert(equalMap(lazy.a,full.a)&&equalMap(lazy.b,full.b)&&equalArray(lazy.ar,full.ar)&&equalArray(lazy.br,full.br),'lazy snapshot differs from full native Map clone');cases++;}return cases;}
if(process.env.MM_VALIDATE){const cases=crossCheck();fs.writeFileSync(process.env.MM_VALIDATE,JSON.stringify({recorderVersion:2,crossChecks:cases,passed:true,coverage:['entry order','shared arrays','empty buckets','missing keys','duplicates','sparse arrays','undefined','NaN','object keys','1024-key maps']},null,2)+'\n');}
function check(op,map,key,value){
 const existing=map.get(key);if(existing!==undefined&&!Array.isArray(existing))throw Error('Unexpected non-array MultiMap value');const kind=op+':'+(existing===undefined?'absent':existing.includes(value)?'present-value':'absent-value')+':'+(existing?.length===0?'empty':existing===undefined?'missing':existing.length===1?'singleton':'multiple');
 counts[kind]=(counts[kind]||0)+1;
 const lazy=map.size>32,result=evaluate(op,map,key,value,lazy),{equal,ae,be}=result;
 counts[lazy?'lazySnapshots':'fullClones']=(counts[lazy?'lazySnapshots':'fullClones']||0)+1;
 if(!equal){failures.push({op,kind,mapSize:map.size,legacyError:ae?.name,freeError:be?.name});flush();throw Error('MultiMap measurement counterexample');}
 counts.calls=(counts.calls||0)+1;counts.sparseArrays=(counts.sparseArrays||0)+(existing?Object.keys(existing).filter(k=>/^\d+$/.test(k)).length!==existing.length:0);
}
function register(map){maps.add(map);counts.created=(counts.created||0)+1;return map;}
function enumeration(value){if(value&&maps.has(value)){counts.enumerations=(counts.enumerations||0)+1;flush();throw Error('STOP: MultiMap own-key enumeration');}return value;}
function ownCall(receiver,fn,name){return function(...args){for(const arg of name==='assign'?args.slice(1):args.slice(0,1))enumeration(arg);return Reflect.apply(fn,receiver,args);};}
const callbacks={__mmCheck:check,__mmRegister:register,__mmEnumeration:enumeration,__mmOwnCall:ownCall};Object.assign(globalThis,callbacks);
function inject(context){if(context&&typeof context==='object')for(const [key,value]of Object.entries(callbacks))Object.defineProperty(context,key,{value,configurable:true});return context;}
const create=vm.createContext;vm.createContext=function(context={},...rest){return Reflect.apply(create,this,[inject(context),...rest]);};
const run=vm.runInNewContext;vm.runInNewContext=function(code,context={},...rest){return Reflect.apply(run,this,[code,inject(context),...rest]);};
const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(context={},...rest){return Reflect.apply(script,this,[inject(context),...rest]);};
let exitEmissions=0,lastExitLateCalls=0;
function flush(){if(!process.env.MM_COUNTS)return;fs.mkdirSync(process.env.MM_COUNTS,{recursive:true});fs.writeFileSync(path.join(process.env.MM_COUNTS,(process.env.MM_SHARD||'prepare')+'.'+process.pid+'.json'),JSON.stringify({recorderVersion:3,counts,failures,exitEmissions,lastExitLateCalls}));}
const originalEmit=process.emit;process.emit=function(event,...args){if(event!=='exit')return Reflect.apply(originalEmit,this,[event,...args]);const before=counts.calls||0;exitEmissions++;try{return Reflect.apply(originalEmit,this,[event,...args]);}finally{lastExitLateCalls=(counts.calls||0)-before;flush();}};
