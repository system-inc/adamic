"""Fresh production Lower observations, paired per root, checkpoints by ten."""
from pathlib import Path
import concurrent.futures,subprocess,json,sys,time
out=Path(__file__).resolve().parent;s=Path('/workspace/loose-slot-scratch/eleven');repo=out.parents[3]
sys.path.insert(0,str(out.parent));from measure import run
while not all((s/(w+'-census')).exists()and(s/(w+'-input.ready')).exists()for w in ['base','head']):time.sleep(3)
def one(task):
 which,index=task;roots=sorted((s/('adapted-'+which)/'src/compiler').rglob('*.ts'));assert len(roots)==79
 d=out/'evidence'/which/f'root-{index:02}'
 if d.with_suffix('.time.json').exists():assert json.loads(d.with_suffix('.time.json').read_text())['exit_code']==0;return False
 result=run([s/(which+'-census'),roots[index-1],d.with_suffix('.json')],d,cwd=s/which)
 print(which,index,roots[index-1].name,result,flush=True);return True
for first in range(1,80,10):
 last=min(first+9,79)
 with concurrent.futures.ThreadPoolExecutor(max_workers=2)as pool:changed=list(pool.map(one,[(w,i)for i in range(first,last+1)for w in ['base','head']]))
 if not any(changed):continue
 subprocess.run(['git','add','stage3/census/loose-slot-census'],cwd=repo,check=True)
 subprocess.run(['git','commit','-m',f'stage3: checkpoint -11 catch census roots {first}-{last} at head and base'],cwd=repo,check=True)
 subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/loose-slot-census'],cwd=repo,check=True)
