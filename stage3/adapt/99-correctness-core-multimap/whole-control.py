import hashlib,json,os,signal,subprocess,time,sys
from pathlib import Path
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/multimap');audit=s/'exit-audit';out=s/'whole';begin=time.monotonic()
while True:
 summary=json.loads((audit/'COUNT.json').read_text())
 if summary['complete']:break
 assert time.monotonic()-begin<900,'count gate still pending'
 time.sleep(5)
assert summary['counterexamples']==summary['enumerations']==0
assert summary['counts']['calls']==4355837
records=[]
for directory in audit.glob('count-*'):
 if not directory.is_dir() or not (directory/'summary.json').exists():continue
 piece=json.loads((directory/'summary.json').read_text())
 for row in piece['units']:
  if row['exit']==0 and row['report'].get('status')=='pass':
   for p in (directory/'calls').glob(row['unit']+'.*.json'):
    record=json.loads(p.read_text());assert record['recorderVersion']==3 and record['exitEmissions']==1 and record['lastExitLateCalls']==0
    records.append(record)
(s/'exit-audit-proof.json').write_text(json.dumps({'passed':True,'processRecords':len(records),'allFlushedAfterExitListeners':True,'lateExitCalls':0,'fullRerunCountIdentical':True},indent=2)+'\n')
profile=s/'whole-tasks.jsonl';assert not profile.exists()
env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-source',STAGE3_TASK_PROFILE=str(profile),NODE_DISABLE_COMPILE_CACHE='1',NODE_OPTIONS='--require='+str(repo/'stage3/oracle/profile-tasks.cjs'))
for key in ['MM_COUNTS','MM_BODIES','STAGE3_ADAPTED_TREE','STAGE3_ARTIFACT']:env.pop(key,None)
print('Count gate passed; running ordinary final lane',flush=True)
with (s/'whole-control.log').open('w') as log:
 child=subprocess.Popen(['bash',str(repo/'stage3/lane/run.sh'),str(out)],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 try:code=child.wait(timeout=1500)
 finally:
  try:os.killpg(child.pid,signal.SIGKILL)
  except ProcessLookupError:pass
  child.wait()
result={'exit':code,'seconds':time.monotonic()-begin,'counted':False,'purpose':'unchanged stage3/lane/run.sh final-tree control and test-union comparison','profile':str(profile)}
(s/'whole-control-result.json').write_text(json.dumps(result,indent=2)+'\n');assert code==0,result
sys.path.insert(0,str(repo/'stage3/lane'))
from shard_plan import load_plan,names_digest
plan=load_plan();assert [u['name'] for u in plan['units']]==summary['acceptedUnits']
records=[];counts={'pass':0,'fail':0,'skip':0};diff=b''
for unit in plan['units']:
 directory=audit/'results'/unit['name'];r=json.loads((directory/'report.json').read_text());assert r['status']=='pass' and r['counts']==unit['expect']
 observed=json.loads((directory/'observations.json').read_text());assert names_digest(observed)==unit['names_digest'];records+=observed
 for k,n in r['counts'].items():counts[k]+=n
 diff+=(directory/'oracle/baseline.diff').read_bytes()
whole_records=[dict(record,task=row['task'],status=status) for row in map(json.loads,profile.read_text().splitlines()) for status,key in [('pass','passes'),('fail','errors')] for record in row[key]]
assert names_digest(records)==names_digest(whole_records)
assert counts==plan['counts']==summary['oracleCounts']
assert diff==(out/'oracle/baseline.diff').read_bytes()
result.update(shardUnionIdentical=True,units=len(plan['units']),tests=len(records),counts=counts,namesDigest=names_digest(records),baselineDiffByteIdentical=True)
(s/'whole-control-result.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
