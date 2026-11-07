import hashlib,json,os,shutil,signal,subprocess,tarfile,time
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor,as_completed
from collections import Counter
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/levenshtein');plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());results=s/'count-results';results.mkdir(exist_ok=True)
progress={'complete':False,'units':len(plan['units']),'completedUnits':[],'calls':0,'differences':0,'classes':{},'counts':{'pass':0,'fail':0,'skip':0},'pieces':[]}
def manifest():
 return {str(p.relative_to(s/'counted')):hashlib.sha256(p.read_bytes()).hexdigest() for base in ['src','built/local'] for p in sorted((s/'counted'/base).rglob('*')) if p.is_file()}
frozen=manifest();(s/'counted-manifest.json').write_text(json.dumps(frozen,indent=2)+'\n')
classes=['equalStrings','emptyS1','emptyS2','caseOnly','giveUp','maxBelowOne']
def run(unit,directory):
 name=unit['name'];env=dict(os.environ,STAGE3_ADAPTED_TREE=str(s/'counted'),STAGE3_SHARD_SECONDS='90',NODE_OPTIONS='--require='+str(s/'preload.cjs'),NODE_DISABLE_COMPILE_CACHE='1',LV_CALL_DIR=str(directory/'calls'),LV_UNIT=name)
 with (directory/(name+'.log')).open('w') as log:
  p=subprocess.Popen(['python3',str(repo/'stage3/lane/shard.py'),name,str(results)],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=p.wait(timeout=110)
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 report=json.loads((results/name/'report.json').read_text());return name,code,report
for start in range(0,len(plan['units']),20):
 begin=time.monotonic();piece=start//20+1;directory=s/f'count-{piece:02d}';directory.mkdir();(directory/'calls').mkdir();subset=plan['units'][start:start+20]
 with ThreadPoolExecutor(max_workers=4 if (s/"control.json").exists() else 2) as pool:rows=[f.result() for f in as_completed([pool.submit(run,u,directory) for u in subset])]
 assert frozen==manifest(),'counted inputs changed'
 failures=[]
 for name,code,report in rows:
  if code or report['status']!='pass':failures.append({'unit':name,'exit':code,'report':report})
  shutil.copytree(results/name,directory/name,ignore=shutil.ignore_patterns('adapted-tree'))
 counts=Counter();calls=0;diffs=0
 for file in (directory/'calls').glob('*.jsonl'):
  assert file.name.rsplit('.',2)[0] in [u['name'] for u in subset]
  for line in file.open():
   r=json.loads(line);calls+=1;diffs+=not r['equal']
   for c in classes:counts[c]+=r[c]
 record={'piece':piece,'units':[u['name'] for u in subset],'seconds':time.monotonic()-begin,'calls':calls,'differences':diffs,'classes':dict(counts),'failed':failures}
 (directory/'piece.json').write_text(json.dumps(record,indent=2)+'\n')
 progress['pieces'].append(record);progress['calls']+=calls;progress['differences']+=diffs
 for c in classes:progress['classes'][c]=progress['classes'].get(c,0)+counts[c]
 if not failures:
  progress['completedUnits']+=record['units']
  for _,_,r in rows:
   for k,n in r['counts'].items():progress['counts'][k]+=n
 progress['complete']=len(progress['completedUnits'])==len(plan['units'])
 (s/'COUNT.json').write_text(json.dumps(progress,indent=2)+'\n')
 (directory/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(directory))+'\n' for p in sorted(directory.rglob('*')) if p.is_file()))
 with tarfile.open(s/f'count-{piece:02d}.tar.gz','w:gz') as tar:tar.add(directory,arcname=directory.name)
 print(json.dumps(record),flush=True)
 for name,_,_ in rows:shutil.rmtree(results/name/'adapted-tree',ignore_errors=True)
 if failures or diffs:raise SystemExit('STOP: oracle failure or counterexample')
 assert record['seconds']<1500
assert progress['counts']==plan['counts']
assert len(set(progress['completedUnits']))==len(plan['units'])
print('COMPLETE',progress['calls'],progress['differences'],flush=True)
