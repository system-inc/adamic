import concurrent.futures,json,subprocess,time
from pathlib import Path
root=Path.cwd();e=root/'review/compiler/loom-v4';binary=e/'oracle-fixed.test'
def check(name):
 start=time.monotonic()
 with (e/(name+'-fixed.log')).open('w') as log:
  r=subprocess.run([str(binary),'-test.run','^'+name+'$','-test.count=1','-test.timeout=90s','-test.v','-test.parallel=1'],cwd=root/'internal/oracle',stdout=log,stderr=subprocess.STDOUT,timeout=95)
 seconds=round(time.monotonic()-start,3);row=dict(package='internal/oracle',test=name,seconds=seconds,exit=r.returncode,grain=seconds<60);print(row,flush=True);return row
with (e/'fixed-results.jsonl').open('w') as out:
 for name in ['TestProduct_V4SanitizedCountedRuntimes'+str(i) for i in range(3)]+['TestRuntimeProductKeyHoldsEveryPart']:
  row=check(name);out.write(json.dumps(row)+'\n');out.flush()
 rows=[json.loads(line) for f in ['results.jsonl','remaining-results.jsonl'] for line in (e/f).read_text().splitlines()];failed=[r['test'] for r in rows if r.get('exit') and 'test' in r]
 with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
  for row in pool.map(check,failed):out.write(json.dumps(row)+'\n');out.flush()
binary.unlink()
