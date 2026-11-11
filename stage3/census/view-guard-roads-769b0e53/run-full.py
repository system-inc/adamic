#!/usr/bin/env python3
"""Resume independent full-mode children, verify each file, checkpoint every child."""
import concurrent.futures,hashlib,json,os,signal,subprocess,sys,threading,runpy
from pathlib import Path
root=Path(__file__).resolve().parent;repo=root.parents[2]
helpers=runpy.run_path(str(root/'pair-guards.py'));raw_findings=helpers['raw_findings'];load=helpers['load']
files=json.loads(Path(sys.argv[1]).read_text());workers=int(sys.argv[2]) if len(sys.argv)>2 else 3
assert all(Path(x).is_file() and x.startswith('/workspace/roads-adapted/src/') for x in files)
lock=threading.Lock();active={};stopping=False;verified=set()
def stop(sig,frame):
 global stopping
 stopping=True
 with lock:
  for proc in active.values():proc.terminate()
signal.signal(signal.SIGTERM,stop);signal.signal(signal.SIGINT,stop)
def checkpoint(label):
 result=subprocess.run(['bash',str(root/'checkpoint.sh'),label],cwd=repo)
 print('checkpoint_wrapper_child_exit='+str(result.returncode),flush=True)
 if result.returncode:raise RuntimeError('checkpoint failed')
def prefix(file):return 'full-'+hashlib.sha256(file.encode()).hexdigest()[:10]
def jobs(file):
 p=prefix(file)
 return [(file,'parent','base',p+'-parent-base'),(file,'parent','guard',p+'-parent-guard'),(file,'head','guard',p+'-head-guard')]
def completed(name):
 p=root/'raw'/(name+'.json')
 return p.exists() and json.loads(p.read_text())['child_exit']==0

def one(job):
 file,side,mode,name=job;p=prefix(file)
 if stopping:return job,False
 if completed(name):return job,False
 env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=file,GOMAXPROCS='2',GOMEMLIMIT='3GiB',ADAMIC_TYPESCRIPT_SOURCE='/workspace/roads-typescript',LATENT_TIMES=str(root/'raw'/(name+'-times.jsonl')))
 env.pop('ROADS_GUARD_MUTANT',None);env.pop('ROADS_GUARD_STREAM',None)
 binary=f'/workspace/roads-{side}-census'
 if mode=='guard':
  binary=f'/workspace/roads-{side}-guard-census';env['ROADS_GUARD_STREAM']=str(root/'raw'/(p+'-'+side+'-records.jsonl'));Path(env['ROADS_GUARD_STREAM']).unlink(missing_ok=True)
 proc=subprocess.Popen([sys.executable,str(root/'run.py'),str(root/'raw'/name),binary,'/workspace/roads-adapted/src/tsc/tsc.ts',str(root/'raw'/(name+'-normal.jsonl'))],cwd=repo,env=env)
 with lock:
  active[name]=proc
  if stopping:proc.terminate()
 code=proc.wait()
 with lock:active.pop(name,None)
 print(json.dumps({'file':file,'run':name,'wrapper_child_exit':code}),flush=True)
 return job,True

def verify(file):
 p=prefix(file)
 if file in verified or not all(completed(j[3]) for j in jobs(file)):return
 base=root/'raw'/(p+'-parent-base-normal.jsonl');ar=load(base)
 proof={}
 for side in ['parent','head']:
  b=root/'raw'/(p+'-'+side+'-guard-normal.jsonl');br=load(b)
  assert raw_findings(base)==raw_findings(b),(file,side,'ordinary findings changed')
  assert helpers['non_telemetry_bytes'](base)==helpers['non_telemetry_bytes'](b),(file,side,'ordinary census bytes changed')
  assert helpers['only_cache_telemetry_changed'](ar,br),(file,side,'changes beyond cache telemetry')
  assert [r.get('units') for r in ar if 'units' in r]==[r.get('units') for r in br if 'units' in r],(file,side,'unit inventory changed')
  proof[side]={'normal_findings_byte_identical':True,'all_header_file_unit_and_finding_bytes_identical':True,'only_cache_telemetry_may_differ':True,'normal_unit_inventory_identical':True}
 result=subprocess.run([sys.executable,str(root/'pair-guards.py'),p],cwd=repo);print('pair_wrapper_child_exit='+str(result.returncode),flush=True)
 if result.returncode:raise RuntimeError('full pair failed '+file)
 result=subprocess.run([sys.executable,str(root/'render-findings.py'),p],cwd=repo);print('render_wrapper_child_exit='+str(result.returncode),flush=True)
 if result.returncode:raise RuntimeError('full render failed '+file)
 (root/(p+'-file.json')).write_text(json.dumps({'file':file,'prefix':p,'common_original_parent_baseline':True,'proof':proof},indent=2)+'\n')
 verified.add(file);checkpoint(p+'-verified')

# Start larger files early; all three children are independent.
ordered=sorted(files,key=lambda x:(-Path(x).stat().st_size,x))
with concurrent.futures.ThreadPoolExecutor(max_workers=workers) as pool:
 future_jobs={pool.submit(one,j):j for file in ordered for j in jobs(file)}
 for future in concurrent.futures.as_completed(future_jobs):
  job,ran=future.result()
  if ran:checkpoint(job[3])
  verify(job[0])
for file in files:verify(file)
print('verified full file checks='+str(len(verified))+' requested='+str(len(files)),flush=True)

sys.exit(0 if len(verified)==len(files) else 2)
