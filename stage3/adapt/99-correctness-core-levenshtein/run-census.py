import json,os,signal,subprocess,time,sys,hashlib
from pathlib import Path
s=Path('/workspace/scratch/levenshtein');label=sys.argv[1];tree=s/label
assert label in ['before','after','mutant']
env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')};env.update(LATENT_FULL='1',LATENT_ONLY_FILE=str(tree/'src/compiler/core.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(s/(label+'-unit-times.jsonl')))
begin=time.monotonic()
with (s/(label+'-census.log')).open('w') as log:
 p=subprocess.Popen([str(s/'latent'),str(tree/'src/compiler'),str(s/(label+'-census.jsonl'))],cwd='/workspace/adamic',env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 try:code=p.wait(timeout=1700)
 except subprocess.TimeoutExpired:code=124
 finally:
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
  p.wait()
result={'exit':code,'seconds':time.monotonic()-begin,'processGroupCleaned':True,'binarySha256':hashlib.sha256((s/'latent').read_bytes()).hexdigest(),'coreSha256':hashlib.sha256((tree/'src/compiler/core.ts').read_bytes()).hexdigest()}
(s/(label+'-census-result.json')).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True);assert code==0
