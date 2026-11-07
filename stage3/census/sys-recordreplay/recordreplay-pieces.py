from pathlib import Path
import collections,concurrent.futures,hashlib,json,os,shutil,signal,subprocess,sys,tarfile,time
sys.path.insert(0,'/workspace/scratch/sys-walls');from run import run,S,ROOT
out=Path(sys.argv[1]); reportrepo=out.parents[2]; tree=S/'recordreplay';results=S/'recordreplay-results';results.mkdir(exist_ok=True)
plan=json.loads((ROOT/'stage3/lane/shard-plan.json').read_text());units=plan['units']
summary={'base':subprocess.check_output(['git','rev-parse','origin/main'],cwd=ROOT,text=True).strip(),'units':len(units),'completedUnits':[],'pieces':[],'evaluations':0,'truthy':0,'falsy':0,'actualValues':{},'oracleCounts':{'pass':0,'fail':0,'skip':0},'complete':False,'measurementOnly':True}
if (out/'COUNT.json').exists():summary=json.loads((out/'COUNT.json').read_text())
def digesttree():return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for base in ['src','built/local'] for p in sorted((tree/base).rglob('*')) if p.is_file()}
frozen=digesttree();(S/'recordreplay-frozen.json').write_text(json.dumps(frozen,indent=2)+'\n')
def unitrun(unit,directory):
 name=unit['name'];calls=directory/'calls'/name;calls.mkdir(parents=True);log=directory/(name+'.log');env=dict(os.environ,STAGE3_ADAPTED_TREE=str(tree),STAGE3_RESULTS=str(results),STAGE3_SHARD_SECONDS='90',SYS_RECORDREPLAY_DIR=str(calls),NODE_OPTIONS='--require='+str(S/'recordreplay-preload.cjs'),NODE_DISABLE_COMPILE_CACHE='1',PYTHONDONTWRITEBYTECODE='1')
 start=time.monotonic();reason=None
 with log.open('w') as stream:
  p=subprocess.Popen(['python3',str(ROOT/'stage3/lane/shard.py'),name,str(results)],cwd=ROOT,env=env,stdout=stream,stderr=subprocess.STDOUT,start_new_session=True)
  try:
   try:p.wait(timeout=300)
   except subprocess.TimeoutExpired:
    reason='unit exceeded 300 seconds';os.killpg(p.pid,signal.SIGTERM)
    try:p.wait(timeout=10)
    except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL)
   code=p.wait()
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 reportpath=results/name/'report.json';report=json.loads(reportpath.read_text()) if reportpath.exists() else None
 return dict(unit=name,child_exit=code,seconds=time.monotonic()-start,stop_reason=reason,report=report)
for start in range(0,len(units),20):
 if start<len(summary['completedUnits']):continue
 begin=time.monotonic();piece=start//20+1;directory=S/f'recordreplay-piece-{piece:02d}';directory.mkdir();subset=units[start:start+20]
 with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:observations=list(pool.map(lambda u:unitrun(u,directory),subset))
 assert time.monotonic()-begin<1500
 stats=collections.Counter();values=collections.Counter();oracle=collections.Counter();failures=[];events=[]
 for row in observations:
  name=row['unit'];report=row['report']
  if row['child_exit']!=0 or report is None or report['status']!='pass':failures.append(row)
  if report:oracle.update(report['counts'])
  paths=sorted((directory/'calls'/name).glob('*.jsonl'));ready=0
  for p in paths:
   rows=[json.loads(l) for l in p.read_text().splitlines()];events.append(dict(unit=name,file=p.name,rows=rows))
   ready+=sum(x['event']=='observer-ready' for x in rows)
   for x in rows:
    if x['event']=='evaluation':
     stats['evaluations']+=1;stats['truthy' if x['truthy'] else 'falsy']+=1;values[json.dumps(x['actual'],sort_keys=True)]+=1
  if not ready:failures.append({'unit':name,'error':'no observer-ready witness'})
  local=results/name/'adapted-tree'
  if local.exists():shutil.rmtree(local)
 unchanged=frozen==digesttree()
 if not unchanged:failures.append({'error':'measurement input changed'})
 record=dict(piece=piece,units=[u['name'] for u in subset],seconds=time.monotonic()-begin,stats=dict(stats),actualValues=dict(values),oracleCounts=dict(oracle),failed=failures,frozenInputsUnchanged=unchanged)
 (directory/'piece.json').write_text(json.dumps(record,indent=2)+'\n');(directory/'children.json').write_text(json.dumps(observations,indent=2)+'\n')
 (directory/'events.json').write_text(json.dumps(events,indent=2)+'\n')
 archive=out/f'piece-{piece:02d}.tar.gz'
 with tarfile.open(archive,'w:gz') as tf:
  tf.add(directory,arcname=f'piece-{piece:02d}')
  for u in subset:
   if (results/u['name']).exists():tf.add(results/u['name'],arcname='results/'+u['name'])
 (out/(archive.name+'.sha256')).write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
 summary['pieces'].append(record)
 if not failures:
  summary['completedUnits']+=record['units']
  for key in ['evaluations','truthy','falsy']:summary[key]+=stats[key]
  for k,v in values.items():summary['actualValues'][k]=summary['actualValues'].get(k,0)+v
  for k,v in oracle.items():summary['oracleCounts'][k]+=v
 summary['complete']=len(summary['completedUnits'])==len(units)
 (out/'COUNT.json').write_text(json.dumps(summary,indent=2)+'\n')
 (out/'STATUS.md').write_text(f"Completed {len(summary['completedUnits'])}/{len(units)} lane units. Evaluations {summary['evaluations']}, truthy {summary['truthy']}, falsy {summary['falsy']}. Complete: {summary['complete']}. Piece child exits and ordinary lane checks are archived.\n")
 for name,args in [('add',['git','add',str(out.relative_to(reportrepo))]),('check',['git','diff','--cached','--check']),('commit',['git','commit','-m',f'Measure sys recordreplay lane piece {piece}']),('push',['git','push','origin','HEAD:typescript/sys-recordreplay'])]:
  code=run(f'recordreplay-piece-{piece:02d}-{name}',args,cwd=reportrepo)
  if code:raise SystemExit(code)
 print(json.dumps({'piece':piece,'completed':len(summary['completedUnits']),'evaluations':summary['evaluations'],'truthy':summary['truthy'],'falsy':summary['falsy'],'failures':len(failures),'seconds':record['seconds']}),flush=True)
 if failures:raise SystemExit('Stopped after failed measurement piece; checkpoint pushed')
assert len(set(summary['completedUnits']))==len(units)
assert summary['oracleCounts']==plan['counts'],(summary['oracleCounts'],plan['counts'])
assert summary['evaluations']==summary['truthy']+summary['falsy']
print('COMPLETE',json.dumps({k:v for k,v in summary.items() if k not in ['pieces','completedUnits']}),flush=True)
