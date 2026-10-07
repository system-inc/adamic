"""Quiet fresh-process medians; includes load, rule execution and serialization."""
import argparse,json,pathlib,statistics,subprocess,time
p=argparse.ArgumentParser();p.add_argument('--artifacts',required=True);p.add_argument('--baseline',required=True);p.add_argument('--native',required=True);p.add_argument('--oracle',required=True);a=p.parse_args()
out=pathlib.Path(a.artifacts).resolve();out.mkdir(parents=True,exist_ok=True);baseline=pathlib.Path(a.baseline)
rows=json.loads((baseline/'results.json').read_text());samples=[]
for case in rows:
 if case['description'] not in ['compiler','repository']:continue
 for round in range(3):
  for variant,binary in ([('go',a.oracle),('native',a.native)] if round%2==0 else [('native',a.native),('go',a.oracle)]):
   name=case['name']+'-'+str(round)+'-'+variant
   with (out/(name+'.stdout')).open('wb') as stdout,(out/(name+'.stderr')).open('wb') as stderr:
    before=time.perf_counter();proc=subprocess.run([binary,case['config'],case['manifest'],*case['flags']],stdout=stdout,stderr=stderr);seconds=time.perf_counter()-before
   data=(out/(name+'.stdout')).read_bytes();errors=(out/(name+'.stderr')).read_bytes();assert proc.returncode==0 and data==(baseline/(case['name']+'-go.stdout')).read_bytes()
   if variant=='native':assert errors==b''
   samples.append(dict(corpus=case['description'],rule=case['flags'][0][2:],variant=variant,round=round,seconds=seconds))
results=[]
for case in rows:
 if case['description'] not in ['compiler','repository']:continue
 values={variant:statistics.median(sample['seconds'] for sample in samples if sample['corpus']==case['description'] and sample['rule']==case['flags'][0][2:] and sample['variant']==variant) for variant in ['go','native']}
 result=dict(corpus=case['description'],rule=case['flags'][0][2:],**values,ratio=values['native']/values['go']);results.append(result);print(json.dumps(result),flush=True)
(out/'results.json').write_text(json.dumps(dict(samples=samples,medians=results),indent=2)+'\n')
print('PASS',flush=True)
