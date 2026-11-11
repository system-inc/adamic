from concurrent.futures import ThreadPoolExecutor
import subprocess,pathlib,json,time
out=pathlib.Path(__file__).resolve().parent
root=pathlib.Path('/workspace/tracing-measured-after/src/compiler')
roots=sorted(root.rglob('*.ts'))
assert len(roots)==79,len(roots)
logs=out/'types';logs.mkdir(exist_ok=True)
def check(root):
 start=time.monotonic()
 with (logs/(root.stem+'.log')).open('w') as log:
  result=subprocess.run(['/workspace/tracing-census-build/adamic','types',str(root)],stdout=log,stderr=subprocess.STDOUT)
 return dict(root=str(root),exit=result.returncode,seconds=time.monotonic()-start)
with ThreadPoolExecutor(max_workers=2) as pool:records=list(pool.map(check,roots))
(out/'types-roots.json').write_text(json.dumps(records,indent=2)+'\n')
print('roots='+str(len(records))+' zero-exit='+str(sum(r['exit']==0 for r in records)))
raise SystemExit(int(any(r['exit'] for r in records)))
