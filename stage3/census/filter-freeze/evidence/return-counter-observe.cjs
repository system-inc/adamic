'use strict';
const fs=require('node:fs'),path=require('node:path'),inspector=require('node:inspector');
const directory=process.env.FILTER_FREEZE_OUT;
if (!directory) throw Error('FILTER_FREEZE_OUT required');
fs.mkdirSync(directory,{recursive:true});
const sites=JSON.parse(fs.readFileSync(path.join(__dirname,'evidence/sites.json'))).map(r=>r.site);
const counts=Object.fromEntries(sites.map(s=>[s,{calls:0,identity:0,frozen_arrays:0,undefined_identity:0}]));
const provenance=new WeakMap();let currentCase=null;let freezeEvents=0,exceptionEvents=0,freezeThrows=0;
Error.stackTraceLimit=100;
const log=(name,row)=>fs.appendFileSync(path.join(directory,`${process.pid}.${name}.jsonl`),JSON.stringify(row)+'\n');
const context=()=>({pid:process.pid,argv:process.argv,cwd:process.cwd(),case:currentCase});
globalThis.__adamicFreezeLookup=value=>provenance.get(value);
globalThis.__adamicFilterFreeze=(site,result,argument)=>{
 const c=counts[site]||(counts[site]={calls:0,identity:0,frozen_arrays:0,undefined_identity:0});c.calls++;
 if(result===argument){c.identity++;
  if(Array.isArray(result)){
   c.frozen_arrays++;freezeEvents++;
   let sites=provenance.get(result);if(!sites)provenance.set(result,sites=[]);
   if(!sites.includes(site))sites.push(site);
   Object.freeze(result);
  }else if(result===undefined)c.undefined_identity++;
 }
 return result;
};
// Pausing on *all* exceptions captures writes even when production catches the exception.
// The original exception is always resumed: no handler is installed to swallow errors.
const session=new inspector.Session();session.connect();
const post=(method,params={})=>{let result,error;session.post(method,params,(e,r)=>{error=e;result=r});if(error)throw error;return result;};
post('Debugger.enable');post('Debugger.setPauseOnExceptions',{state:'all'});
session.on('Debugger.paused',({params:p})=>{
 try{
  if(p.reason!=='exception'&&p.reason!=='promiseRejection')return;
  exceptionEvents++;
  const description=p.data?.description||'';
  if(!/TypeError/.test(description)||!/read only|not extensible|Cannot (?:add|delete|assign)|frozen|Cannot redefine property/.test(description))return;
  freezeThrows++;
  const frames=p.callFrames.map(f=>({function:f.functionName,url:f.url,scriptId:f.location.scriptId,line:f.location.lineNumber+1,column:f.location.columnNumber+1}));
  const candidates=[];
  // Identify exact frozen arrays in local scopes using their object identity in our WeakMap.
  for(const f of p.callFrames.slice(0,12))for(const scope of f.scopeChain.filter(s=>s.type==='local'||s.type==='block'||s.type==='catch')){
   const props=post('Runtime.getProperties',{objectId:scope.object.objectId,ownProperties:true,generatePreview:false}).result;
   for(const prop of props){const value=prop.value;if(!value?.objectId||value.subtype!=='array')continue;
    const found=post('Runtime.callFunctionOn',{objectId:value.objectId,functionDeclaration:'function(){return globalThis.__adamicFreezeLookup(this)}',returnByValue:true,silent:true});
    if(found.result.value)candidates.push({variable:prop.name,sites:found.result.value,frame:frames[p.callFrames.indexOf(f)]});
   }
  }
  log('throws',{...context(),description,frames,frozen_aliases:candidates,observed_sites:Object.entries(counts).filter(([,v])=>v.frozen_arrays).map(([s])=>s)});
 }catch(e){log('observer-errors',{...context(),stack:e.stack});}
 finally{post('Debugger.resume');}
});
// Keep source URLs: V8 paused call frames may leave url empty.
const scripts=new Map();session.on('Debugger.scriptParsed',({params:p})=>{scripts.set(p.scriptId,p.url)});
function flush(code,signal=null){
 fs.writeFileSync(path.join(directory,`${process.pid}.counts.json`),JSON.stringify({...context(),exit:code,signal,counts,freezeEvents,exceptionEvents,freezeThrows,scripts:Object.fromEntries(scripts)},null,2)+'\n');
}
process.on('exit',code=>flush(code));
for(const signal of ['SIGTERM','SIGINT'])process.once(signal,()=>{flush(null,signal);process.kill(process.pid,signal)});
// Upstream workers use Mocha's Runnable.run; observe test and hook titles without changing results.
const {createRequire}=require('node:module');
try{
 const Mocha=createRequire(path.join(process.cwd(),'package.json'))('mocha');
 const original=Mocha.Runnable.prototype.run;
 Mocha.Runnable.prototype.run=function(...args){currentCase={title:this.fullTitle(),file:this.file,type:this.type};return Reflect.apply(original,this,args)};
}catch(e){if(e.code!=='MODULE_NOT_FOUND')throw e;}
// Harness.evalFile can load the instrumented compiler into a separate VM global.
// Bridge only contexts executing this instrumentation; don't alter ordinary fixture globals.
const vm=require('node:vm');
for(const name of ['runInNewContext','runInContext']){
 const original=vm[name];vm[name]=function(code,sandbox,...rest){
  if(typeof code==='string'&&code.includes('__adamicFilterFreeze')){
   sandbox=sandbox||{};
   Object.defineProperty(sandbox,'__adamicFilterFreeze',{value:globalThis.__adamicFilterFreeze,configurable:true});
   Object.defineProperty(sandbox,'__adamicFreezeLookup',{value:globalThis.__adamicFreezeLookup,configurable:true});
  }
  return Reflect.apply(original,this,[code,sandbox,...rest]);
 };
}
