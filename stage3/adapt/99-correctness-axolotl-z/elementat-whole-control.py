import hashlib,json,os,signal,subprocess,time
from pathlib import Path
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/elementat-b');out=s/'whole-control';begin=time.monotonic()
summary=json.loads((repo/'stage3/adapt/99-correctness-axolotl-z/ELEMENTAT-COUNT.json').read_text());assert summary['complete'] and len(summary['completedUnits'])==398
profile=s/'whole-control-tasks.jsonl';assert not profile.exists()
env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-source',STAGE3_TASK_PROFILE=str(profile),NODE_DISABLE_COMPILE_CACHE='1',NODE_OPTIONS='--require='+str(repo/'stage3/oracle/profile-tasks.cjs'))
env.pop('EA_CALL_DIR',None);env.pop('STAGE3_ADAPTED_TREE',None);env.pop('STAGE3_ARTIFACT',None)
with (s/'whole-control.log').open('w') as log:
 child=subprocess.Popen(['bash',str(repo/'stage3/lane/run.sh'),str(out)],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 try:code=child.wait(timeout=1400)
 except BaseException:
  rows=[r.split(None,2) for r in subprocess.check_output(['ps','-eo','pid,ppid,comm'],text=True).splitlines()[1:]];ids={child.pid}
  while True:
   expanded=ids|{int(p) for p,parent,_ in rows if int(parent) in ids}
   if expanded==ids:break
   ids=expanded
  for pid in sorted(ids,reverse=True):
   try:os.kill(pid,signal.SIGKILL)
   except ProcessLookupError:pass
  child.wait();raise
result={'exit':code,'seconds':time.monotonic()-begin,'counted':False,'purpose':'unchanged stage3/lane/run.sh control and exact test-union comparison','profile':str(profile)}
(s/'whole-control-result.json').write_text(json.dumps(result,indent=2)+'\n');assert code==0,result
import sys
sys.path.insert(0,str(repo/'stage3/lane'))
from shard_plan import load_plan,names_digest
plan=load_plan();assert [u['name'] for u in plan['units']]==summary['completedUnits']
records=[];counts={'pass':0,'fail':0,'skip':0};diff=b''
for unit in plan['units']:
 directory=s/'count-results'/unit['name'];r=json.loads((directory/'report.json').read_text());assert r['status']=='pass' and r['counts']==unit['expect']
 observed=json.loads((directory/'observations.json').read_text());assert names_digest(observed)==unit['names_digest'];records+=observed
 for k,n in r['counts'].items():counts[k]+=n
 diff+=(directory/'oracle/baseline.diff').read_bytes()
whole_records=[dict(record,task=row['task'],status=status) for row in map(json.loads,profile.read_text().splitlines()) for status,key in [('pass','passes'),('fail','errors')] for record in row[key]]
assert names_digest(records)==names_digest(whole_records)
assert counts==plan['counts']==summary['oracleCounts']
assert diff==(out/'oracle/baseline.diff').read_bytes()
result.update(shardUnionIdentical=True,units=len(plan['units']),tests=len(records),counts=counts,namesDigest=names_digest(records),baselineDiffByteIdentical=True)
(s/'whole-control-result.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result),flush=True)
