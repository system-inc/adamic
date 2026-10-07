import {execFileSync} from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import {statistics,subtractStartup} from './statistics.mjs';
const dir=path.resolve('scratch/emitter-speed/sieve');
const oracle=path.resolve('oracle/node.mjs');
const node=[process.execPath,'--disable-warning=ExperimentalWarning',oracle];
const commands={before:[dir+'/native'],final:[dir+'/native-final'],js:[...node,dir+'/driver.mjs'],node:[...node,dir+'/driver.a']};
const K=1000, rounds=3;
const label=process.argv[2] ?? 'final-million';
const corpus=process.argv[3] ?? dir+'/corpus.txt';
function sample(name,k,label){const prefix=dir+'/'+(process.argv[2] ?? 'final-million')+'-'+label;execFileSync(dir+'/wait4',[prefix+'.json',prefix+'.stdout',prefix+'.stderr','300000',...commands[name],corpus,String(k)]);const r=JSON.parse(fs.readFileSync(prefix+'.json'));r.stdout=fs.readFileSync(prefix+'.stdout','utf8');r.stderr=fs.readFileSync(prefix+'.stderr','utf8');if(r.exitCode!==0||r.timedOut||r.stderr)throw Error(JSON.stringify(r));return r;}
const load=()=>fs.readFileSync('/proc/loadavg','utf8').trim();
const lines=fs.readFileSync(corpus,'utf8').trimEnd().split('\n').length;
const report={K,rounds,corpus,lines,loadBefore:load(),quota:fs.readFileSync('/sys/fs/cgroup/cpu.max','utf8').trim(),commands,samples:[]};
const want=sample('node',K,'expected').stdout;
for(let round=0;round<rounds;round++) {const names=Object.keys(commands);const order=names.slice(round).concat(names.slice(0,round));for(const name of order){const s={};for(const phase of round%2?['measured','startup']:['startup','measured']){s[phase]=sample(name,phase==='startup'?0:K,`${round}-${name}-${phase}`);if(s[phase].stdout!==(phase==='startup'?'0:0:2166136261:333555777\n':want))throw Error('checksum mismatch');}report.samples.push({round,name,...s,metrics:subtractStartup(s.measured,s.startup,K*lines)});}}
report.loadAfter=load();report.summary={};for(const name of Object.keys(commands)){report.summary[name]=statistics(report.samples.filter(s=>s.name===name).map(s=>s.metrics.wallMsPerRequest));}fs.writeFileSync(dir+'/'+label+'-timings.json',JSON.stringify(report,null,2));console.log(JSON.stringify(report.summary));
