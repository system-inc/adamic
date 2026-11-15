import json,subprocess,sys
from pathlib import Path
from measure import run
s=Path('/workspace/loose-slot-scratch');repo=s/'head';out=Path(__file__).resolve().parent
assert (s/"apply-main.done").exists() and (s/"apply-head.done").exists(), "adaptation not complete"
for first in range(1,80,10):
 last=min(first+9,79)
 for which in ['main','head']:
  tree=s/('adapted-'+which)/'src/compiler';roots=sorted(tree.rglob('*.ts'));assert len(roots)==79,len(roots)
  for index in range(first,last+1):
   root=roots[index-1];d=out/'evidence'/which/f'root-{index:02}'
   if d.with_suffix('.time.json').exists():continue
   result=run([s/(which+'-census'),root,d.with_suffix('.json')],d,cwd=s/which)
   print(which,index,root.name,result,flush=True)
 subprocess.run(['git','add','stage3/census/loose-slot-census'],cwd=repo,check=True)
 subprocess.run(['git','commit','-m',f'stage3: loose-slot census roots {first}-{last} at both SHAs'],cwd=repo,check=True)
 subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/loose-slot-census'],cwd=repo,check=True)
