'use strict';
const fs=require('fs'),path=require('path'),vm=require('vm');
function record(s1,s2,max,a,b){const dir=process.env.LV_CALL_DIR;if(!dir)return;fs.appendFileSync(path.join(dir,(process.env.LV_UNIT||'whole')+'.'+process.pid+'.jsonl'),JSON.stringify({s1,s2,max,original:a===undefined?'undefined':a,adapted:b===undefined?'undefined':b,equal:Object.is(a,b),equalStrings:s1===s2,emptyS1:s1.length===0,emptyS2:s2.length===0,caseOnly:s1!==s2&&s1.toLowerCase()===s2.toLowerCase(),giveUp:a===undefined,maxBelowOne:max<1})+'\n');}
globalThis.__levenshteinRecord=record;
function inject(context){if(context&&typeof context==='object')Object.defineProperty(context,'__levenshteinRecord',{value:record,configurable:true});return context;}
const create=vm.createContext;vm.createContext=function(context={},...rest){return Reflect.apply(create,this,[inject(context),...rest]);};
const run=vm.runInNewContext;vm.runInNewContext=function(code,context={},...rest){return Reflect.apply(run,this,[code,inject(context),...rest]);};
const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(context={},...rest){return Reflect.apply(script,this,[inject(context),...rest]);};
