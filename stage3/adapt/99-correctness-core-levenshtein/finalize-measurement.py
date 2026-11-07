import subprocess,os,json,time
from pathlib import Path
s=Path('/workspace/scratch/levenshtein');repo=Path('/workspace/adamic');count=json.loads((s/'COUNT.json').read_text());assert count['complete'] and count['differences']==0
assert json.loads((s/'control.json').read_text())['exit']==0
# COUNT.complete is published before the final archive closes. Require producer exit.
begin=time.monotonic()
while 'accepted count exit 0' not in (s/'resume-count.log').read_text():
 assert time.monotonic()-begin<180,'final count archive did not close within bound'
 print('waiting for count producer exit and closed archives',flush=True);time.sleep(1)

env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
for name,args in [('reconcile',['python3',str(s/'reconcile.py')]),('checked-sites',['node',str(s/'checked-sites.cjs')]),('implement',['python3',str(s/'implement.py')]),('contract-before-apply',['node',str(s/'contract.cjs')])]:
 with (s/(name+'-checkpoint.log')).open('w') as log:r=subprocess.run(args,cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT)
 print(name,r.returncode,flush=True);assert r.returncode==0,name
 if name=='checked-sites':(s/'checked-read-sites.json').write_text((s/(name+'-checkpoint.log')).read_text())
with (s/'after-apply-driver.log').open('w') as log:r=subprocess.run(['python3',str(s/'after-apply.py')],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT)
print('after-apply',r.returncode,flush=True);raise SystemExit(r.returncode)
