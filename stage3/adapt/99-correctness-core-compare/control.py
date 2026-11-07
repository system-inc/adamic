import hashlib,json,os,signal,subprocess,time,sys
from pathlib import Path
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/core-compare');out=s/'control';summary=json.loads((s/'COUNT.json').read_text());assert summary['complete'] and summary['mixed']==summary['differences']==0
assert json.loads((s/'RECONCILED.json').read_text())['complete']
profile=s/'control-tasks.jsonl';assert not profile.exists();env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-source',STAGE3_TASK_PROFILE=str(profile),NODE_DISABLE_COMPILE_CACHE='1',NODE_OPTIONS='--require='+str(repo/'stage3/oracle/profile-tasks.cjs'));begin=time.monotonic()
with (s/'control.log').open('w') as log:
 p=subprocess.Popen(['bash',str(repo/'stage3/lane/run.sh'),str(out)],cwd=repo,env=env,stdout=log,stderr=log,start_new_session=True)
 try:code=p.wait(timeout=1800)
 finally:
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
  p.wait()
result={'exit':code,'seconds':time.monotonic()-begin,'counted':False};(s/'CONTROL.json').write_text(json.dumps(result,indent=2)+'\n');assert code==0,result
sys.path.insert(0,str(repo/'stage3/lane'));from shard_plan import load_plan,names_digest
plan=load_plan();assert [u['name'] for u in plan['units']]==summary['units'];records=[];counts={'pass':0,'fail':0,'skip':0};diff=b''
for u in plan['units']:
 d=s/'results'/u['name'];report=json.loads((d/'report.json').read_text());observed=json.loads((d/'observations.json').read_text());assert report['status']=='pass' and report['counts']==u['expect'];assert names_digest(observed)==u['names_digest'];records+=observed
 for k,n in report['counts'].items():counts[k]+=n
 diff+=(d/'oracle/baseline.diff').read_bytes()
whole=[dict(rec,task=row['task'],status=status) for row in map(json.loads,profile.read_text().splitlines()) for status,key in [('pass','passes'),('fail','errors')] for rec in row[key]]
assert names_digest(records)==names_digest(whole);assert counts==plan['counts'];assert diff==(out/'oracle/baseline.diff').read_bytes();result.update(complete=True,unitUnionIdentical=True,units=len(plan['units']),tests=len(records),counts=counts,namesDigest=names_digest(records),baselineDiffByteIdentical=True);(s/'CONTROL.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
