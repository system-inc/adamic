// Scalar counters retain no source or result arrays. Loaded before the bundled tsc entrypoint.
const fs=require('fs');
const input=JSON.parse(fs.readFileSync(process.env.ARRAY_COPIES_SITES,'utf8'));
const counters=Object.fromEntries(input.sites.filter(s=>s.depth>0&&s.classification!=='unresolved').map(s=>[s.id,{calls:0,throws:0,executions:0,elements:0}]));
function record(id,result,args,method){
 const c=counters[id];
 let fresh=Array.isArray(result);
 if(method.startsWith('helper:')){
  if(args.some(a=>a===result))fresh=false;
  if(['helper:flatMap','helper:toSorted','helper:sortAndDeduplicate'].includes(method)&&result?.length===0)fresh=false;
 }
 if(fresh){c.executions++;c.elements+=result.length;}
 return result;
}
globalThis.__acCall=(id,fn,args,method)=>{counters[id].calls++;try{return record(id,Reflect.apply(fn,undefined,args),args,method);}catch(error){counters[id].throws++;throw error;}};
globalThis.__acMethod=(id,receiver,name,argsThunk,method)=>{
 counters[id].calls++;
 try{const fn=receiver[name],args=argsThunk();return record(id,Reflect.apply(fn,receiver,args),args,method);}
 catch(error){counters[id].throws++;throw error;}
};
globalThis.__acSpread=(id,value)=>{const c=counters[id];c.calls++;c.executions++;c.elements+=value.length;return value;};
process.on('exit',code=>{
 const sites=Object.entries(counters).map(([id,c])=>({id,...c,meanLength:c.executions?c.elements/c.executions:0,meanElementsPerCall:c.calls?c.elements/c.calls:0}));
 const metadata=new Map(input.sites.map(s=>[s.id,s]));
 const scopeTotals={};for(const [name,predicate] of [['requestedFamilies',s=>!metadata.get(s.id).additionalCoreHelper],['additionalCoreHelpers',s=>metadata.get(s.id).additionalCoreHelper]]){const rows=sites.filter(predicate);scopeTotals[name]={calls:rows.reduce((n,s)=>n+s.calls,0),executions:rows.reduce((n,s)=>n+s.executions,0),elements:rows.reduce((n,s)=>n+s.elements,0)};}
 fs.writeFileSync(process.env.ARRAY_COPIES_OUTPUT,JSON.stringify({scopeTotals,exitCode:code,node:process.version,peakRssKiB:process.resourceUsage().maxRSS,totalCalls:sites.reduce((n,s)=>n+s.calls,0),totalThrows:sites.reduce((n,s)=>n+s.throws,0),totalExecutions:sites.reduce((n,s)=>n+s.executions,0),totalElements:sites.reduce((n,s)=>n+s.elements,0),sites:sites.sort((a,b)=>b.executions-a.executions||a.id.localeCompare(b.id))},null,2)+'\n');
});
