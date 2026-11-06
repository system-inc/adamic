"""One foreground measurement and immediate ordinary publication. Never loops over pieces."""
import argparse,datetime,gzip,hashlib,json,os,pathlib,signal,subprocess,time
O=pathlib.Path(__file__).resolve().parent; R=O.parents[2]; T=pathlib.Path('/workspace/container-fit-adapted'); B=pathlib.Path('/workspace/container-fit-build-v2')
p=argparse.ArgumentParser();p.add_argument('mode',choices=['latent','per-file']);p.add_argument('label',choices=['base','log','head']);p.add_argument('piece');p.add_argument('files',nargs='*');a=p.parse_args()
D=O/'evidence'/a.mode/a.label/a.piece;D.mkdir(parents=True,exist_ok=False)
def dump(n,v):(D/n).write_text(json.dumps(v,indent=2)+'\n')
sel=json.loads(pathlib.Path(a.files[0][1:]).read_text()) if a.files and a.files[0].startswith('@') else dict(schema=1,files=[dict(file='src/compiler/'+f,all_units=True,units=[]) for f in a.files]);dump('selection.json',sel)
expected_manifest=json.loads((O/'evidence/source-manifest.json').read_text())
assert expected_manifest=={str(f.relative_to(T)):hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted((T/'src').rglob('*')) if f.is_file()}
env=dict(os.environ,LATENT_FULL='0',LATENT_SPECULATIVE='0',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(D/'times.jsonl'),ANY_MEMBERS_CENSUS_ROOT=str(T),ANY_MEMBERS_CENSUS_PIECE=str(D/'selection.json'))
for k in list(env):
 if k.startswith('LATENT_MUTANT_') or k in ['LATENT_PIECE_PLAN','LATENT_PIECE_ID','LATENT_ONLY_FILE','ADAMIC_FIT_PAIRS_LOG']:env.pop(k)
if a.label=='log':env['ADAMIC_FIT_PAIRS_LOG']=str(D/'pairs.log')
cmd=[str(B/a.label/('latent' if a.mode=='latent' else 'census')),str(T/'src/compiler'),str(D/'records.jsonl')]
binary_sha=hashlib.sha256(pathlib.Path(cmd[0]).read_bytes()).hexdigest()
assert binary_sha==json.loads((O/'evidence/latent-build'/(a.label+'.json')).read_text())['selector_sha256']
start=time.monotonic();dump('started.json',dict(binary_sha256=binary_sha,utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),command=cmd,environment={k:env[k] for k in env if k.startswith(('LATENT_','ANY_MEMBERS_','ADAMIC_FIT_'))}))
with (D/'process.log').open('w') as log:
 child=subprocess.Popen(cmd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 try:
  while True:
   try:code=child.wait(timeout=20);break
   except subprocess.TimeoutExpired:
    elapsed=time.monotonic()-start
    f=D/('times.jsonl' if a.mode=='latent' else 'records.jsonl')
    print(a.mode,a.label,a.piece,round(elapsed,1),'seconds',f.stat().st_size if f.exists() else 0,'progress bytes',flush=True)
    if elapsed>=1500:
     os.killpg(child.pid,signal.SIGQUIT)
     try:code=child.wait(timeout=10)
     except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);code=child.wait()
     break
 finally:
  if child.poll() is None:os.killpg(child.pid,signal.SIGKILL);child.wait()
metrics=dict(returncode=code,wall_seconds=time.monotonic()-start,utc_end=datetime.datetime.now(datetime.timezone.utc).isoformat(),timed_out=time.monotonic()-start>=1500)
validation_error=None
try:
 if code==0:
  rows=[json.loads(x) for x in (D/'records.jsonl').read_text().splitlines()]
  if a.mode=='latent':
   norm=lambda s:s.replace(str(T)+'/', '')
   assert rows[0]['latent_mode']=='first-error'
   assert {norm(r['file']) for r in rows[1:]}=={f['file'] for f in sel['files']}
   units=[norm(u['where']) for r in rows[1:] for u in r['units']];assert len(units)==len(set(units))
   assert set(units)==set(sel['expected_units'])
   for r in rows[1:]:
    for finding in r['findings']:
     if finding['phase']=='lowering':assert norm(finding['unit']) in set(units)
   timers=[json.loads(x) for x in (D/'times.jsonl').read_text().splitlines()]
   assert {t['where'] for t in timers}=={u['where'] for r in rows[1:] for u in r['units'] if u['status'] in ['attempted','panic']}
   metrics.update(units=len(units),timers=len(timers),unit_wall_seconds=sum(t['wall_seconds'] for t in timers),unit_cpu_seconds=sum(t['cpu_seconds'] for t in timers))
  else:assert len([r for r in rows if len(r['roots'])==1 and r['roots'][0].endswith('.ts')])==79
except Exception as error:
 validation_error=repr(error);metrics['validation_error']=validation_error
for f in list(D.iterdir()):
 if f.suffix in ['.jsonl','.log']:
  metrics[f.name+'_sha256']=hashlib.sha256(f.read_bytes()).hexdigest()
  with gzip.open(str(f)+'.gz','wb') as z:z.write(f.read_bytes())
  f.unlink()
dump('metrics.json',metrics)
subprocess.run(['git','add','-f',str(D)],cwd=R,check=True)
paths=subprocess.check_output(['git','diff','--cached','--name-only'],cwd=R,text=True).splitlines();assert paths and all(f.startswith('stage3/census/container-fit/') for f in paths)
subprocess.run(['git','diff','--cached','--check'],cwd=R,check=True)
subprocess.run(['git','commit','-m',f'stage3/census: container fit {a.mode} {a.label} {a.piece}'],cwd=R,check=True)
with (O/'evidence/publication.log').open('a') as f:
 result=subprocess.run(['timeout','--kill-after=10s','120','git','push','origin','HEAD:refs/heads/typescript/container-fit-census'],cwd=R,stdout=f,stderr=subprocess.STDOUT)
print(json.dumps(metrics),flush=True)
if result.returncode:raise SystemExit('ordinary push failed, stop before next piece')
if code:raise SystemExit('measurement failed, evidence committed and pushed')

if validation_error:raise SystemExit("measurement validation failed; evidence committed and pushed: "+validation_error)
