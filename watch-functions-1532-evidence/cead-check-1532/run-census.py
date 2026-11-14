from pathlib import Path
import os,subprocess,json,time,hashlib
s=Path('/tmp/watch-file-check-cead');tree=Path('/tmp/watch-functions-1532/main-tree');f=tree/'src/compiler/sys.ts';base=(s/'sys-main.ts').read_bytes()
while not (s/'latent').exists():time.sleep(1)
assert f.read_bytes()==base
try:
 for side in ['main','after']:
  f.write_bytes((s/('sys-'+side+'.ts')).read_bytes());out=s/('census-'+side);out.mkdir()
  env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')};env.update(LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(f),LATENT_TIMES=str(out/'unit-times.jsonl'),TMPDIR=str(s/'tmp'),GOMAXPROCS='2',GOMEMLIMIT='6GiB')
  cmd=[str(s/'latent'),str(tree/'src/compiler'),str(out/'raw.jsonl')];t=time.monotonic();frozen=hashlib.sha256(f.read_bytes()).hexdigest()
  with (out/'census.log').open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=1800)
  assert hashlib.sha256(f.read_bytes()).hexdigest()==frozen
  (out/'result.json').write_text(json.dumps({'command':cmd,'flags':{k:v for k,v in env.items() if k.startswith('LATENT_')},'exit':r.returncode,'seconds':time.monotonic()-t,'sourceSha256':frozen},indent=2)+'\n');assert r.returncode==0
finally:f.write_bytes(base)
(s/'census-done').write_text('complete\n')
