import concurrent.futures,json,subprocess,time
from pathlib import Path
root=Path.cwd();e=root/'review/compiler/loom-v4';selected=json.loads((e/'selected-tests.json').read_text());done={(r['package'],r.get('test')) for r in map(json.loads,(e/'results.jsonl').read_text().splitlines()) if 'test' in r};jobs=[(p,t) for p,tests in selected.items() for t in tests if (p,t) not in done]
for p in sorted({p for p,t in jobs}):
 binary=e/(p.replace('/','-')+'.test')
 if not binary.exists():
  with (e/(p.replace('/','-')+'-remaining-build.log')).open('w') as log:
   r=subprocess.run(['go','test','-c','-o',str(binary),'./'+p],stdout=log,stderr=subprocess.STDOUT,timeout=180)
  if r.returncode:raise RuntimeError('build failed: '+p)
def check(job):
 p,t=job;start=time.monotonic()
 with (e/(p.replace('/','-')+'-'+t+'-remaining.log')).open('w') as log:
  try:r=subprocess.run([str(e/(p.replace('/','-')+'.test')),'-test.run','^'+t+'$','-test.count=1','-test.timeout=90s','-test.v','-test.parallel=1'],cwd=root/p,stdout=log,stderr=subprocess.STDOUT,timeout=95);code=r.returncode
  except subprocess.TimeoutExpired:code=124
 seconds=round(time.monotonic()-start,3);return dict(package=p,test=t,seconds=seconds,exit=code,grain=seconds<60)
with (e/'remaining-results.jsonl').open('w') as out,concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
 for row in pool.map(check,jobs):
  out.write(json.dumps(row)+'\n');out.flush();print(row,flush=True)
for p in {p for p,t in jobs}:(e/(p.replace('/','-')+'.test')).unlink()
