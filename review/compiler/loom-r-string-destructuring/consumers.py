from pathlib import Path
import concurrent.futures, subprocess, time, json
out=Path(__file__).resolve().parent
jobs=json.loads((out/'consumer-run-selection.json').read_text())
def run(job):
 selector='^('+'|'.join(job['tests'])+')$' if job['tests'] else '^$'
 args=['go','test',job['package'],'-run',selector,'-count=1','-timeout','90s','-v','-parallel','4']
 name=job['package'].removeprefix('./').replace('/','_');start=time.monotonic()
 with (out/('consumer-'+name+'.log.txt')).open('w') as log:
  try:code=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT,timeout=150).returncode
  except subprocess.TimeoutExpired:code=124
 return dict(**job,command=args,exit=code,seconds=round(time.monotonic()-start,3))
rows=[]
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 for row in pool.map(run,jobs):
  rows.append(row);(out/'consumer-results.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(row),flush=True)
assert all(r['exit']==0 for r in rows)
