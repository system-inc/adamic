"""Two independent children at a time; checkpoint only complete groups of ten."""
import concurrent.futures,json,subprocess,time
from pathlib import Path
from measure import run
s=Path('/workspace/loose-slot-scratch');repo=s/'head';out=Path(__file__).resolve().parent
while not ((s/'apply-main.done').exists() and (s/'apply-head.done').exists()):time.sleep(5)
def one(which,index):
 tree=s/('adapted-'+which)/'src/compiler';roots=sorted(tree.rglob('*.ts'));assert len(roots)==79
 root=roots[index-1];d=out/'evidence'/which/f'root-{index:02}'
 if d.with_suffix('.time.json').exists():
  assert json.loads(d.with_suffix('.time.json').read_text())['exit_code']==0
  return False
 result=run([s/(which+'-census'),root,d.with_suffix('.json')],d,cwd=s/which)
 print(which,index,root.name,result,flush=True);return True
for first in range(1,80,10):
 last=min(first+9,79);jobs=[(which,index)for index in range(first,last+1)for which in ['main','head']]
 with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
  changed=list(pool.map(lambda task:one(*task),jobs))
 if not any(changed):continue
 subprocess.run(['git','add','stage3/census/loose-slot-census'],cwd=repo,check=True)
 subprocess.run(['git','commit','-m',f'stage3: loose-slot census roots {first}-{last} at both SHAs'],cwd=repo,check=True)
 subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/loose-slot-census'],cwd=repo,check=True)
