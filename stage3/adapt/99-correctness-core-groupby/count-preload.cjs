'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),crypto=require('node:crypto');
const root=process.env.CORE_COUNT_TREE,output=process.env.CORE_COUNT_DIR;
if(root&&process.cwd()===root){
 const file=root+'/src/compiler/core.ts';let text=fs.readFileSync(file,'utf8');
 if(!text.includes('__coreWaveCount')){
  const beforeSha256=crypto.createHash('sha256').update(text).digest('hex');if(beforeSha256!=='395b00fe5d458447ea495371593a1d646879a72a7882b59a92cb48dfede8d0f8')throw Error('count source drift');
  const ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),s=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);const edits=[];
  for(const name of ['groupBy','clone']){
   const ds=s.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text===name&&n.body);if(ds.length!==1)throw Error('count owner '+name);
   const returns=ds[0].body.statements.filter(ts.isReturnStatement);if(returns.length!==1)throw Error('count return '+name);
   edits.push({at:ds[0].body.getStart(s)+1,text:` (globalThis as any).__coreWaveCount("${name}", ${name==='groupBy'?'values, undefined, keySelector':'object'}); `});
   const at=returns[0].getStart(s);edits.push({at,text:`(globalThis as any).__coreWaveCount("${name}", ${name==='groupBy'?'values':'object'}, result); `});
  }
  for(const e of edits.sort((a,b)=>b.at-a.at))text=text.slice(0,e.at)+e.text+text.slice(e.at);
  fs.writeFileSync(file,text);
  fs.writeFileSync(output+'/instrumentation.json',JSON.stringify({measurementOnly:true,beforeSha256,afterSha256:crypto.createHash('sha256').update(text).digest('hex'),tree:root,owners:['groupBy','clone'],lineCountPreserved:true,insertions:edits},null,2)+'\n');
 }
}
const counts={groupBy:0,clone:0,completed:{groupBy:0,clone:0},groupByClasses:{undefined:0,empty:0,dense:0,sparse:0},groupByKeys:Object.create(null),groupBySelectors:Object.create(null),groupByCallers:Object.create(null),cloneClasses:{nullish:0,object:0,primitive:0}};
function record(name,input,result,selector){
 if(arguments.length===3){counts.completed[name]++;if(name==='groupBy'){const keys=Object.keys(result).sort().join(',');counts.groupByKeys[keys]=(counts.groupByKeys[keys]||0)+1;}return;}
 counts[name]++;
 if(name==='groupBy'){
  const selectorName=selector?.name||'<anonymous>';counts.groupBySelectors[selectorName]=(counts.groupBySelectors[selectorName]||0)+1;
  const error={};Error.captureStackTrace(error,record);const caller=String(error.stack).split('\n').slice(1).map(s=>s.trim()).find(s=>!/\bat (?:Object\.)?groupBy\b/.test(s)&&!s.includes('count-preload'))||'<unknown>';counts.groupByCallers[caller]=(counts.groupByCallers[caller]||0)+1;
  const cls=input===undefined?'undefined':input.length===0?'empty':Object.keys(input).filter(k=>/^(0|[1-9][0-9]*)$/.test(k)&&Number(k)<input.length).length===input.length?'dense':'sparse';counts.groupByClasses[cls]++;
 }else counts.cloneClasses[input==null?'nullish':typeof input==='object'||typeof input==='function'?'object':'primitive']++;
}
globalThis.__coreWaveCount=record;
function inject(context){if(context&&typeof context==='object')Object.defineProperty(context,'__coreWaveCount',{value:record,configurable:true});return context;}
const create=vm.createContext;vm.createContext=function(context={},...rest){return Reflect.apply(create,this,[inject(context),...rest]);};
const run=vm.runInNewContext;vm.runInNewContext=function(code,context={},...rest){return Reflect.apply(run,this,[code,inject(context),...rest]);};
const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(code,...rest){return Reflect.apply(script,this,[inject(code||{}),...rest]);};
process.on('exit',()=>{if(output&&(counts.groupBy||counts.clone))fs.appendFileSync(path.join(output,'counts.jsonl'),JSON.stringify({pid:process.pid,...counts})+'\n');});
