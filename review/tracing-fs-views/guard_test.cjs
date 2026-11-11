const fs=require('fs'),cp=require('child_process'),assert=require('assert/strict');
const root='/workspace/adamic/stage3/adapt/',report=[];
for(const dir of ['62-temporary-tracing-write','63-temporary-tracing-legend']){
 const api=require(root+dir+'/adapt.cjs'),p=require(root+dir+'/pins.json');
 const text=cp.execFileSync('git',['-C','/workspace/tracing-before','show','refs/stage3/'+dir+'-before:src/compiler/tracing.ts'],{encoding:'utf8'});
 for(const newline of ['\r\n','\n']){
  const before=text.replace(/\r\n/g,'\n').replace(/\n/g,newline),after=api.plan(before);
  assert.equal(api.plan(after),after);assert(after.includes(p.callAfter));
  report.push({dir,newline:JSON.stringify(newline),status:'accept and idempotent'});
 }
 const old=p.binding==='legend'?'(fs as { writeFileSync(path: string, data: string | undefined): void }).writeFileSync(legendPath, JSON.stringify(legend))':'(fs as { writeSync(fd: number, data: string | undefined): number }).writeSync(typesFd, JSON.stringify(descriptor))';
 const mutants={view:text.replace(p.callBefore,old),ownerDrift:text.replace(p.callBefore,p.callBefore+'; void 0'),nonliteral:text.replace(p.literal,p.literal.replace(p.binding==='legend'?'[]':'{',p.binding==='legend'?'Array()':'Object.assign({}) /*')),toJSON:text.replace(p.binding==='legend'?'const legend: TraceRecord[] = [];':'const descriptor = {',p.binding==='legend'?'const legend: TraceRecord[] = [{toJSON: undefined}];':'const descriptor = { toJSON: undefined,'),reassigned:text.replace(p.callBefore,p.binding+' = undefined; '+p.callBefore),missingCall:text.replace(p.callBefore,'void 0'),duplicateOwner:text+'\n'+p.before};
 for(const [name,mutant] of Object.entries(mutants)){assert.throws(()=>api.plan(mutant),undefined,name);report.push({dir,name,status:'refused'})}
}
console.log(JSON.stringify(report,null,2));
