import concurrent.futures,json,os,shutil,subprocess,sys,time
from pathlib import Path
repo=Path('/workspace/adamic');lane=repo/'stage3/lane';tree=Path(sys.argv[1]);root=Path(sys.argv[2]);root.mkdir(parents=True,exist_ok=False)
manifest=(lane/'shards.json').read_bytes();(root/'shards.json').write_bytes(manifest);units=json.loads(manifest)['units'];started=time.monotonic();rows=[]
def run(unit):
 env=dict(os.environ,STAGE3_RESULTS=str(root),STAGE3_CACHE=str(root/'caches'/unit['name']),STAGE3_ADAPTED_TREE=str(tree),NODE_DISABLE_COMPILE_CACHE='1',PYTHONDONTWRITEBYTECODE='1')
 begin=time.monotonic()
 with (root/(unit['name']+'.log')).open('w') as log:code=subprocess.run(unit['command'],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT).returncode
 output=root/unit['name'];report=json.loads((output/'report.json').read_text()) if (output/'report.json').is_file() else {}
 row=dict(name=unit['name'],command=unit['command'],exit=code,seconds=time.monotonic()-begin,status=report.get('status','missing'),counts=report.get('counts'),errors=report.get('errors'))
 if unit['name']=='stage3-lane-byte-comparisons' and (output/'adapted-tree').exists():
  saved=output/'saved-api';saved.mkdir()
  for source,target in [('tests/baselines/reference/api/typescript.d.ts','reference.d.ts'),('tests/baselines/local/api/typescript.d.ts','local.d.ts'),('built/local/typescript.d.ts','built.d.ts')]:
   p=output/'adapted-tree'/source
   if p.is_file():shutil.copyfile(p,saved/target)
 if (output/'adapted-tree').exists():shutil.rmtree(output/'adapted-tree')
 return row
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
 for job in concurrent.futures.as_completed([pool.submit(run,u) for u in units]):
  row=job.result();rows.append(row)
  with (root/'ledger.jsonl').open('a') as f:f.write(json.dumps(row)+'\n')
  print(json.dumps(row),flush=True)
assert (lane/'shards.json').read_bytes()==manifest,'manifest changed during measurement'
report=dict(concurrency=4,units=sorted(rows,key=lambda r:r['name']),wall_seconds=time.monotonic()-started,maximum=max(r['seconds'] for r in rows),all_commands_executed=len(rows)==len(units))
(root/'ledger.json').write_text(json.dumps(report,indent=2)+'\n')
raise SystemExit(int(any(r['exit'] or r['status']!='pass' or r['seconds']>=30 for r in rows)))
