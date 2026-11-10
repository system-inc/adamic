from pathlib import Path
import subprocess,os,time,json,collections,shutil
from concurrent.futures import ThreadPoolExecutor
s=Path('/workspace/scratch/each-node');repo=Path('/workspace/adamic');plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());results=s/'lane';results.mkdir(exist_ok=True)
def run(item):
 name=item['name'];counts=s/'counts'/name;counts.mkdir(parents=True);env=dict(os.environ,STAGE3_ADAPTED_TREE=str(s/'measure'),STAGE3_SHARD_SECONDS='90',NODE_OPTIONS='--require='+str(s/'preload.cjs'),NODE_DISABLE_COMPILE_CACHE='1',EACH_NODE_COUNTS=str(counts),PYTHONDONTWRITEBYTECODE='1');start=time.monotonic()
 with (s/(name+'.log')).open('w') as f:p=subprocess.run(['python3',str(repo/'stage3/lane/shard.py'),name,str(results)],cwd=repo,env=env,stdout=f,stderr=subprocess.STDOUT)
 report=json.loads((results/name/'report.json').read_text());rows=[json.loads(p.read_text()) for p in sorted(counts.glob('*.json'))];record=dict(unit=name,exit=p.returncode,seconds=time.monotonic()-start,report=report,processes=rows);(s/(name+'-result.json')).write_text(json.dumps(record)+'\n');shutil.rmtree(results/name/'adapted-tree',ignore_errors=True);return record
records=[]
with ThreadPoolExecutor(max_workers=4) as pool:
 for r in pool.map(run,plan['units']):
  records.append(r);print(json.dumps({k:r[k] for k in ['unit','exit','seconds']}),flush=True)
  if r['exit'] or r['report']['status']!='pass' or not r['processes']:raise RuntimeError(r['unit']+' failed: '+str(r['report']['errors']))
assert [r['unit'] for r in records]==[u['name'] for u in plan['units']]
oracle=collections.Counter();totals=collections.Counter();functions={}
for r in records:
 oracle.update(r['report']['counts'])
 for p in r['processes']:
  for k in ['defined','throws','exposed']:totals[k]+=p[k]
  for f in p['functions']:
   row=functions.setdefault(f['source'],dict(f,calls=0,throws=0));row['calls']+=f['calls'];row['throws']+=f['throws']
assert dict(oracle)==plan['counts']
result=dict(acceptedUnits=len(records),oracle=dict(oracle),totals=dict(totals),functions=list(functions.values()),complete=True)
(s/'lane-summary.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
if totals['exposed']:raise SystemExit('STOP: second-argument exposure is nonzero')
