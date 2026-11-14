"""Run at most ten ordinary roots, three independent children at a time."""
import concurrent.futures,subprocess,sys
from pathlib import Path
first=int(sys.argv[1]);last=int(sys.argv[2]);assert 1<=first<=last<=79 and last-first<10
def run(i):
 rc=subprocess.call([sys.executable,str(Path(__file__).with_name('run-one.py')),str(i)])
 print(f'scheduler root={i} child_exit={rc}',flush=True)
 return rc
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:codes=list(pool.map(run,range(first,last+1)))
sys.exit(next((rc for rc in codes if rc),0))
