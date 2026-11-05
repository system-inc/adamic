'use strict';
const fs = require('node:fs'),path = require('node:path'),{createRequire}=require('node:module');
let active=null;
const counts=new Map();
globalThis.__debugVisibleCensus=(id,ok)=>{
 const title=globalThis.__visibleControlTitle ? "[planted at startup] " + globalThis.__visibleControlTitle : active; const key=JSON.stringify([id,title]);const row=counts.get(key)||{id,test:title,runs:0,wouldFail:0};
 row.runs++; if(!ok)row.wouldFail++;counts.set(key,row);return ok;
};
process.on('exit',()=>{if(process.env.VISIBLE_COUNTS)fs.appendFileSync(process.env.VISIBLE_COUNTS,[...counts.values()].map(r=>JSON.stringify({...r,pid:process.pid})).join('\n')+'\n')});
if(['run.js','_mocha','mocha.js'].includes(path.basename(process.argv[1]||''))){
 const Mocha=createRequire(path.join(process.cwd(),'package.json'))('mocha');
 const run=Mocha.Runnable.prototype.run;
 Mocha.Runnable.prototype.run=function(...args){active=this.titlePath().filter(Boolean).join(' ');return Reflect.apply(run,this,args)};
 const emit=Mocha.Runner.prototype.emit;
 Mocha.Runner.prototype.emit=function(...args){const[event,runnable]=args;
 if(process.env.VISIBLE_IDENTITIES&&['pass','fail','pending'].includes(event))fs.appendFileSync(process.env.VISIBLE_IDENTITIES,JSON.stringify({event,title:runnable.titlePath().filter(Boolean).join(' '),pid:process.pid})+'\n');
 if(event==='test end'||event==='hook end')active=null;return Reflect.apply(emit,this,args)};
}
const vm=require('node:vm'),run=vm.runInNewContext;
vm.runInNewContext=function(code,context,...rest){if(context&&typeof code==='string'&&code.includes('__debugVisibleCensus'))Object.defineProperty(context,'__debugVisibleCensus',{value:globalThis.__debugVisibleCensus,configurable:true});return Reflect.apply(run,this,[code,context,...rest])};
if(process.env.VISIBLE_CONTROLS){
 const ts=require('typescript');
 const manifest=JSON.parse(fs.readFileSync('/workspace/scratch/99t-counter-manifest.json','utf8'));
 active='[planted failing input at shim startup]';
 for(const {id,condition}of manifest){
  const text=ts.transpileModule(`function predicate(node: any, host: any, sourceFile: any, f: any, value: any, values: any, s: any) { return (${condition}); }`,{compilerOptions:{target:ts.ScriptTarget.ES2024}}).outputText;
  const predicate=new Function('SyntaxKind','FlowFlags',text+';return predicate;')(ts.SyntaxKind,ts.FlowFlags);
  const result=predicate.call({source:{},target:{},mapper1:{},mapper2:{}},{kind:ts.SyntaxKind.Identifier},{},{kind:ts.SyntaxKind.Identifier},{flags:ts.FlowFlags.Label,antecedent:1},{},null,{});
  if(result!==false)throw Error('control did not fail '+id);
  globalThis.__debugVisibleCensus(id,result);
 }
 active=null;
}

if(process.env.VISIBLE_ACTUAL_CONTROLS) require("./startup-controls.cjs");
