import os,time,subprocess,signal,json,shutil,hashlib
from pathlib import Path
s=Path('/workspace/scratch/core-extend');repo=Path('/workspace/adamic');phases={}
while not all((s/n/'patch-set.md').exists() for n in ['before','after']) or not all((s/n).exists() for n in ['latent','adamic']):
 print('waiting for both complete applies and compiler binaries',flush=True);time.sleep(30)
def run(name,args,env=None):
 start=time.monotonic()
 with (s/(name+'.log')).open('w') as log:
  p=subprocess.Popen(args,cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=p.wait(timeout=1400)
  except subprocess.TimeoutExpired:code=124
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 phases[name]={'exit':code,'seconds':time.monotonic()-start};(s/'verification.json').write_text(json.dumps(phases,indent=2)+'\n');print(name,phases[name],flush=True);return code
for name in ['before','after']:
 roots=sorted(str(p) for p in (s/name/'src/compiler').rglob('*.ts'));assert len(roots)==79
 run('types-'+name,[str(s/'adamic'),'types',*roots])
 env=dict(os.environ,LATENT_FULL='1',LATENT_ONLY_FILE=str(s/name/'src/compiler/core.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(s/('times-'+name+'.jsonl')))
 run('census-'+name,[str(s/'latent'),str(s/name/'src/compiler'),str(s/('sites-'+name+'.jsonl'))],env)
mutant=s/'mutant';shutil.copytree(s/'after/src',mutant/'src');pins=json.loads((repo/'stage3/adapt/99-correctness-core-extend/owners.json').read_text());p=mutant/'src/compiler/core.ts';text=p.read_text()
for pin in pins:
 assert text.count(pin['after'])==1;text=text.replace(pin['after'],pin['before'])
p.write_text(text)
env=dict(os.environ,LATENT_FULL='1',LATENT_ONLY_FILE=str(p),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(s/'times-mutant.jsonl'))
run('census-mutant',[str(s/'latent'),str(mutant/'src/compiler'),str(s/'sites-mutant.jsonl')],env)
