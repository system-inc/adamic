'use strict';
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');
const [table,directory,oracleReport]=process.argv.slice(2);
const data=JSON.parse(fs.readFileSync(table,'utf8'));
const files=fs.readdirSync(directory).filter(f=>f.endsWith('.json')).sort();
const totals=new Map();
for(const file of files){
 const observation=JSON.parse(fs.readFileSync(path.join(directory,file),'utf8'));
 for(const site of observation.sites){
  if(!data.sites[site.id]||data.sites[site.id].form==='concise arrow body')throw Error('unexpected coverage id');
  if(!totals.has(site.id))totals.set(site.id,{hits:0,values:new Map()});
  const record=totals.get(site.id);record.hits+=site.hits;
  for(const value of site.values)record.values.set(JSON.stringify(value),value);
 }
}
for(const [id,site] of data.sites.entries()){
 if(site.judgment?.kind==='adaptation') { const proof=site.judgment.proof;site.judgment.proof={file:site.judgment.rewrite.file,beforeSha256:proof.beforeSha256,afterSha256:proof.afterSha256,byteIdentical:proof.byteIdentical}; }
 if(site.form==='concise arrow body'){site.runtime={instrumented:false,reason:'Reserved for the language lead.'};continue;}
 if(site.callee==='sys.exit')site.implementationEvidence={exhaustive:false,known:[{file:'src/compiler/sys.ts',line:1588,text:'exit(exitCode?: number): void { disableCPUProfiler(() => process.exit(exitCode)); }'},{file:'src/harness/fakesHosts.ts',line:131,text:'exit(exitCode?: number): void { this.exitCode = exitCode; throw processExitSentinel; }'},{file:'src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts',line:1283,text:'exit(exitCode?: number): void { this.exitCode = exitCode; throw new Error(exitMessage); }'},{file:'src/testRunner/unittests/helpers/tsc.ts',line:55,text:'sys.exit = exitCode => sys.exitCode = exitCode;'},{file:'src/testRunner/unittests/tsc/cancellationToken.ts',line:57,text:'sys.exit = exitCode => sys.exitCode = exitCode;'},{file:'src/testRunner/unittests/tscWatch/incremental.ts',line:52,text:'if (incremental) sys.exit = exitCode => sys.exitCode = exitCode;'},{file:'src/testRunner/unittests/tscWatch/listFilesOnly.ts',line:12,text:'sys.exit = exitCode => sys.exitCode = exitCode;'}],inference:'The expression-bodied test overrides return their assigned numeric exit code. The production implementation can terminate or return undefined while profiling completes; the throwing test implementations yield no completed read.'};
 const record=totals.get(id);
 site.runtime={instrumented:true,hits:record?.hits||0,distinctValues:[...(record?.values.values()||[])]};
 if(site.judgment.kind==='unclassified')site.judgment.reason='The annotation-only no-value-return proof does not cover this callee or forwarding expression. Neither a byte-identical adaptation nor impossibility of one has been established. Runtime values are observations, not a proof over all implementations or executions.';
}
const oracle=JSON.parse(fs.readFileSync(oracleReport,'utf8'));
data.runtimeCoverage={command:'VOID_READS_COVERAGE=/tmp/void-reads-coverage NODE_OPTIONS=--require=/workspace/adamic/stage3/census/void-reads/observe.cjs timeout 720 bash stage3/oracle/run.sh /tmp/void-reads-runtime /tmp/void-reads-oracle --workers 4 --limit-seconds 540',processFiles:files.length,hitSites:totals.size,totalHits:[...totals.values()].reduce((sum,r)=>sum+r.hits,0),oracle,limitations:['Only completed expression evaluations are counted; a throw or process exit before a result produces no value hit.','Zero hits mean unobserved, not unreachable.','Identity keys for objects, functions and symbols are process-local; primitive values are recorded exactly.','Finite corpus observations do not establish all possible values of host callbacks.']};
if(process.argv[5]){
 const control=JSON.parse(fs.readFileSync(process.argv[5],'utf8'));
 const originalDiff=fs.readFileSync(path.join(path.dirname(process.argv[5]),'baseline.diff'));
 const instrumentedDiff=fs.readFileSync(path.join(path.dirname(oracleReport),'baseline.diff'));
 data.runtimeCoverage.apiBaselineControl={status:control.status,counts:control.counts,baselineDiffs:control.baseline_diffs,baselineDiffSha256:crypto.createHash('sha256').update(originalDiff).digest('hex'),identicalToInstrumentedDiff:originalDiff.equals(instrumentedDiff),command:control.phases.tests.command};
 if(!originalDiff.equals(instrumentedDiff))throw Error('API control differs');
}
fs.writeFileSync(table,JSON.stringify(data,null,2)+'\n');
console.log(JSON.stringify({hitSites:totals.size,totalHits:data.runtimeCoverage.totalHits,nonUndefined:data.sites.filter(s=>s.runtime.distinctValues?.some(v=>v.type!=='undefined')).map(s=>({file:s.file,line:s.line,values:s.runtime.distinctValues}))}));
