import os,sys,subprocess,json,time,signal,hashlib,gzip,shutil,resource
from pathlib import Path
out=Path('/workspace/scratch/probe-report/stage3/census/find-ancestor-probe');repo=out.parents[2]
adapted=Path('/workspace/scratch/fa-adapted')
base={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
base.update(LATENT_ASSERT_NO_OUTPUT='1',GOMEMLIMIT='5GiB',GOGC='70',ADAMIC_PROBE_BINDINGS='1')
def checkpoint(event):
 for cmd in [['git','add','-f','--',str(out)]]:
  p=subprocess.run(cmd,cwd=repo);print(event,'add child_exit='+str(p.returncode),flush=True);p.check_returncode()
 tree=subprocess.check_output(['git','write-tree'],cwd=repo,text=True).strip()
 sha=subprocess.check_output(['git','commit-tree',tree,'-p','1060360b'],cwd=repo,input='census: probe binding '+event+' checkpoint\n',text=True).strip()
 ref='refs/heads/typescript/find-ancestor-probe-census-checkpoint-'+event
 with (out/('checkpoint-'+event+'.log')).open('w') as f:
  p=subprocess.run(['git','push','origin',sha+':'+ref],cwd=repo,stdout=f,stderr=subprocess.STDOUT)
 (out/('checkpoint-'+event+'.json')).write_text(json.dumps(dict(sha=sha,ref=ref,exit=p.returncode),indent=2)+'\n')
 print(event,'push child_exit='+str(p.returncode),sha,flush=True);p.check_returncode()
def findings(path):
 rows=[json.loads(l) for l in path.read_text().splitlines()]
 return [f for row in rows for f in row.get('findings',[])]
active={};provenance={}
import atexit
def cleanup():
 for state in active.values():
  p=state['p']
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
  p.wait()
atexit.register(cleanup)
for signum in [signal.SIGTERM,signal.SIGINT]:signal.signal(signum,lambda sig,frame:sys.exit(128+sig))
for side,pin in [('parent','555b1645'),('head','8afae944')]:
 tree=Path('/workspace/scratch/probe-'+side);binary=Path('/workspace/scratch/probe-'+side+'-entry-bin')
 sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=tree,text=True).strip()
 assert sha==subprocess.check_output(['git','rev-parse',pin],cwd=tree,text=True).strip()
 assert subprocess.run(['git','diff','--quiet',pin,'--','internal','stage3/census/latent','stage3/meter'],cwd=tree).returncode==0
 provenance[side]=dict(sha=sha,checker=subprocess.check_output(['git','rev-parse','HEAD'],cwd=tree/'cohere',text=True).strip(),typescript_go=subprocess.check_output(['git','rev-parse','HEAD'],cwd=tree/'cohere/TypeScript',text=True).strip(),binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest())
(out/'provenance.json').write_text(json.dumps(provenance,indent=2)+'\n')
(out/'REPORT.md').write_text('# Probe binding census checkpoint\n\nTask 448a51k. Adapted source hashes match c6f6688c. Both own-pin guarded entry binaries built. Measurements pending. The supplied lists have 30 parent refusal-scan sites and five head sites, so all 35 are included. No resolution verdict yet.\n')
checkpoint('prepared')
for side in ['parent','head']:
 name=side+'-entry-run';cmd=['/workspace/scratch/probe-'+side+'-entry-bin',str(adapted/'src/tsc/tsc.ts'),str(out/(name+'.jsonl'))]
 env=dict(base,LATENT_TIMES=str(out/(name+'-units.jsonl')))
 stdout=(out/(name+'.stdout')).open('wb');stderr=(out/(side+'-stderr.txt')).open('wb')
 p=subprocess.Popen(cmd,cwd='/workspace/scratch/probe-'+side,env=env,stdout=stdout,stderr=stderr,start_new_session=True)
 active[side]=dict(p=p,stdout=stdout,stderr=stderr,started=time.monotonic(),last=time.monotonic(),stamp=None,cmd=cmd,stopped=None)
 (out/(name+'.pid.json')).write_text(json.dumps(dict(pid=p.pid,command=cmd,flags={k:v for k,v in env.items() if k.startswith('LATENT_') or k in ['ADAMIC_PROBE_BINDINGS','GOMEMLIMIT','GOGC']}),indent=2)+'\n')
lastbeat=0
while active:
 for side,state in list(active.items()):
  p=state['p'];name=side+'-entry-run'
  paths=[out/(name+'.jsonl'),out/(name+'-units.jsonl'),out/(name+'.stdout'),out/(side+'-stderr.txt')]
  stamp=[(str(f),f.stat().st_size,f.stat().st_mtime_ns) for f in paths if f.exists()]
  if stamp!=state['stamp']:state['last']=time.monotonic();state['stamp']=stamp
  if p.poll() is None and time.monotonic()-state['last']>1800:
   state['stopped']='silent for 1800 seconds';os.killpg(p.pid,signal.SIGTERM)
   try:p.wait(timeout=10)
   except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL)
  if p.poll() is not None:
   rc=p.wait()
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   state['stdout'].close();state['stderr'].close()
   status=dict(command=state['cmd'],cwd='/workspace/scratch/probe-'+side,exit=rc,wall_seconds=time.monotonic()-state['started'],stopped=state['stopped'])
   (out/(name+'.status.json')).write_text(json.dumps(status,indent=2)+'\n')
   print(side,'own child_exit='+str(rc),flush=True)
   raw=out/(side+'-stderr.txt')
   with raw.open('rb') as src,gzip.open(str(raw)+'.gz','wb') as dst:shutil.copyfileobj(src,dst)
   # Keep stderr whole, compressed. Hash authenticates decompression.
   (out/(side+'-stderr.sha256')).write_text(hashlib.sha256(raw.read_bytes()).hexdigest()+'  '+raw.name+'\n');raw.unlink()
   if rc==0:
    actual=findings(out/(name+'.jsonl'));expected=findings(Path('/workspace/scratch/probe-evidence')/(name+'.jsonl'))
    validation=dict(identical=actual==expected,actual_findings=len(actual),expected_findings=len(expected),raw_stream_identical=(out/(name+'.jsonl')).read_bytes()==(Path('/workspace/scratch/probe-evidence')/(name+'.jsonl')).read_bytes())
    (out/(side+'-stream-validation.json')).write_text(json.dumps(validation,indent=2)+'\n')
    print(side,'findings identical='+str(validation['identical']),flush=True)
   checkpoint(name)
   del active[side]
 if time.monotonic()-lastbeat>60:
  print('active '+','.join(active),flush=True);lastbeat=time.monotonic()
 time.sleep(1)
print('measurement controller child_exit=0',flush=True)
