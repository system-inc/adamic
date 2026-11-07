import hashlib,json,os,shutil,signal,subprocess,tarfile,time,fcntl
from collections import Counter
from concurrent.futures import ThreadPoolExecutor,as_completed
from pathlib import Path
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/gov-a');out=repo/'stage3/adapt/30-indexed-reads';plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());units=plan['units'];results=s/'count-results';results.mkdir(exist_ok=True)
summary={'ruling':{'task':'b5w3ycg','at':'2026-11-07 00:12Z','road':'a, measured isArray split and array filter'},'base':json.loads((s/'raw/base.json').read_text()),'unitInventory':len(units),'completedUnits':[],'pieces':[],'calls':0,'kinds':{'array':0,'plain-record':0,'other':0},'arrays':{'sparse':0,'withNonIndexProperties':0,'withPresentUndefined':0,'presentUndefinedElements':0,'accessorIndices':0,'lengths':{},'holes':{}},'oracleCounts':{'pass':0,'fail':0,'skip':0},'countScope':'one complete union of the unchanged 398 lane manifest shards; seeds excluded','complete':False}
if (out/'GETOWNVALUES-COUNT.json').exists():summary=json.loads((out/'GETOWNVALUES-COUNT.json').read_text())
def snapshot():return {str(p.relative_to(s/'counted')):hashlib.sha256(p.read_bytes()).hexdigest() for base in ['src','built/local'] for p in sorted((s/'counted'/base).rglob('*')) if p.is_file()}
frozen=snapshot();(s/'raw/counted-manifest.json').write_text(json.dumps(frozen,indent=2)+'\n')
def git(*args):
 with (s/'git-count.log').open('a') as log:subprocess.run(['git',*args],cwd=repo,stdout=log,stderr=log,check=True)
