import hashlib,json,os,signal,subprocess,tarfile,time,shutil
from concurrent.futures import ThreadPoolExecutor,as_completed
from collections import Counter
from pathlib import Path
s=Path('/workspace/scratch/elementat-declaration');repo=Path('/workspace/adamic');out=repo/'stage3/census/elementat-declaration'
while not (s/'lane-ready.json').exists():time.sleep(1)
plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());units=plan['units'];results=s/'lane-results';results.mkdir(exist_ok=True)
summary={'namespaceCoverage':True,'inventory':len(units),'completedUnits':[],'pieces':[],'calls':0,'differences':0,'reads':0,'tdz':0,'readKinds':{},'tsObjectReads':0,'exportReads':0,'exportUninitialized':0,'exportTdz':0,'exportSites':{},'counts':{'pass':0,'fail':0,'skip':0},'complete':False}
def snapshot():return {str(p.relative_to(s/'counted')):hashlib.sha256(p.read_bytes()).hexdigest() for base in ['src','built/local'] for p in sorted((s/'counted'/base).rglob('*')) if p.is_file()}
frozen=snapshot();(s/'counted-manifest.json').write_text(json.dumps(frozen,indent=2)+'\n')
def run(unit,d):
 name=unit['name'];env=dict(os.environ,STAGE3_ADAPTED_TREE=str(s/'counted'),EA_CALL_DIR=str(d/'calls'),EA_SHARD=name,STAGE3_SHARD_SECONDS='90',NODE_OPTIONS='--require='+str(out/'preload.cjs'),NODE_DISABLE_COMPILE_CACHE='1')
 with (d/(name+'.log')).open('w') as log:
  p=subprocess.Popen(['python3',str(repo/'stage3/lane/shard.py'),name,str(results)],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  timed=False
  try:p.wait(timeout=120)
  except subprocess.TimeoutExpired:timed=True
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 report=json.loads((results/name/'report.json').read_text()) if (results/name/'report.json').exists() else {}
 return {'name':name,'exit':p.returncode,'timedOut':timed,'report':report}
for start in range(0,len(units),20):
 begun=time.monotonic();num=start//20+1;d=s/f'piece-{num:02d}';d.mkdir();(d/'calls').mkdir();subset=units[start:start+20]
 with ThreadPoolExecutor(max_workers=4) as pool:rows=[f.result() for f in as_completed([pool.submit(run,u,d) for u in subset])]
 record={'piece':num,'seconds':time.monotonic()-begun,'units':[u['name'] for u in subset],'children':rows,'calls':0,'differences':0,'reads':0,'tdz':0,'readKinds':{},'tsObjectReads':0,'exportReads':0,'exportUninitialized':0,'exportTdz':0,'exportSites':{},'counts':{'pass':0,'fail':0,'skip':0}}
 kinds=Counter();exportsites=Counter()
 for file in sorted((d/'calls').glob('*.jsonl')):
  for line in file.open():
   r=json.loads(line)
   if file.name.endswith('.exports.jsonl'):
    record['exportReads']+=1;record['exportUninitialized']+=not r['initialized'];record['exportTdz']+=r['tdz'];exportsites[r['site']]+=1
   elif file.name.endswith('.reads.jsonl'):
    record['reads']+=1;record['tdz']+=r['tdz'];kinds[r['kind']]+=1;record['tsObjectReads']+=r['tsObject']
   else:record['calls']+=1;record['differences']+=not r['equal']
 record['readKinds']=dict(kinds);record['exportSites']=dict(exportsites)
 for row in rows:
  name=row['name'];p=results/name
  if p.exists():
   shutil.copytree(p,d/name,ignore=shutil.ignore_patterns('adapted-tree'))
  if row['exit']==0 and row['report'].get('status')=='pass':
   summary['completedUnits'].append(name)
   for k,v in row['report']['counts'].items():record['counts'][k]+=v
 record['sourceFrozen']=frozen==snapshot()
 (d/'piece.json').write_text(json.dumps(record,indent=2)+'\n')
 for k in ['calls','differences','reads','tdz','tsObjectReads','exportReads','exportUninitialized','exportTdz']:summary[k]+=record[k]
 for k,v in record['counts'].items():summary['counts'][k]+=v
 for k,v in kinds.items():summary['readKinds'][k]=summary['readKinds'].get(k,0)+v
 for k,v in exportsites.items():summary['exportSites'][k]=summary['exportSites'].get(k,0)+v
 summary['pieces'].append({k:v for k,v in record.items() if k!='children'});summary['complete']=len(summary['completedUnits'])==len(units)
 (out/'lane-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
 files=sorted(p for p in d.rglob('*') if p.is_file());(d/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(d))+'\n' for p in files))
 archive=out/f'lane-{num:02d}.tar.gz'
 with tarfile.open(archive,'w:gz') as tar:tar.add(d,arcname=d.name)
 (out/f'lane-{num:02d}.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
 print(json.dumps({'piece':num,'completed':len(summary['completedUnits']),'calls':summary['calls'],'differences':summary['differences'],'seconds':record['seconds']}),flush=True)
 for row in rows:shutil.rmtree(results/row['name']/'adapted-tree',ignore_errors=True)
 assert record['sourceFrozen']
 assert len(summary['completedUnits'])==start+len(subset),record
 assert record['differences']==0 and record['tdz']==0,record
assert summary['counts']==plan['counts'],summary
assert summary['calls']<=summary['reads']+summary['exportReads'],summary
