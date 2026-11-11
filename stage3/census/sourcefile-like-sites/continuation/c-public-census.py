import pathlib,time,json,sys
sys.dont_write_bytecode=True
from measure import run
from variants import checkpoint,invoke
from concurrent.futures import ThreadPoolExecutor,as_completed
p=pathlib.Path(__file__).resolve().parent
names=['scanner.ts']+json.loads((p/'variant-files.json').read_text())
while not all((p/'c/raw'/(n.replace('/','__').removesuffix('.ts')+'.status.json')).exists() for n in names):time.sleep(2)
while not (p/'c-public/source-marker-diff.json').exists():time.sleep(2)
root=pathlib.Path('/workspace/scratch/baobab-road2-c-public')
with ThreadPoolExecutor(max_workers=2) as pool:
 pending={pool.submit(run,'c-public',root/'src/compiler',[n],1500,False,2):n for n in names}
 for future in as_completed(pending):
  future.result();checkpoint('road2-c-public-'+pending[future].replace('/','__'))
invoke(['node',str(p/'source-audit-c.cjs'),str(root),'c-public','stock'],p/'c-public/source-audit.log',600)
invoke(['python3',str(p/'cohere-c.py'),'c-public'],p/'c-public/cohere-controller.log',1500)
print('c public census workflow child exit=0',flush=True)
