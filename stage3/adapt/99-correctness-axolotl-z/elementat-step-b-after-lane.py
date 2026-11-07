import hashlib,json,os,shutil,signal,subprocess,tarfile,time
from concurrent.futures import ThreadPoolExecutor,as_completed
from pathlib import Path
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/elementat-b/step-b');out=repo/'stage3/adapt/99-correctness-axolotl-z';plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());results=s/'lane-results';results.mkdir(exist_ok=True)
manifest={str(p.relative_to(s/'after')):hashlib.sha256(p.read_bytes()).hexdigest() for base in ['src','built/local'] for p in sorted((s/'after'/base).rglob('*')) if p.is_file()};(s/'raw/after-lane-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
progress={'units':398,'completedUnits':[],'counts':{'pass':0,'fail':0,'skip':0},'pieces':[],'complete':False};pending=[]
def git(*args):
 with (s/'after-lane-git.log').open('a') as log:subprocess.run(['git',*args],cwd=repo,stdout=log,stderr=log,check=True)
def publish():
 # The native census is the other checkpoint writer. Its marker is written only
 # after its final plain push exits, so complete archives are staged serially.
 if not (s/'census-finished.json').exists():return
 while pending:
  number,archive,record,snapshot=pending.pop(0);target=out/archive.name;shutil.copyfile(archive,target);(out/(archive.stem.replace('.tar','')+'.sha256')).write_text(hashlib.sha256(target.read_bytes()).hexdigest()+'  '+target.name+'\n')
  (out/'elementat-after-lane.json').write_text(json.dumps(snapshot,indent=2)+'\n');git('add',str(out.relative_to(repo)));git('commit','-m',f'never cut alone: lands with chg_pyyr04zd, because the census reads its stop\n\n99z: fallback whole-lane piece {number}, {len(record["units"])} shards; exact identities and baselines pass');git('push','origin','typescript/elementat-any');print('pushed after-lane piece',number,flush=True)
def run(unit,directory):
 env=dict(os.environ,STAGE3_ADAPTED_TREE=str(s/'after'),STAGE3_RESULTS=str(results),STAGE3_SHARD_SECONDS='90',NODE_OPTIONS='',NODE_DISABLE_COMPILE_CACHE='1');env.pop('EA_CALL_DIR',None)
 with (directory/(unit['name']+'.log')).open('w') as log:
  child=subprocess.Popen(['python3',str(repo/'stage3/lane/shard.py'),unit['name'],str(results)],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=child.wait(timeout=100)
  finally:
   try:os.killpg(child.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   child.wait()
 report=json.loads((results/unit['name']/'report.json').read_text());return unit,code,report
for start in range(0,398,20):
 begun=time.monotonic();number=start//20+1;directory=s/f'after-lane-{number:02d}';directory.mkdir();subset=plan['units'][start:start+20];workers=4 if (s/'census-finished.json').exists() else 2
 with ThreadPoolExecutor(max_workers=workers) as pool:observations=[f.result() for f in as_completed([pool.submit(run,u,directory) for u in subset])]
 failed=[]
 for unit,code,report in observations:
  if code or report['status']!='pass':failed.append(report)
  shutil.copytree(results/unit['name'],directory/unit['name'],ignore=shutil.ignore_patterns('adapted-tree'))
 assert manifest=={str(p.relative_to(s/'after')):hashlib.sha256(p.read_bytes()).hexdigest() for base in ['src','built/local'] for p in sorted((s/'after'/base).rglob('*')) if p.is_file()}
 record={'piece':number,'units':[u['name'] for u in subset],'seconds':time.monotonic()-begun,'workers':workers,'deadlineSeconds':90,'failed':failed};assert record['seconds']<1400
 (directory/'piece.json').write_text(json.dumps(record,indent=2)+'\n');(directory/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(directory))+'\n' for p in sorted(directory.rglob('*')) if p.is_file()))
 archive=s/f'elementat-after-lane-{number:02d}.tar.gz'
 with tarfile.open(archive,'w:gz') as tar:tar.add(directory,arcname=f'piece-{number:02d}')
 if not failed:
  progress['completedUnits'] += record['units']
  for _,_,r in observations:
   for key,n in r['counts'].items():progress['counts'][key]+=n
 progress['pieces'].append(record);progress['complete']=len(progress['completedUnits'])==398
 pending.append((number,archive,record,json.loads(json.dumps(progress))));(s/'raw/after-lane-progress.json').write_text(json.dumps(progress,indent=2)+'\n');publish()
 print(json.dumps({'piece':number,'seconds':record['seconds'],'completed':len(progress['completedUnits']),'counts':progress['counts'],'failed':len(failed),'checkpointQueued':bool(pending)}),flush=True)
 for u,_,_ in observations:shutil.rmtree(results/u['name']/'adapted-tree',ignore_errors=True)
 if failed:raise SystemExit('after lane failure; raw evidence retained')
assert progress['counts']==plan['counts']
while pending:
 publish()
 if pending:time.sleep(5)
print('COMPLETE fallback lane',flush=True)
