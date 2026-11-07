import hashlib,json,os,shutil,signal,subprocess,time
from concurrent.futures import ThreadPoolExecutor
from collections import Counter
from pathlib import Path
s=Path('/workspace/scratch/core-extend');repo=Path('/workspace/adamic');plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());results=s/'count-results';results.mkdir(exist_ok=True)
while not (s/'count-preparation.json').exists() or 'harness' not in json.loads((s/'count-preparation.json').read_text()):
 print('waiting for counted harness',flush=True);time.sleep(30)
assert all(p['exit']==0 for p in json.loads((s/'count-preparation.json').read_text()).values())
summary={'complete':False,'completedUnits':[],'pieces':[],'calls':{},'shapes':{},'comparisonFailures':0,'uncompared':0,'oracleCounts':{'pass':0,'fail':0,'skip':0}}
if (s/'count-summary.json').exists():summary=json.loads((s/'count-summary.json').read_text())
def run(pair):
 unit,piece=pair;name=unit['name'];env=dict(os.environ,STAGE3_ADAPTED_TREE=str(s/'counted'),STAGE3_RESULTS=str(results),COPY_CALL_DIR=str(piece/'calls'),COPY_PIECE=name,STAGE3_SHARD_SECONDS='90',NODE_OPTIONS='--require='+str(s/'preload.cjs'),NODE_DISABLE_COMPILE_CACHE='1')
 with (piece/(name+'.log')).open('w') as log:
  p=subprocess.Popen(['python3',str(repo/'stage3/lane/shard.py'),name,str(results)],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=p.wait(timeout=110)
  except subprocess.TimeoutExpired:code=124
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 report=json.loads((results/name/'report.json').read_text()) if (results/name/'report.json').exists() else {'status':'missing'}
 return name,code,report
for start in range(len(summary['completedUnits']),len(plan['units']),20):
 begin=time.monotonic();piece=s/f'piece-{start//20+1:02d}';piece.mkdir();(piece/'calls').mkdir();units=plan['units'][start:start+20]
 with ThreadPoolExecutor(max_workers=2) as pool:observed=list(pool.map(run,[(u,piece) for u in units]))
 failures=[{'name':n,'exit':c,'report':r} for n,c,r in observed if c or r['status']!='pass']
 for f in sorted((piece/'calls').glob('*.jsonl')):
  for line in f.open():
   row=json.loads(line);owner=row['owner'];summary['calls'][owner]=summary['calls'].get(owner,0)+1
   summary['uncompared']+=int(row['comparisons'].get('uncompared',False))
   summary['comparisonFailures']+=int(any(v is False for v in row['comparisons'].values()))
   for i,shape in enumerate(row['shapes']):
    key=owner+':'+str(i+1);counts=summary['shapes'].setdefault(key,{'plain':0,'notPlain':0,'inheritedEnumerable':0,'accessors':0,'symbolKeys':0})
    counts['plain' if shape['plain'] else 'notPlain']+=1
    for field,src in [('inheritedEnumerable','inheritedEnumerable'),('accessors','accessors'),('symbolKeys','symbols')]:counts[field]+=bool(shape[src])
 record={'index':start//20+1,'seconds':time.monotonic()-begin,'units':[u['name'] for u in units],'failures':failures};summary['pieces'].append(record)
 if not failures:
  summary['completedUnits'] += [u['name'] for u in units]
  for _,_,r in observed:
   for k,v in r['counts'].items():summary['oracleCounts'][k]+=v
 for n,_,_ in observed:
  shutil.copytree(results/n,piece/n,ignore=shutil.ignore_patterns('adapted-tree'))
  shutil.rmtree(results/n/'adapted-tree',ignore_errors=True)
 summary['complete']=len(summary['completedUnits'])==len(plan['units'])
 (s/'count-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
 print(json.dumps({'piece':record['index'],'seconds':record['seconds'],'completed':len(summary['completedUnits']),'calls':summary['calls'],'comparisonFailures':summary['comparisonFailures'],'uncompared':summary['uncompared'],'failedUnits':len(failures)}),flush=True)
 if failures:raise SystemExit('Count stopped on failed unit')
assert summary['oracleCounts']==plan['counts']
