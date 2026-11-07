"""Count one bounded lane piece using the unchanged shard runner and guards."""
import concurrent.futures, collections, hashlib, json, os, pathlib, signal, subprocess, sys, time
repo=pathlib.Path('/workspace/adamic'); scratch=pathlib.Path(os.environ.get('MM_SCRATCH','/workspace/scratch/multimap/count-v3')); tree=pathlib.Path('/workspace/scratch/multimap/counted'); output=repo/'stage3/adapt/99-correctness-core-multimap'; start=int(sys.argv[1]); count=int(sys.argv[2]); piece=scratch/f'count-{start:03d}{os.environ.get("MM_ATTEMPT_SUFFIX","")}';piece.mkdir(exist_ok=False);(piece/'calls').mkdir()
units=json.loads((repo/'stage3/lane/shards.json').read_text())['units'][start:start+count]
env=dict(os.environ,STAGE3_ADAPTED_TREE=str(tree),STAGE3_RESULTS=str(scratch/'results'),STAGE3_SHARD_SECONDS='90',NODE_OPTIONS='--require='+str(output/'measure-preload.cjs'),MM_BODIES='/workspace/scratch/multimap/bodies.json',MM_COUNTS=str(piece/'calls'),NODE_DISABLE_COMPILE_CACHE='1')
def run(unit):
 name=unit['name'];begin=time.monotonic()
 with (piece/(name+'.log')).open('w') as log:
  p=subprocess.Popen(['python3',str(repo/'stage3/lane/shard.py'),name,str(scratch/'results')],cwd=repo,env=dict(env,MM_SHARD=name),stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=p.wait(timeout=110)
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 report=scratch/'results'/name/'report.json';report=json.loads(report.read_text()) if report.exists() else {}
 view=scratch/'results'/name/'adapted-tree'
 if view.exists():
  import shutil
  shutil.rmtree(view)
 row={'unit':name,'exit':code,'seconds':time.monotonic()-begin,'report':report};(piece/(name+'.json')).write_text(json.dumps(row,indent=2)+'\n');print(json.dumps({'unit':name,'exit':code,'status':report.get('status')}),flush=True)
 return row
begin=time.monotonic()
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:rows=list(pool.map(run,units))
counts=collections.Counter();failures=[]
for p in (piece/'calls').glob('*.json'):
 r=json.loads(p.read_text());counts.update(r['counts']);failures+=r['failures']
result={'start':start,'end':start+len(units),'seconds':time.monotonic()-begin,'units':rows,'counts':dict(counts),'failures':failures,'complete':all(r['exit']==0 and r['report'].get('status')=='pass' for r in rows)}
(piece/'summary.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({k:result[k] for k in ['start','end','seconds','counts','complete']}),flush=True)
assert result['complete'] and not failures and counts['enumerations']==0
