import hashlib,json,os,pathlib,shutil,signal,subprocess,sys,time
from concurrent.futures import ThreadPoolExecutor,as_completed
repo=pathlib.Path(__file__).resolve().parents[3];tree=pathlib.Path(sys.argv[1]).resolve();out=pathlib.Path(sys.argv[2]).resolve();out.mkdir(parents=True,exist_ok=False)
plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());results=out/'results';results.mkdir()
def manifest():return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for base in ['src','built/local'] for p in sorted((tree/base).rglob('*')) if p.is_file()}
original=manifest();(out/'manifest.json').write_text(json.dumps(original,indent=2)+'\n')
progress={'units':len(plan['units']),'completedUnits':[],'counts':{'pass':0,'fail':0,'skip':0},'pieces':[],'complete':False}
def run(unit,directory):
 env=dict(os.environ,STAGE3_ADAPTED_TREE=str(tree),STAGE3_RESULTS=str(results),STAGE3_SHARD_SECONDS='90',NODE_OPTIONS=('--require='+os.environ['QUEUE_BOX_COUNT_PRELOAD']) if os.environ.get('QUEUE_BOX_COUNT_PRELOAD') else '',NODE_DISABLE_COMPILE_CACHE='1')
 with (directory/(unit['name']+'.log')).open('w') as log:
  child=subprocess.Popen(['python3',str(repo/'stage3/lane/shard.py'),unit['name'],str(results)],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:
   code=child.wait(timeout=100)
  except subprocess.TimeoutExpired:code=124
  finally:
   try:os.killpg(child.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   child.wait()
 report=json.loads((results/unit['name']/'report.json').read_text());return unit,code,report
for start in range(0,len(plan['units']),20):
 begun=time.monotonic();number=start//20+1;directory=out/f'piece-{number:02d}';directory.mkdir();subset=plan['units'][start:start+20]
 with ThreadPoolExecutor(max_workers=4) as pool:observations=[f.result() for f in as_completed([pool.submit(run,u,directory) for u in subset])]
 failed=[]
 for unit,code,report in observations:
  if code or report['status']!='pass':failed.append({'unit':unit['name'],'exit':code,'report':report})
  shutil.copytree(results/unit['name'],directory/unit['name'],ignore=shutil.ignore_patterns('adapted-tree'))
 assert original==manifest(),'measured source or emitted JavaScript changed'
 record={'piece':number,'units':[u['name'] for u in subset],'seconds':time.monotonic()-begun,'workers':4,'deadlineSeconds':90,'failed':failed};assert record['seconds']<1400
 (directory/'piece.json').write_text(json.dumps(record,indent=2)+'\n')
 if not failed:
  progress['completedUnits']+=record['units']
  for _,_,report in observations:
   for key,n in report['counts'].items():progress['counts'][key]+=n
 progress['pieces'].append(record);progress['complete']=len(progress['completedUnits'])==len(plan['units']);(out/'progress.json').write_text(json.dumps(progress,indent=2)+'\n')
 print(json.dumps({'piece':number,'seconds':record['seconds'],'completed':len(progress['completedUnits']),'counts':progress['counts'],'failed':len(failed)}),flush=True)
 for unit,_,_ in observations:shutil.rmtree(results/unit['name']/'adapted-tree',ignore_errors=True)
 if os.environ.get('QUEUE_BOX_COUNT_DIR'):
  countRows=[json.loads(p.read_text()) for p in pathlib.Path(os.environ['QUEUE_BOX_COUNT_DIR']).glob('*.json')]
  assert not any(r['counts'].get('counterexamples',0) for r in countRows),'Real queue counterexample; stop before construction'
 if failed:raise SystemExit('Lane piece failed; evidence retained')
assert progress['counts']==plan['counts'];print('COMPLETE boxed queue lane',flush=True)
