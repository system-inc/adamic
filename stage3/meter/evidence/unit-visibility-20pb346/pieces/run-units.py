"""Resume bounded full-mode unit pairs; commit and push each completed side.

Two independent pairs may run concurrently. Each pair runs before then after.
Run from the provisioned repository. No whole latent pass is launched.
"""
from pathlib import Path
import concurrent.futures as futures
import gzip, hashlib, json, os, signal, subprocess, threading, time
repo=Path('/workspace/adamic')
root=Path('/workspace/scratch/20pb346-2')
evidence=Path(__file__).resolve().parent
source=root/'adapted/src/compiler'
plan_bytes=(evidence/'unit-plan.json').read_bytes()
plan=json.loads(plan_bytes)
plan_hash=hashlib.sha256(plan_bytes).hexdigest()
lock=threading.Lock()
active_lock=threading.RLock()
active={}
stopping=threading.Event()
results=evidence/'unit-results'
results.mkdir(exist_ok=True)
binaries={side:root/('bounded-'+side) for side in ('before','after')}
binary_hash={side:hashlib.sha256(path.read_bytes()).hexdigest() for side,path in binaries.items()}

def atomic_json(path,value):
 temporary=path.with_suffix(path.suffix+'.pending')
 temporary.write_text(json.dumps(value,indent=2)+'\n');temporary.replace(path)

def status():
 with active_lock:
  running={label:dict(pid=p.pid,seconds=round(time.monotonic()-start,1)) for label,(p,start) in active.items()}
 complete=list(results.glob('*/identity.json'))
 atomic_json(root/'unit-supervisor-status.json',dict(completed_pairs=len(complete),planned_pairs=len(plan['pieces']),active=running,updated=time.time()))

def publish(folder,message,logfile):
 with lock:
  subprocess.run(['git','add','--',str(folder)],cwd=repo,check=True)
  changed=subprocess.run(['git','diff','--cached','--quiet'],cwd=repo).returncode
  if changed:
   with logfile.open('a') as log:subprocess.run(['git','commit','-m',message],cwd=repo,stdout=log,stderr=subprocess.STDOUT,check=True)
  for attempt in range(3):
   with logfile.open('a') as log:
    push=subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/meter-stream-records-2'],cwd=repo,stdout=log,stderr=subprocess.STDOUT)
   if push.returncode==0:break
   if attempt==2:raise RuntimeError('plain push failed; result remains committed')
   time.sleep(5)

def stop(signum,frame):
 stopping.set()
 with active_lock:
  for p,_ in active.values():
   if p.poll() is None:
    try:os.killpg(p.pid,signal.SIGTERM)
    except ProcessLookupError:pass
 raise KeyboardInterrupt
signal.signal(signal.SIGTERM,stop)
signal.signal(signal.SIGINT,stop)

