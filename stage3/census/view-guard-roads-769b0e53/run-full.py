#!/usr/bin/env python3
"""Resume file-specific full-mode original/seam pairs; checkpoint every child."""
import concurrent.futures,fcntl,gzip,hashlib,json,os,signal,subprocess,sys,threading
from pathlib import Path
root=Path(__file__).resolve().parent;repo=root.parents[2];sys.path.insert(0,str(root));import runpy
helpers=runpy.run_path(str(root/'pair-guards.py'));raw_findings=helpers['raw_findings'];load=helpers['load']
files=json.loads(Path(sys.argv[1]).read_text());workers=int(sys.argv[2]) if len(sys.argv)>2 else 3
assert all(Path(x).is_file() and x.startswith('/workspace/roads-adapted/src/') for x in files)
lock=threading.Lock();active={};stopping=False

def stop(sig,frame):
 global stopping
 stopping=True
 with lock:
  for proc in active.values():proc.terminate()
signal.signal(signal.SIGTERM,stop);signal.signal(signal.SIGINT,stop)

def checkpoint(label):
 with lock:
  result=subprocess.run(['bash',str(root/'checkpoint.sh'),label],cwd=repo)
  print('checkpoint_wrapper_child_exit='+str(result.returncode),flush=True)
  if result.returncode:raise RuntimeError('checkpoint failed')

def one(file):
 identifier=hashlib.sha256(file.encode()).hexdigest()[:10];prefix='full-'+identifier
 for side in ['parent','head']:
  for mode in ['base','guard']:
   if side=='head' and mode=='base':continue
   if stopping:return
   name=f'{prefix}-{side}-{mode}';output=root/'raw'/(name+'-normal.jsonl');metadata=root/'raw'/(name+'.json')
   if metadata.exists() and json.loads(metadata.read_text())['child_exit']==0:continue
   env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=file,GOMAXPROCS='2',ADAMIC_TYPESCRIPT_SOURCE='/workspace/roads-typescript',LATENT_TIMES=str(root/'raw'/(name+'-times.jsonl')))
   env.pop('ROADS_GUARD_MUTANT',None);env.pop('ROADS_GUARD_STREAM',None)
   binary=f'/workspace/roads-{side}-census'
   if mode=='guard':
    binary=f'/workspace/roads-{side}-guard-census';env['ROADS_GUARD_STREAM']=str(root/'raw'/(f'{prefix}-{side}-records.jsonl'))
    Path(env['ROADS_GUARD_STREAM']).unlink(missing_ok=True)
   proc=subprocess.Popen([sys.executable,str(root/'run.py'),str(root/'raw'/name),binary,'/workspace/roads-adapted/src/tsc/tsc.ts',str(output)],cwd=repo,env=env)
   with lock:active[name]=proc
   code=proc.wait()
   with lock:active.pop(name,None)
   print(json.dumps({'file':file,'run':name,'wrapper_child_exit':code}),flush=True)
   checkpoint(name)
   if code:raise RuntimeError(f'{name}: wrapper child exit {code}')
  a=root/'raw'/f'{prefix}-parent-base-normal.jsonl';b=root/'raw'/f'{prefix}-{side}-guard-normal.jsonl'
  assert raw_findings(a)==raw_findings(b),(file,side,'ordinary findings changed')
  ar,br=load(a),load(b)
  assert [r.get('units') for r in ar if 'units' in r]==[r.get('units') for r in br if 'units' in r],(file,side,'unit inventory changed')
 result=subprocess.run([sys.executable,str(root/'pair-guards.py'),prefix],cwd=repo)
 print('pair_wrapper_child_exit='+str(result.returncode),flush=True)
 if result.returncode:raise RuntimeError('full pair failed '+file)
 (root/(prefix+'-file.json')).write_text(json.dumps({'file':file,'prefix':prefix,'normal_findings_byte_identical':True,'normal_unit_inventory_identical':True},indent=2)+'\n')
 checkpoint(prefix+'-verified')

with concurrent.futures.ThreadPoolExecutor(max_workers=workers) as pool:
 results=list(pool.map(one,files))
print('full file checks completed='+str(len(results)),flush=True)
