'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
const out=process.env.SCOUT_NAMESPACE_COUNTS;
if(out){
 const key='patchWriteFileEnsuringDirectory',counts={pid:process.pid,thread:require('node:worker_threads').threadId,argv:process.argv,installs:0,reads:0,controls:0,values:0,sites:[]};let controlDepth=0;
 const wrapped=new WeakSet();
 function wrapDefine(original){if(wrapped.has(original))return original;function defineProperty(object,name,descriptor){
  if(name!==key||typeof descriptor?.get!=='function')return Reflect.apply(original,this,arguments);
  const get=descriptor.get;counts.installs++;if(counts.sites.length<12)counts.sites.push(String(new Error().stack).split('\n').slice(2,6));
  function observed(){if(controlDepth)counts.controls++;else counts.reads++;return Reflect.apply(get,this,arguments);}
  const result=Reflect.apply(original,this,[object,name,{...descriptor,get:observed}]);
  controlDepth++;try{const via=Reflect.get(object,name),direct=Reflect.apply(get,object,[]);if(via!==direct)throw Error('namespace control changed identity');if(typeof via==='function')counts.values++;}finally{controlDepth--;}
  return result;
 }wrapped.add(defineProperty);return defineProperty;}
 Object.defineProperty=wrapDefine(Object.defineProperty);
 const create=vm.createContext,run=vm.runInContext,scriptRun=vm.Script.prototype.runInContext;
 function install(context){const native=run('Object.defineProperty',context);if(!wrapped.has(native)){const object=run('Object',context);object.defineProperty=wrapDefine(native);}return context;}
 vm.createContext=function(sandbox,options){return install(Reflect.apply(create,this,arguments));};
 vm.runInNewContext=function(code,sandbox,options){const context=install(Reflect.apply(create,vm,[sandbox,options]));return Reflect.apply(run,vm,[code,context,options]);};
 vm.Script.prototype.runInNewContext=function(sandbox,options){const context=install(Reflect.apply(create,vm,[sandbox,options]));return Reflect.apply(scriptRun,this,[context,options]);};
 process.on('exit',()=>{fs.mkdirSync(out,{recursive:true});fs.writeFileSync(path.join(out,`${process.pid}-${counts.thread}.json`),JSON.stringify(counts,null,2)+'\n');});
}
