"""Three groups concurrently; each retains the copied per-root child harness."""
import concurrent.futures,subprocess,sys
from pathlib import Path
base=Path(__file__).resolve().parent
pairs=[(1,10),(11,20),(21,30),(31,40),(41,50),(51,60),(61,70),(71,79)]
def group(pair):
 a,z=pair
 return subprocess.call([sys.executable,str(base/'run.py'),f'group-{a}-{z}',sys.executable,str(base/'group.py'),str(a),str(z)])
for start in range(0,len(pairs),3):
 with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:codes=list(pool.map(group,pairs[start:start+3]))
 if any(codes):sys.exit(next(c for c in codes if c))
 subprocess.run([sys.executable,str(base/'checkpoint.py')],check=True)
