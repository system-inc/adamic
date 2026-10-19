from pathlib import Path
import subprocess,time,json
out=Path(__file__).resolve().parent;rows=[]
for job in json.loads((out/'consumer-selection.json').read_text()):
 selector='^('+'|'.join(job['tests'])+')$' if job['tests'] else '^$'
 args=['go','test',job['package'],'-run',selector,'-count=1','-timeout','90s','-json','-parallel','4'];name=job['package'].removeprefix('./').replace('/','_');start=time.monotonic();path=out/('consumer-'+name+'.jsonl')
 with path.open('w') as log:
  try:code=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT,timeout=150).returncode
  except subprocess.TimeoutExpired:code=124
 tests=[]
 for l in path.read_text().splitlines():
  try:r=json.loads(l)
  except:continue
  if r.get('Test') and r['Action'] in ['pass','fail','skip']:tests.append(dict(test=r['Test'],action=r['Action'],seconds=r.get('Elapsed',0)))
 row=dict(**job,command=args,exit=code,seconds=round(time.monotonic()-start,3),results=tests,over_60s=[x for x in tests if x['action']!='skip' and x['seconds']>=60]);rows.append(row);(out/'consumer-results.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(row),flush=True)
