import json,os,subprocess,time,signal,hashlib
from pathlib import Path
s=Path('/workspace/scratch/levenshtein');repo=Path('/workspace/adamic');env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'));phases={}
def run(name,args,cwd=repo,limit=1700):
 start=time.monotonic()
 with (s/(name+'.log')).open('w') as log:
  p=subprocess.Popen(args,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=p.wait(timeout=limit)
  except subprocess.TimeoutExpired:code=124
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 phases[name]={'exit':code,'seconds':time.monotonic()-start};(s/'post-apply.json').write_text(json.dumps(phases,indent=2)+'\n');print(name,phases[name],flush=True);assert code==0,(name,code)
# Recheck both full trees with the same ordinary binary and all 79 roots.
for label in ['before','after']:
 roots=sorted((s/label/'src/compiler').rglob('*.ts'));assert len(roots)==79;run(label+'-final-types',[str(s/'adamic'),'types',*map(str,roots)],limit=300)
run('owner-contract',['node',str(repo/'stage3/adapt/99-correctness-core-levenshtein/contract.cjs')])
run('after-install',['npm','ci','--no-audit','--no-fund'],s/'after')
run('after-build',['npm','run','build'],s/'after')
run('comparison',['node',str(s/'compare.cjs')])
# Preserve parsed comparison JSON alongside the unmodified raw stdout log.
(s/'comparison.json').write_text((s/'comparison.log').read_text())
run('accepted-witness-final',['node',str(s/'witness.cjs')])
(s/'accepted-witness-final.json').write_text((s/'accepted-witness-final.log').read_text())
