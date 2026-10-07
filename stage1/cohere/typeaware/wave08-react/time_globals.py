"""Quiet alternating fresh-process Globals medians, retaining full byte parity."""
import argparse,json,pathlib,statistics,subprocess,time
p=argparse.ArgumentParser();p.add_argument('--baseline',required=True);p.add_argument('--artifacts',required=True);a=p.parse_args()
base=pathlib.Path(a.baseline).resolve();out=pathlib.Path(a.artifacts).resolve();out.mkdir(parents=True,exist_ok=True)
repo=pathlib.Path(__file__).resolve().parents[4];samples=[]
for corpus in ['compiler','repository']:
 if corpus=='repository':config=repo/'tsconfig.json'
 else:
  first=pathlib.Path((base/'compiler.manifest').read_text().splitlines()[0])
  config=next(parent/'tsconfig.json' for parent in first.parents if parent.name=='compiler' and (parent/'tsconfig.json').exists())
 truth=(base/(corpus+'-go.stdout')).read_bytes()
 for round in range(3):
  for variant in (['oracle','native'] if round%2==0 else ['native','oracle']):
   name=corpus+'-'+str(round)+'-'+variant
   with (out/(name+'.stdout')).open('wb') as stdout,(out/(name+'.stderr')).open('wb') as stderr:
    start=time.perf_counter();result=subprocess.run([base/variant,config,base/(corpus+'.manifest')],stdout=stdout,stderr=stderr);elapsed=time.perf_counter()-start
   assert result.returncode==0 and (out/(name+'.stdout')).read_bytes()==truth
   if variant=='native':assert (out/(name+'.stderr')).read_bytes()==b''
   samples.append(dict(corpus=corpus,variant=variant,round=round,seconds=elapsed))
medians=[]
for corpus in ['compiler','repository']:
 values={variant:statistics.median(x['seconds'] for x in samples if x['corpus']==corpus and x['variant']==variant) for variant in ['oracle','native']}
 row=dict(corpus=corpus,**values,ratio=values['native']/values['oracle']);medians.append(row);print(json.dumps(row),flush=True)
(out/'results.json').write_text(json.dumps(dict(samples=samples,medians=medians),indent=2)+'\n')
print('PASS',flush=True)
