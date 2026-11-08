const fs=require('fs'),path=require('path'),crypto=require('crypto'),assert=require('assert');
const out=__dirname,read=name=>JSON.parse(fs.readFileSync(path.join(out,name),'utf8'));
const stock=read('stock.json'),adapted=read('adapted.json'),runtime=read('runtime.json'),baseline=read('baseline.json'),memory=read('counted-memory.json'),provenance=read('provenance.json'),steps=read('run-steps.json');
const digest=data=>crypto.createHash('sha256').update(data).digest('hex');
for(const [name,data]of [['stock',stock],['adapted',adapted]]){
 assert.strictEqual(data.diagnostics.length,0,name+' diagnostics');
 assert.strictEqual(new Set(data.sites.map(s=>s.id)).size,data.sites.length,name+' site IDs');
 for(const grouping of ['byMethod','byFile'])for(const category of ['object','mixed','unresolved'])for(const field of ['inLoop','notInLoop','uncertainOuterLoop'])assert.strictEqual(Object.values(data[grouping]).reduce((n,c)=>n+c[category][field],0),data.totals[category][field],name+' '+grouping+' '+category+' '+field);
 for(const row of provenance.inputs[name].compilerFiles)assert.strictEqual(digest(fs.readFileSync(path.join(provenance.inputs[name].root,row.file))),row.sha256,name+' source changed: '+row.file);
}
const selected=stock.sites.filter(s=>s.depth&&s.classification!=='unresolved');
assert.deepStrictEqual(runtime.sites.map(s=>s.id).sort(),selected.map(s=>s.id).sort());
for(const [field,total]of [['calls','totalCalls'],['executions','totalExecutions'],['elements','totalElements'],['throws','totalThrows']])assert.strictEqual(runtime.sites.reduce((n,s)=>n+s[field],0),runtime[total]);
for(const s of runtime.sites){assert(s.executions<=s.calls);assert(s.elements>=0);assert.strictEqual(s.meanLength,s.executions?s.elements/s.executions:0);}
for(const result of [runtime,baseline,memory])assert.strictEqual(result.exitCode,0);
assert(steps.steps.every(s=>s.exitCode===0));assert(steps.seconds<40*60);
assert.strictEqual(fs.readFileSync(path.join(out,'baseline.log'),'utf8'),fs.readFileSync(path.join(out,'counted.log'),'utf8'));
assert.strictEqual(provenance.stock,'050880ce59e30b356b686bd3144efe24f875ebc8');
const result={passed:true,staticPrograms:{stock:{diagnostics:0,compilerFiles:stock.compilerFiles},adapted:{diagnostics:0,compilerFiles:adapted.compilerFiles}},sourceHashesUnchanged:true,staticTotalsReconcile:true,uniqueSiteIds:true,runtimeSelectedSites:selected.length,sitesWithCalls:runtime.sites.filter(s=>s.calls).length,sitesWithFreshResults:runtime.sites.filter(s=>s.executions).length,baselineAndCountedOutputEqual:true,runtimeExitCode:runtime.exitCode,baselineExitCode:baseline.exitCode,runtimeSeconds:steps.seconds,deadlineSeconds:40*60};
fs.writeFileSync(path.join(out,'verification.json'),JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result));
