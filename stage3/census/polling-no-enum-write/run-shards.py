#!/usr/bin/env python3
"""Run every manifest shard in its own bounded process; retain child exits."""
import concurrent.futures,json,os,pathlib,subprocess,time
repo=pathlib.Path('/workspace/adamic')
here=pathlib.Path(__file__).resolve().parent
results=pathlib.Path('/workspace/scratch/polling-shards');results.mkdir()
logs=here/'shard-logs';logs.mkdir()
units=json.loads((repo/'stage3/lane/shards.json').read_text())['units']
env=dict(os.environ,STAGE3_ADAPTED_TREE='/workspace/scratch/polling-after',STAGE3_SHARD_SECONDS='1700')
def run(unit):
 name=unit['name'];start=time.monotonic()
 with (logs/(name+'.log')).open('w') as log:
  try: code=subprocess.run(['python3','stage3/lane/shard.py',name,str(results)],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=1750).returncode
  except subprocess.TimeoutExpired: code=124
 row=dict(name=name,exit=code,seconds=time.monotonic()-start)
 p=results/name/'report.json'
 if p.exists():
  report=json.loads(p.read_text());row.update(counts=report['counts'],status=report['status'],errors=report['errors'])
  (logs/(name+'.json')).write_text(p.read_text())
 return row
rows=[]
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
 for row in pool.map(run,units):
  rows.append(row)
  (here/'shard-exits.json').write_text(json.dumps(rows,indent=2)+'\n')
  print(len(rows),row['name'],row['exit'],flush=True)
counts={k:sum(r.get('counts',{}).get(k,0) for r in rows) for k in ('pass','fail','skip')}
passed=len(rows)==398 and all(r['exit']==0 and r.get('status')=='pass' and r['seconds']<1800 for r in rows) and counts=={'pass':106366,'fail':1,'skip':0}
(here/'shard-summary.json').write_text(json.dumps(dict(status='pass' if passed else 'fail',units=len(rows),counts=counts),indent=2)+'\n')
raise SystemExit(0 if passed else 1)
