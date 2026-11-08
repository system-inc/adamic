"""Run the whole manifest in bounded pieces, preserving every attempt."""
import concurrent.futures,json,os,shutil,signal,subprocess,sys,time
from pathlib import Path
repo=Path('/workspace/adamic');lane=repo/'stage3/lane';tree=Path(sys.argv[1]);root=Path(sys.argv[2]);root.mkdir(parents=True,exist_ok=True)
manifest=(lane/'shards.json').read_bytes();snapshot=root/'shards.json'
if snapshot.exists():assert snapshot.read_bytes()==manifest
else:snapshot.write_bytes(manifest)
units=json.loads(manifest)['units'];started=time.monotonic();history=[]
prior_seconds=json.loads((root/'ledger.json').read_text()).get('wall_seconds',0) if (root/'ledger.json').exists() else 0
ledger=root/'ledger.jsonl'
if ledger.exists():history=[json.loads(line) for line in ledger.read_text().splitlines()]
for row in history:row.setdefault('attempt','four-way-initial')
latest={row['name']:row for row in history};passed={name:row for name,row in latest.items() if row['exit']==0 and row['status']=='pass'}
def archive(name,label):
 target=root/'attempts'/label;target.mkdir(parents=True,exist_ok=True)
 for source in [root/name,root/(name+'.log')]:
  if source.exists():shutil.move(str(source),str(target/source.name))
def run(unit,label):
 name=unit['name'];archive(name,label)
 env=dict(os.environ,STAGE3_RESULTS=str(root),STAGE3_CACHE=str(root/'caches'/name),STAGE3_ADAPTED_TREE=str(tree),NODE_DISABLE_COMPILE_CACHE='1',PYTHONDONTWRITEBYTECODE='1')
 begin=time.monotonic()
 with (root/(name+'.log')).open('w') as log:
  child=subprocess.Popen(unit['command'],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=child.wait(timeout=1800)
  except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);code=child.wait()
  finally:
   try:os.killpg(child.pid,signal.SIGKILL)
   except ProcessLookupError:pass
 output=root/name;report=json.loads((output/'report.json').read_text()) if (output/'report.json').is_file() else {}
 row=dict(name=name,command=unit['command'],exit=code,seconds=time.monotonic()-begin,status=report.get('status','missing'),counts=report.get('counts'),errors=report.get('errors'),attempt=label)
 if name=='stage3-lane-byte-comparisons' and (output/'adapted-tree').exists():
  saved=output/'saved-api';saved.mkdir()
  for source,target in [('tests/baselines/reference/api/typescript.d.ts','reference.d.ts'),('tests/baselines/local/api/typescript.d.ts','local.d.ts'),('built/local/typescript.d.ts','built.d.ts')]:
   p=output/'adapted-tree'/source
   if p.is_file():shutil.copyfile(p,saved/target)
 if (output/'adapted-tree').exists():shutil.rmtree(output/'adapted-tree')
 return row
def retain(row):
 history.append(row)
 with ledger.open('a') as stream:stream.write(json.dumps(row)+'\n')
 latest[row['name']]=row;print(json.dumps(row),flush=True)
pending=[u for u in units if u['name'] not in passed][:100];label='two-way-'+str(len(history));failed=[]
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 for job in concurrent.futures.as_completed([pool.submit(run,u,label) for u in pending]):
  row=job.result();retain(row)
  if row['exit'] or row['status']!='pass':failed.append(row)
for row in failed:
 assert row['errors'] and all('timed out' in error for error in row['errors']),row
 unit=next(u for u in units if u['name']==row['name']);retry=run(unit,'standalone-'+str(len(history)));retain(retry)
 assert retry['exit']==0 and retry['status']=='pass',retry
assert (lane/'shards.json').read_bytes()==manifest
final=[latest[u['name']] for u in units if u['name'] in latest];report=dict(concurrency=2,standaloneTimeoutRetries=True,units=final,wall_seconds=prior_seconds+time.monotonic()-started,piece_seconds=time.monotonic()-started,maximum=max(r['seconds'] for r in final),all_commands_executed=len(final)==len(units),attempts=history)
(root/'ledger.json').write_text(json.dumps(report,indent=2)+'\n')
raise SystemExit(int(any(r['exit'] or r['status']!='pass' or r['seconds']>=30 for r in final)))
