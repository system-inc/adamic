from pathlib import Path
import subprocess,time,os,json,statistics,hashlib
root=Path('/workspace/wave-27-scratch'); artifacts=root/'final-source'; dest=root/'quiet-timing';dest.mkdir(exist_ok=True)
results={}
for name,config in [('repository','/workspace/adamic/tsconfig.json'),('compiler',str(root/'typescript/src/compiler/tsconfig.json'))]:
 runs={'native':[],'go':[]}; truths=[]
 for i in range(3):
  for engine in (['native','go'] if i%2==0 else ['go','native']):
   binary=artifacts/('wave27' if engine=='native' else 'wave27-oracle')
   env=dict(os.environ,ADAMIC_TSGO_TIMING='1')
   with (dest/f'{name}-{i}-{engine}.stdout').open('wb') as out,(dest/f'{name}-{i}-{engine}.stderr').open('wb') as err:
    start=time.perf_counter_ns();subprocess.run([str(binary),config,str(root/f'{name}.manifest')],stdout=out,stderr=err,env=env,check=True);elapsed=time.perf_counter_ns()-start
   data=(dest/f'{name}-{i}-{engine}.stdout').read_bytes();truths.append(data);runs[engine].append(elapsed)
 assert all(data==truths[0] for data in truths)
 medians={engine:statistics.median(values) for engine,values in runs.items()}
 results[name]={'process_ns':runs,'median_ns':medians,'native_over_go':medians['native']/medians['go'],'bytes':len(truths[0]),'sha256':hashlib.sha256(truths[0]).hexdigest()}
(dest/'timing.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(results,indent=2))