def run(piece):
 assert hashlib.sha256(Path(piece['file']).read_bytes()).hexdigest()==piece['source_sha256']
 folder=results/piece['name'];folder.mkdir(exist_ok=True)
 raw={}
 for side in ('before','after'):
  if stopping.is_set():raise RuntimeError('stopping')
  work=root/'unit-runs'/piece['name']/side;work.mkdir(parents=True,exist_ok=True)
  summary=folder/(side+'-result.json')
  if summary.exists():
   result=json.loads(summary.read_text())
   assert result['plan_sha256']==plan_hash and result['binary_sha256']==binary_hash[side]
   assert result['exit_code']==0,result
   raw[side]=gzip.decompress((folder/(side+'-latent.jsonl.gz')).read_bytes())
   assert hashlib.sha256(raw[side]).hexdigest()==result['sha256']
   publish(folder,'meter: recover '+piece['name']+' '+side+' unit piece',work/'publication.txt')
  else:
   env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
   env.update(LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=piece['file'],LATENT_PROOF_UNITS='\n'.join(piece['units']),LATENT_TIMES=str(work/'latent-times.jsonl'),GOMEMLIMIT='4GiB')
   start=time.monotonic()
   with (work/'output.txt').open('w') as log:
    process=subprocess.Popen([str(binaries[side]),str(source),str(work/'latent.jsonl')],env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
    atomic_json(work/'process.json',dict(pid=process.pid,start_monotonic=start,argv=[str(binaries[side]),str(source),str(work/'latent.jsonl')]))
    label=piece['name']+'/'+side
    with active_lock:active[label]=(process,start)
    try:
     try:code=process.wait(timeout=1200)
     except subprocess.TimeoutExpired:
      os.killpg(process.pid,signal.SIGTERM)
      try:code=process.wait(timeout=10)
      except subprocess.TimeoutExpired:os.killpg(process.pid,signal.SIGKILL);code=process.wait()
    finally:
     if process.poll() is None:
      os.killpg(process.pid,signal.SIGTERM)
      try:process.wait(timeout=10)
      except subprocess.TimeoutExpired:os.killpg(process.pid,signal.SIGKILL);process.wait()
     with active_lock:active.pop(label,None)
   wall=time.monotonic()-start
   data=(work/'latent.jsonl').read_bytes()
   rows=[json.loads(line) for line in data.splitlines()]
   keys=[u['where']+'|'+u['kind'] for row in rows[1:] for u in row['units']]
   result=dict(side=side,exit_code=code,wall_seconds=wall,wall_budget_seconds=1200,GOMEMLIMIT='4GiB',plan_sha256=plan_hash,binary_sha256=binary_hash[side],sha256=hashlib.sha256(data).hexdigest(),bytes=len(data),units=len(keys),file=piece['file'],proof_only_selection=True)
   for filename in ('latent.jsonl','latent-times.jsonl','output.txt'):
    destination=folder/(side+'-'+filename+'.gz')
    temporary=destination.with_suffix('.pending')
    with temporary.open('wb') as target:
     with gzip.GzipFile(fileobj=target,mode='wb',mtime=0) as writer:writer.write((work/filename).read_bytes())
    temporary.replace(destination)
   if code==0:
    assert len(rows)==2 and rows[1]['file']==piece['file']
    assert keys==piece['units'],(piece['name'],side,'unit coverage differs')
   atomic_json(summary,result)
   raw[side]=data
   if side=='after':
    assert raw['before']==data,piece['name']+' latent bytes differ'
    atomic_json(folder/'identity.json',dict(identical=True,sha256=result['sha256'],bytes=len(data),units=len(piece['units']),plan_sha256=plan_hash))
   publish(folder,'meter: save '+piece['name']+' '+side+' unit piece',work/'publication.txt')
   assert code==0,result
   raw[side]=data
   print(json.dumps(dict(piece=piece['name'],**result)),flush=True)
  if side=='after':
   assert raw['before']==raw['after'],piece['name']+' latent bytes differ'
   if not (folder/'identity.json').exists():
    atomic_json(folder/'identity.json',dict(identical=True,sha256=hashlib.sha256(raw[side]).hexdigest(),bytes=len(raw[side]),units=len(piece['units']),plan_sha256=plan_hash))
    publish(folder,'meter: prove '+piece['name']+' unit piece identity',work/'publication.txt')
 return piece['name']

try:
 with futures.ThreadPoolExecutor(max_workers=2) as pool:
  todo=iter(plan['pieces']);pending={}
  for _ in range(2):
   piece=next(todo,None)
   if piece:pending[pool.submit(run,piece)]=piece['name']
  while pending:
   done,_=futures.wait(pending,timeout=55,return_when=futures.FIRST_COMPLETED)
   status()
   for task in done:
    name=pending.pop(task)
    try:task.result()
    except BaseException:
     stopping.set()
     with active_lock:
      for p,_ in active.values():
       if p.poll() is None:
        try:os.killpg(p.pid,signal.SIGTERM)
        except ProcessLookupError:pass
     raise
    piece=next(todo,None)
    if piece:pending[pool.submit(run,piece)]=piece['name']
 atomic_json(root/'unit-supervisor.done',dict(completed=True,pairs=len(plan['pieces'])))
finally:
 status()
