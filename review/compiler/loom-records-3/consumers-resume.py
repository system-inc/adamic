from pathlib import Path
import subprocess,time,json
out=Path(__file__).resolve().parent;rows=[]
jobs=json.loads((out/'consumer-selection.json').read_text()); previous=json.loads((out/'consumer-results.json').read_text())
for job in jobs[len(previous)-1:]:
 selector='^('+'|'.join(job['tests'])+')$' if job['tests'] else '^$'
 args=['go','test',job['package'],'-run',selector,'-count=1','-timeout','90s','-v','-parallel','4'];name=job['package'].removeprefix('./').replace('/','_');start=time.monotonic()
 with (out/('consumer-'+name+'-resume.log')).open('w') as log:
  try:code=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT,timeout=180).returncode
  except subprocess.TimeoutExpired:code=124
 row=dict(**job,command=args,exit=code,seconds=round(time.monotonic()-start,3));rows.append(row);(out/'consumer-resume-results.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(row),flush=True)
 if code:break
