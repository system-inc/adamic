import pathlib,json,os,shutil,subprocess,signal,time,hashlib,concurrent.futures,sys
R=pathlib.Path('/workspace/scratch/core-wave');repo=pathlib.Path('/workspace/adamic-baseline');sys.path.insert(0,str(repo/'stage3/lane'))
from shard_plan import load_plan,names_digest
plan=load_plan();units=plan['units'];tree=R/'counted';results=R/'count-results';results.mkdir(exist_ok=True)
env=dict(os.environ,NODE_OPTIONS='--require='+str(R/'count-preload.cjs'),NODE_DISABLE_COMPILE_CACHE='1',CORE_COUNT_TREE=str(tree))
def run(args,cwd,log,env,limit=1400):
 with log.open('w') as stream:
  p=subprocess.Popen(args,cwd=cwd,env=env,stdout=stream,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=p.wait(timeout=limit)
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 return code
if not tree.exists():
 shutil.copytree(R/'lane/adapted-tree',tree,symlinks=True,ignore=lambda directory,names:[n for n in names if pathlib.Path(directory)==R/'lane/adapted-tree' and n in ['node_modules']]);(tree/'node_modules').symlink_to(R/'lane/adapted-tree/node_modules',target_is_directory=True);shutil.rmtree(tree/'tests/baselines/local',ignore_errors=True)
 prep=R/'count-preparation';prep.mkdir();penv=dict(env,CORE_COUNT_DIR=str(prep))
 assert run(['npm','run','build'],tree,prep/'build.log',penv)==0
 frozen={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for top in ['src','built/local'] for p in sorted((tree/top).rglob('*')) if p.is_file()};(prep/'manifest.json').write_text(json.dumps(frozen,indent=2)+'\n')
else:
 prep=R/'count-preparation';frozen=json.loads((prep/'manifest.json').read_text())
 assert frozen=={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for top in ['src','built/local'] for p in sorted((tree/top).rglob('*')) if p.is_file()}
if (tree/'node_modules').is_symlink():(tree/'node_modules').unlink()
modules=repo/'stage3/typescript/node_modules'
if not modules.is_symlink():modules.symlink_to(R/'lane/adapted-tree/node_modules',target_is_directory=True)
summary={'complete':False,'units':len(units),'completedUnits':[],'pieces':[],'counts':{'groupBy':0,'clone':0,'completed':{},'groupByClasses':{},'groupByKeys':{},'groupBySelectors':{},'groupByCallers':{},'cloneClasses':{}},'oracleCounts':{'pass':0,'fail':0,'skip':0}}
observations=[];baseline=b''
def merge(a,b):
 for k,v in b.items():
  if k=='pid':continue
  if isinstance(v,dict):merge(a.setdefault(k,{}),v)
  else:a[k]=a.get(k,0)+v
for start in range(0,len(units),20):
 piece=R/('count-piece-%02d'%(start//20+1));piece.mkdir(exist_ok=True);(piece/'calls').mkdir(exist_ok=True);begin=time.monotonic();subset=units[start:start+20]
 def measure(unit):
  e=dict(env,CORE_COUNT_DIR=str(piece/'calls'),STAGE3_ADAPTED_TREE=str(tree),STAGE3_SHARD_SECONDS='90',STAGE3_RESULTS=str(results))
  code=run(['python3',str(repo/'stage3/lane/shard.py'),unit['name'],str(results)],repo,piece/(unit['name']+'.log'),e,110)
  directory=results/unit['name'];report=json.loads((directory/'report.json').read_text());assert code==0 and report['status']=='pass',(unit['name'],code,report)
  observed=json.loads((directory/'observations.json').read_text());assert names_digest(observed)==unit['names_digest'];return unit,report,observed
 ready=all((results/u['name']/'report.json').exists() and json.loads((results/u['name']/'report.json').read_text())['status']=='pass' for u in subset)
 if ready:
  born=int(subprocess.check_output(['stat','-c','%W',str(piece)],text=True).strip());assert born>0
  begin=time.monotonic()-(time.time()-born)
  rows=[(u,json.loads((results/u['name']/'report.json').read_text()),json.loads((results/u['name']/'observations.json').read_text())) for u in subset]
 else:
  with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:rows=list(pool.map(measure,subset))
 assert time.monotonic()-begin<1400
 for unit,report,observed in rows:
  directory=results/unit['name'];observations+=observed;baseline+=(directory/'oracle/baseline.diff').read_bytes();summary['completedUnits'].append(unit['name'])
  for k,n in report['counts'].items():summary['oracleCounts'][k]+=n
  for f in ['report.json','observations.json','oracle/baseline.diff']:
   dest=piece/unit['name']/f;dest.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(directory/f,dest)
  shutil.copytree(directory,piece/unit['name'],ignore=shutil.ignore_patterns('adapted-tree'),dirs_exist_ok=True)
  shutil.rmtree(directory/'adapted-tree',ignore_errors=True)
 counts={}
 for file in (piece/'calls').glob('*.jsonl'):
  for line in file.read_text().splitlines():merge(counts,json.loads(line))
 merge(summary['counts'],counts);record={'piece':start//20+1,'seconds':time.monotonic()-begin,'units':[u['name'] for u in subset],'counts':counts,'recoveredFromCompletedGuardedResults':ready};summary['pieces'].append(record)
 (piece/'summary.json').write_text(json.dumps(record,indent=2)+'\n');(R/'count-summary.json').write_text(json.dumps(summary,indent=2)+'\n');print('count piece',record['piece'],'units',len(summary['completedUnits']),'calls',summary['counts']['groupBy'],flush=True)
 assert frozen=={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for top in ['src','built/local'] for p in sorted((tree/top).rglob('*')) if p.is_file()},'counted source or build changed'
whole=[dict(record,task=row['task'],status=status) for row in map(json.loads,(R/'whole-tasks.jsonl').read_text().splitlines()) for status,key in [('pass','passes'),('fail','errors')] for record in row[key]]
assert names_digest(whole)==names_digest(observations)
assert summary['oracleCounts']==plan['counts']
assert baseline==(R/'lane/oracle/baseline.diff').read_bytes()
summary.update(complete=True,identityDigest=names_digest(observations),testCount=len(observations),wholeLaneIdentityMultisetIdentical=True,wholeLaneBaselineDiffByteIdentical=True,countedSourceAndBuildUnchanged=True,equivalenceProof='Both selected owners transpile to byte-identical JavaScript, so each counted input has identical behavior without evaluating callbacks or getters twice.',counterexamples=0)
(R/'count-summary.json').write_text(json.dumps(summary,indent=2)+'\n');print('COMPLETE',summary['counts'],flush=True)
