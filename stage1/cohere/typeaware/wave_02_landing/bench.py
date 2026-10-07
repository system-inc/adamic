#!/usr/bin/env python3
"""Time already-verified landing binaries, alternating order and checking counts."""
import argparse,json,statistics,subprocess,time
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('--artifacts',required=True);p.add_argument('--typescript-source',required=True);args=p.parse_args()
own=Path(__file__).resolve().parent;root=own.parents[3];artifacts=Path(args.artifacts).resolve();output=artifacts/'isolated-bench';output.mkdir(exist_ok=True)
records=[]
for suite in ['original','continuation1','continuation2','continuation3','continuation4']:
 folder=artifacts/suite
 native=folder/('native' if suite in ['continuation3','continuation4'] else 'coverage')
 go=folder/('oracle' if suite in ['continuation3','continuation4'] else 'coverage-oracle')
 for corpus,config,manifest in [('compiler',Path(args.typescript_source)/'src/compiler/tsconfig.json',artifacts/'continuation3/compiler.manifest'),('repository',root/'tsconfig.json',root/'stage1/cohere/typeaware/validation-volume/repository.manifest')]:
  for round in range(3):
   counts=[]
   for implementation in (['native','go'] if round%2 else ['go','native']):
    name=f'{suite}-{corpus}-{round+1}-{implementation}';command=[str(native if implementation=='native' else go),str(config),str(manifest),'--count'];start=time.perf_counter_ns()
    with (output/(name+'.stdout')).open('wb') as stdout,(output/(name+'.stderr')).open('wb') as stderr:result=subprocess.run(command,cwd=root,stdout=stdout,stderr=stderr)
    elapsed=time.perf_counter_ns()-start;assert result.returncode==0,(name,result.returncode)
    if implementation=='native':assert not (output/(name+'.stderr')).read_bytes(),name
    counts.append((output/(name+'.stdout')).read_bytes());records.append({'suite':suite,'corpus':corpus,'round':round+1,'implementation':implementation,'args':command,'exit':result.returncode,'process_ns':elapsed})
   assert counts[0]==counts[1],(suite,corpus,'count mismatch')
  print(suite,corpus,*(f'{implementation}={statistics.median(r["process_ns"] for r in records if r["suite"]==suite and r["corpus"]==corpus and r["implementation"]==implementation)/1e9:.9f}s' for implementation in ['native','go']),flush=True)
(output/'timings.json').write_text(json.dumps(records,indent=2)+'\n')
