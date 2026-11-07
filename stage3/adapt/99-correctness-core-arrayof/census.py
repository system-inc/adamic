from pathlib import Path
import subprocess,json,os,time,signal,sys,hashlib
s=Path('/workspace/scratch/arrayof');label=sys.argv[1];tree=Path(sys.argv[2]);env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')};env.update(LATENT_FULL='1',LATENT_ONLY_FILE=str(tree/'src/compiler/core.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(s/(label+'-unit-times.jsonl')))
def snapshot():
 return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()}
frozen=snapshot();(s/(label+'-census-source-manifest.json')).write_text(json.dumps(frozen,indent=2)+'\n')
start=time.monotonic()
with (s/(label+'-census.log')).open('w') as log:
 p=subprocess.Popen([str(s/'latent'),str(tree/'src/compiler'),str(s/(label+'-census.jsonl'))],cwd='/workspace/adamic',env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 try:rc=p.wait(timeout=1700)
 finally:
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
  p.wait()
assert snapshot()==frozen,'source changed during census'
(s/(label+'-census-result.json')).write_text(json.dumps(dict(exit=rc,seconds=time.monotonic()-start,processGroupCleaned=True),indent=2)+'\n');assert rc==0
