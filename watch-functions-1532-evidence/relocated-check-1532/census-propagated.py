import pathlib,os,subprocess,time,json,hashlib
s=pathlib.Path('/tmp/watch-file-check-1532');tree=pathlib.Path('/tmp/watch-functions-1532/main-tree');target=tree/'src/compiler/sys.ts'
while not (s/'latent').exists() or not (s/'stock-propagated.json').exists(): time.sleep(1)
base=(s/'sys-main.ts').read_bytes();assert target.read_bytes()==base
results={}
try:
 for side in ['propagated']:
  target.write_bytes((s/('sys-'+side+'.ts')).read_bytes())
  out=s/('census-'+side);out.mkdir()
  env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
  env.update(LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(target),LATENT_TIMES=str(out/'unit-times.jsonl'),TMPDIR=str(s/'tmp'),GOTMPDIR=str(s/'tmp'),GOCACHE=str(s/'go-cache'),GOMAXPROCS='2',GOMEMLIMIT='6GiB')
  cmd=[str(s/'latent'),str(tree/'src/compiler'),str(out/'raw.jsonl')]
  frozen=hashlib.sha256(target.read_bytes()).hexdigest();started=time.monotonic()
  with (out/'census.log').open('w') as log:
   result=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=1800)
  assert hashlib.sha256(target.read_bytes()).hexdigest()==frozen
  results[side]={'exit':result.returncode,'seconds':time.monotonic()-started,'command':cmd,'flags':{k:v for k,v in env.items() if k.startswith('LATENT_')},'sourceSha256':frozen}
  (out/'result.json').write_text(json.dumps(results[side],indent=2)+'\n')
  assert result.returncode==0
finally: target.write_bytes(base)
(s/'census-propagated-results.json').write_text(json.dumps(results,indent=2)+'\n')
