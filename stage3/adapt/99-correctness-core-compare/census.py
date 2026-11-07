import json,os,time,signal,subprocess,hashlib,sys
from pathlib import Path
s=Path('/workspace/scratch/core-compare');side=sys.argv[1];tree=s/side;out=s/('census-'+side);out.mkdir(exist_ok=False)
def manifest():return {str(p.relative_to(tree/'src')):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()}
frozen=manifest();(out/'source-manifest.json').write_text(json.dumps(frozen,indent=2)+'\n')
env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')};env.update(LATENT_FULL='1',LATENT_ONLY_FILE=str(tree/'src/compiler/core.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(out/'unit-times.jsonl'))
started=time.monotonic()
with (out/'census.log').open('w') as log:
 p=subprocess.Popen([str(s/'latent'),str(tree/'src/compiler'),str(out/'raw.jsonl')],env=env,stdout=log,stderr=log,start_new_session=True)
 try:code=p.wait(timeout=1800)
 finally:
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
  p.wait()
assert frozen==manifest(),'census source writes'
result={'exit':code,'seconds':time.monotonic()-started,'sourceUnchanged':True,'processGroupCleaned':True,'flags':{k:v for k,v in env.items() if k.startswith('LATENT_')}};(out/'result.json').write_text(json.dumps(result,indent=2)+'\n');assert code==0,result
print(json.dumps(result),flush=True)
