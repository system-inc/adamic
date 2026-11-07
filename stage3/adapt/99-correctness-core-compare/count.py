import os,json,subprocess,signal,time,hashlib,tarfile,shutil,tempfile
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor,as_completed
from collections import Counter
r=Path('/workspace/adamic');s=Path('/workspace/scratch/core-compare');units=json.loads((r/'stage3/lane/shard-plan.json').read_text())['units'];results=s/'results';results.mkdir(exist_ok=True)
# Product views supply the locked packages from stage3/typescript, not the tree.
modules=r/'stage3/typescript/node_modules'
if (s/'counted/node_modules').exists():
 assert (s/'counted/package-lock.json').read_bytes()==(r/'stage3/typescript/package-lock.json').read_bytes(),'package lock drift'
 assert not modules.exists(),'package destination already exists'
 shutil.move(str(s/'counted/node_modules'),str(modules))
summary={'complete':False,'units':[],'counts':{},'mixed':0,'differences':0,'calls':0,'pieces':[]}
if (s/'COUNT.json').exists():summary=json.loads((s/'COUNT.json').read_text())
assert all(not p['failed'] for p in summary['pieces'])
shm=Path(tempfile.mkdtemp(prefix='core-compare-',dir='/dev/shm'))
print('RESUME '+str(len(summary['units']))+' units; synchronous memory-backed records; two workers',flush=True)
def snapshot():
 return {str(p.relative_to(s/'counted')):hashlib.sha256(p.read_bytes()).hexdigest() for base in ['src','built/local'] for p in sorted((s/'counted'/base).rglob('*')) if p.is_file()}
frozen=snapshot();(s/'counted-manifest.json').write_text(json.dumps(frozen,indent=2)+'\n');(s/'recorder-provenance.json').write_text(json.dumps({'version':4,'sha256':hashlib.sha256((s/'count-preload.cjs').read_bytes()).hexdigest(),'schema':'BINARY-SCHEMA.json'},indent=2)+'\n')
def run(u,d):
 temporary=shm/u['name'];temporary.mkdir()
 env=dict(os.environ,STAGE3_ADAPTED_TREE=str(s/'counted'),COMPARE_CALL_DIR=str(temporary),COMPARE_SHARD=u['name'],STAGE3_SHARD_SECONDS='90',NODE_OPTIONS='--require='+str(s/'count-preload.cjs'),NODE_DISABLE_COMPILE_CACHE='1')
 try:
  with (d/(u['name']+'.log')).open('w') as log:
   p=subprocess.Popen(['python3',str(r/'stage3/lane/shard.py'),u['name'],str(results)],cwd=r,env=env,stdout=log,stderr=log,start_new_session=True)
   try:code=p.wait(timeout=110)
   finally:
    try:os.killpg(p.pid,signal.SIGKILL)
    except ProcessLookupError:pass
    p.wait()
  report=json.loads((results/u['name']/'report.json').read_text());return u,code,report
 finally:
  for file in temporary.iterdir():shutil.move(str(file),str(d/'calls'/file.name))
  temporary.rmdir()
for start in range(0,len(units),20):
 if start<len(summary['units']):continue
 d=s/f'piece-{start//20+1:02d}';d.mkdir();(d/'calls').mkdir();begin=time.monotonic()
 with ThreadPoolExecutor(max_workers=2) as pool: observations=[f.result() for f in as_completed([pool.submit(run,u,d) for u in units[start:start+20]])]
 assert frozen==snapshot(),'measurement compiler changed'
 counts=Counter();mixed=diff=calls=0
 histogram=Counter()
 for f in sorted((d/'calls').glob('*.bin')):histogram.update(f.read_bytes())
 for code,n in histogram.items():
  assert code<81,('unexpected input or result',code)
  pair,values=divmod(code,9);types=['undefined','number','string'];a,b=types[pair//3],types[pair%3];original,adapted=values//3-1,values%3-1
  counts[a+'/'+b]+=n;mixed+=({a,b}=={'string','number'})*n;diff+=(original!=adapted)*n;calls+=n
 (d/'binary-histogram.json').write_text(json.dumps(dict(sorted(histogram.items())),indent=2)+'\n')
 failed=[]
 for u,code,report in observations:
  if code or report['status']!='pass':failed.append({'unit':u['name'],'exit':code,'report':report})
  shutil.copytree(results/u['name'],d/u['name'],ignore=shutil.ignore_patterns('adapted-tree'))
  shutil.rmtree(results/u['name']/'adapted-tree',ignore_errors=True)
 record={'piece':start//20+1,'seconds':time.monotonic()-begin,'units':[u['name'] for u in units[start:start+20]],'calls':calls,'counts':dict(counts),'mixed':mixed,'differences':diff,'failed':failed,'recorderVersion':4,'bytesPerCall':1,'storage':'memory-backed synchronous files copied after each unit','workers':2}
 (d/'piece.json').write_text(json.dumps(record,indent=2)+'\n')
 summary['pieces'].append(record);summary['calls']+=calls;summary['mixed']+=mixed;summary['differences']+=diff
 for k,n in counts.items():summary['counts'][k]=summary['counts'].get(k,0)+n
 if not failed:summary['units']+=record['units']
 summary['complete']=len(summary['units'])==len(units)
 (s/'COUNT.json').write_text(json.dumps(summary,indent=2)+'\n')
 files=sorted(p for p in d.rglob('*') if p.is_file());(d/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(d))+'\n' for p in files))
 with tarfile.open(s/(d.name+'.tar.gz'),'w:gz',compresslevel=1) as tar:tar.add(d,arcname=d.name)
 print(json.dumps({k:record[k] for k in ['piece','seconds','calls','counts','mixed','differences']}),flush=True)
 if mixed or diff or failed:raise SystemExit('STOP: mixed input, counterexample, or failed lane unit; see COUNT.json')
 shutil.rmtree(d)
assert len(summary['units'])==len(set(summary['units']))==len(units)
shm.rmdir()
print('COMPLETE',flush=True)