def run(unit,directory):
 name=unit['name'];env=dict(os.environ,STAGE3_ADAPTED_TREE=str(s/'counted'),STAGE3_RESULTS=str(results),GOV_CALL_DIR=str(directory/'calls'),GOV_SHARD=name,STAGE3_SHARD_SECONDS='90',NODE_OPTIONS='--require='+str(s/'count-preload.cjs'),NODE_DISABLE_COMPILE_CACHE='1')
 with (directory/(name+'.log')).open('w') as log:
  p=subprocess.Popen(['python3',str(repo/'stage3/lane/shard.py'),name,str(results)],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=p.wait(timeout=110)
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 report=json.loads((results/name/'report.json').read_text());return unit,code,report
for start in range(0,len(units),20):
 if start<len(summary['completedUnits']):continue
 begun=time.monotonic();piece=start//20+1;directory=s/f'piece-{piece:02d}';directory.mkdir(exist_ok=False);(directory/'calls').mkdir();subset=units[start:start+20]
 with ThreadPoolExecutor(max_workers=4) as pool:observations=[f.result() for f in as_completed([pool.submit(run,u,directory) for u in subset])]
 assert time.monotonic()-begun<1400
 assert frozen==snapshot(),'counted inputs changed'
 failed=[];oracle=Counter()
 packed=directory/'packed';packed.mkdir()
 with (packed/'shard-evidence.jsonl').open('w') as evidence:
  for unit,code,report in sorted(observations,key=lambda row:row[0]['name']):
   name=unit['name']
   if code!=0 or report['status']!='pass':failed.append({'unit':name,'exit':code,'report':report})
   oracle.update(report['counts'])
   for p in sorted((results/name).rglob('*')):
    if p.is_file() and 'adapted-tree' not in p.relative_to(results/name).parts:
     evidence.write(json.dumps({'unit':name,'path':str(p.relative_to(results/name)),'text':p.read_text(errors='backslashreplace')})+'\n')
   evidence.write(json.dumps({'unit':name,'path':'shard-driver.log','text':(directory/(name+'.log')).read_text()})+'\n')
 calls=[]
 for p in sorted((directory/'calls').glob('*.jsonl')):
  for line in p.open():
   row=json.loads(line);assert row['shard'] in [u['name'] for u in subset];calls.append(row)
 with (packed/'calls.jsonl').open('w') as f:
  for row in calls:f.write(json.dumps(row)+'\n')
 kinds=Counter(row['kind'] for row in calls);arrays=[row for row in calls if row['kind']=='array'];lengths=Counter(str(row['length']) for row in arrays);holes=Counter(str(row['holes']) for row in arrays)
 stats={'sparse':sum(row['sparse'] for row in arrays),'withNonIndexProperties':sum(bool(row['nonIndexProperties']) for row in arrays),'withPresentUndefined':sum(row['presentUndefined']>0 for row in arrays),'presentUndefinedElements':sum(row['presentUndefined'] for row in arrays),'accessorIndices':sum(row['accessorIndices'] for row in arrays),'lengths':dict(lengths),'holes':dict(holes)}
 record={'piece':piece,'seconds':time.monotonic()-begun,'units':[u['name'] for u in subset],'calls':len(calls),'kinds':dict(kinds),'arrays':stats,'oracleCounts':dict(oracle),'failed':failed,'unitDeadlineSeconds':90,'sourceManifestUnchanged':True}
 (packed/'piece.json').write_text(json.dumps(record,indent=2)+'\n')
 if piece==1:
  # Compact preparation into a single physical evidence file.
  prep={p.name:p.read_text() for p in (s/'raw').iterdir() if p.is_file() and p.suffix in ['.json','.log'] and p.name not in ['backend-witnesses.json']}
  (packed/'preparation.json').write_text(json.dumps(prep,indent=2)+'\n')
 if not failed:
  summary['completedUnits'] += [u['name'] for u in subset];summary['calls']+=len(calls)
  for kind,n in kinds.items():summary['kinds'][kind]+=n
  for key,value in stats.items():
   if isinstance(value,dict):
    for k,n in value.items():summary['arrays'][key][k]=summary['arrays'][key].get(k,0)+n
   else:summary['arrays'][key]+=value
  for k,n in oracle.items():summary['oracleCounts'][k]+=n
 summary['pieces'].append(record);summary['complete']=len(summary['completedUnits'])==len(units)
 (out/'GETOWNVALUES-COUNT.json').write_text(json.dumps(summary,indent=2)+'\n')
 (packed/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name+'\n' for p in sorted(packed.iterdir()) if p.name!='SHA256SUMS'))
 archive=out/f'getownvalues-count-{piece:02d}.tar.gz'
 with tarfile.open(archive,'w:gz') as tf:
  for p in sorted(packed.iterdir()):tf.add(p,arcname=f'piece-{piece:02d}/'+p.name)
 (out/f'getownvalues-count-{piece:02d}.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
 with (s/'git.lock').open('a') as lock:
  fcntl.flock(lock,fcntl.LOCK_EX);git('add',str(out.relative_to(repo)));git('diff','--cached','--check');git('commit','-m',f'never cut alone: lands with chg_yj14ddxz, because the census reads its stop\n\ngetOwnValues Step A: counted lane piece {piece}, {len(subset)} shards, {len(calls)} calls.');git('push','origin','typescript/getownvalues')
 print(json.dumps({'piece':piece,'seconds':record['seconds'],'completed':len(summary['completedUnits']),'calls':summary['calls'],'kinds':summary['kinds'],'sparse':summary['arrays']['sparse'],'failedUnits':len(failed)}),flush=True)
 for unit,_,_ in observations:shutil.rmtree(results/unit['name']/'adapted-tree',ignore_errors=True)
 if failed:raise SystemExit('Incomplete piece pushed; inspect failure before continuing.')
assert summary['oracleCounts']==plan['counts'],summary['oracleCounts']
assert len(summary['completedUnits'])==len(set(summary['completedUnits']))==len(units)
print('COMPLETE',json.dumps({'calls':summary['calls'],'kinds':summary['kinds'],'arrays':summary['arrays'],'oracleCounts':summary['oracleCounts']}),flush=True)
