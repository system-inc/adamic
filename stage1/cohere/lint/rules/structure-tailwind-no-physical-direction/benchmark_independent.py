#!/usr/bin/env python3
"""Startup-inclusive findings throughput with the independent parser on supported positive inputs."""
import argparse,json,subprocess,time,statistics
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[4]
p=argparse.ArgumentParser();p.add_argument('--scratch',type=Path,required=True);a=p.parse_args();S=a.scratch.resolve();copy=S/'independent-non-jsx';rows=json.loads((S/'non-jsx-upstream.json').read_text());records=[]
def run(label,cmd):
 with (S/(label+'.log')).open('wb') as out,(S/(label+'.stderr')).open('wb') as err:
  start=time.perf_counter();r=subprocess.run(list(map(str,cmd)),cwd=ROOT,stdout=out,stderr=err);seconds=time.perf_counter()-start
 assert r.returncode==0 and not (S/(label+'.stderr')).read_bytes(),label
 return int((S/(label+'.log')).read_text()),seconds
for name in sorted({r['rule'] for r in rows}):
 for row in rows:
  if row['rule']!=name:continue
  raw=S/'bench-own.json';value={k:row[k] for k in ('name','source','rule','options')};raw.write_text(json.dumps([value]));count,_=run('bench-own-positive',[S/'oracle',raw,'--count'])
  if count>0:break
 assert count>0;raw.write_text(json.dumps([value]*200));record={'rule':name,'cases':200,'source':value['source'],'scope':'startup-inclusive, raw-source input and own parsing; sanitized native','samples':[]}
 driver=copy/'stage1/cohere/lint/rules'/HERE.name/'validation.ts'
 for backend,cmd in [('Go',[S/'oracle',raw,'--count']),('Node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',driver,raw,'--parse','--count']),('native',[copy/'native',raw,'--parse','--count'])]:
  times=[];counts=[]
  for index in range(3):
   n,seconds=run('bench-own-'+backend+'-'+str(index),cmd);counts.append(n);times.append(seconds)
  assert counts==[count*200]*3
  record['samples'].append({'backend':backend,'findings':counts[0],'seconds':times,'findingsPerSecond':counts[0]/statistics.median(times)})
 records.append(record)
(S/'independent-benchmarks.json').write_text(json.dumps(records,indent=2));print(json.dumps(records,indent=2),flush=True)
