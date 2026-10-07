from pathlib import Path
import os, subprocess, time, statistics, json, hashlib
root=Path('/workspace/wave-09-final')
rows=[]
for name,config,manifest in [('compiler','/workspace/TypeScript-050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/tsconfig.json','/workspace/wave-09-validation/compiler.manifest'),('repository','/workspace/adamic/tsconfig.json','/workspace/wave-09-validation/repository.manifest')]:
 for round_ in range(3):
  outputs=[]
  for impl in (['native','go'] if round_%2==0 else ['go','native']):
   binary=root/('wave09' if impl=='native' else 'wave09-oracle')
   stdout=root/f'bench-{name}-{round_+1}-{impl}.stdout'
   stderr=stdout.with_suffix('.stderr')
   env=dict(os.environ)
   if impl=='native': env['ADAMIC_TSGO_TIMING']='1'
   with stdout.open('wb') as out,stderr.open('wb') as err:
    started=time.perf_counter_ns(); subprocess.run([str(binary),config,manifest],stdout=out,stderr=err,env=env,check=True); elapsed=time.perf_counter_ns()-started
   data=stdout.read_bytes();outputs.append(data)
   rows.append(dict(corpus=name,round=round_+1,implementation=impl,process_ns=elapsed,bytes=len(data),sha256=hashlib.sha256(data).hexdigest(),timing=stderr.read_text().strip()))
  assert outputs[0]==outputs[1]
medians={}
for name in ['compiler','repository']:
 medians[name]={impl:statistics.median(r['process_ns'] for r in rows if r['corpus']==name and r['implementation']==impl)/1e9 for impl in ['native','go']}
(root/'measurements.json').write_text(json.dumps(dict(rounds=rows,median_process_seconds=medians),indent=2)+'\n')
print(json.dumps(medians,indent=2))
