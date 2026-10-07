import { measure, requireChecksum } from './harness.mjs';
import { subtractStartup, statistics } from './statistics.mjs';
import { writeFileSync } from 'node:fs';
import { execFileSync } from 'node:child_process';
const directory='/workspace/scratch/stack-check-scc/handler';
const options={repo:'/workspace/adamic',timeoutMs:300000};
const experiment={directory,options,helper:directory+'/wait4'};
const workload={requestsPerIteration:3,corpus:directory+'/sieve.txt',commands:{
 before:[directory+'/native-before'],after:[directory+'/native-after'],
 node:[process.execPath,'--disable-warning=ExperimentalWarning','/workspace/adamic/oracle/node.mjs',directory+'/driver.a']
}};
const report={iterations:1000,rounds:3,requests:3000,nativeFlags:'native.Flags(Options{}): -O2, no sanitizers or counting',harnessCommit:'0a3e651',loadBefore:execFileSync('uptime',{encoding:'utf8'}),checks:[],samples:[]};
const oracle=await measure(experiment,workload,'node',1,'node-check');
report.checks.push(oracle);
for(const side of ['before','after']) {
 const check=await measure(experiment,workload,side,1,side+'-check');
 requireChecksum(check,oracle.stdout,side+' preflight');report.checks.push(check);
}
const reference=await measure(experiment,workload,'node',1000,'node-K');report.checks.push(reference);
for(let round=0;round<3;round++)for(const side of round%2?['after','before']:['before','after']) {
 const pair={round,side};
 for(const phase of round%2?['measured','startup']:['startup','measured']) {
  pair[phase]=await measure(experiment,workload,side,phase==='startup'?0:1000,`${side}-${round}-${phase}`);
  requireChecksum(pair[phase],phase==='startup'?'0:0:2166136261:333555777\n':reference.stdout,side+' '+phase);
 }
 pair.metrics=subtractStartup(pair.measured,pair.startup,3000);report.samples.push(pair);
}
report.loadAfter=execFileSync('uptime',{encoding:'utf8'});
report.summary={};
for(const side of ['before','after']) {
 const samples=report.samples.filter(x=>x.side===side);
 report.summary[side]={wall:statistics(samples.map(x=>x.metrics.wallMsPerRequest)),cpu:statistics(samples.map(x=>x.metrics.cpuMsPerRequest))};
}
writeFileSync(directory+'/measurements.json',JSON.stringify(report,null,2)+'\n');
console.log(JSON.stringify(report.summary,null,2));
