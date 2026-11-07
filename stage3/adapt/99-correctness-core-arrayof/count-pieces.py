import hashlib,json,os,shutil,signal,subprocess,time
from pathlib import Path
from collections import Counter
from concurrent.futures import ThreadPoolExecutor
s=Path('/workspace/scratch/arrayof');repo=Path('/workspace/adamic');units=json.loads((repo/'stage3/lane/shard-plan.json').read_text())['units'];results=s/'results';results.mkdir(exist_ok=True)
def run(u):
 env=dict(os.environ,STAGE3_ADAPTED_TREE=str(s/'counted'),STAGE3_SHARD_SECONDS='90',ARRAYOF_CALL_DIR=str(piece/'calls'),ARRAYOF_UNIT=u['name'],NODE_OPTIONS='--require='+str(s/'preload.cjs'),NODE_DISABLE_COMPILE_CACHE='1')
 with (piece/(u['name']+'.log')).open('w') as log:
  p=subprocess.Popen(['python3',str(repo/'stage3/lane/shard.py'),u['name'],str(results)],env=env,cwd=repo,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:rc=p.wait(timeout=100)
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 report=json.loads((results/u['name']/'report.json').read_text());shutil.copytree(results/u['name'],piece/u['name'],ignore=shutil.ignore_patterns('adapted-tree'));shutil.rmtree(results/u['name']/'adapted-tree',ignore_errors=True)
 return dict(name=u['name'],exit=rc,report=report)
def snapshot():
 return {str(p.relative_to(s/'counted')):hashlib.sha256(p.read_bytes()).hexdigest() for base in ['src','built/local'] for p in sorted((s/'counted'/base).rglob('*')) if p.is_file()}
frozen=snapshot();(s/'counted-manifest.json').write_text(json.dumps(frozen,indent=2)+'\n')
summary=dict(complete=False,pieces=[],counts=dict.fromkeys(['non-negative integer','negative','fractional','NaN','infinite'],0),results=0,counterexamples=0,completedUnits=[])
for start in range(0,len(units),20):
 begun=time.monotonic();piece=s/f'piece-{start//20+1:02d}';piece.mkdir();(piece/'calls').mkdir()
 with ThreadPoolExecutor(max_workers=4) as pool:observed=list(pool.map(run,units[start:start+20]))
 assert snapshot()==frozen,'measurement source or emitted files changed'
 counts=Counter();returned=0;bad=0
 for f in (piece/'calls').glob('*.jsonl'):
  for line in f.open():
   row=json.loads(line)
   if row['event']=='call':counts[row['classification']]+=1
   else:returned+=1;bad+=not row['equal']
 record=dict(units=observed,seconds=time.monotonic()-begun,counts=dict(counts),results=returned,counterexamples=bad)
 (piece/'summary.json').write_text(json.dumps(record,indent=2)+'\n')
 for c,n in counts.items():summary['counts'][c]+=n
 summary['results']+=returned;summary['counterexamples']+=bad;summary['pieces'].append(dict(piece=piece.name,seconds=record['seconds'],counts=dict(counts),results=returned,counterexamples=bad))
 failed=[r for r in observed if r['exit']!=0 or r['report']['status']!='pass']
 if not failed:summary['completedUnits'] += [r['name'] for r in observed]
 summary['complete']=len(summary['completedUnits'])==len(units)
 (s/'COUNT.json').write_text(json.dumps(summary,indent=2)+'\n')
 print(json.dumps(summary['pieces'][-1]),flush=True)
 assert not failed,failed
 assert bad==0 and sum(counts.values())==returned
 assert all(counts[c]==0 for c in ['negative','fractional','NaN','infinite'])
print('COMPLETE',flush=True)
