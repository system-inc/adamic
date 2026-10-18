import json,os,subprocess,time,sys,gzip
from pathlib import Path
out=Path(__file__).resolve().parent
os.sched_setaffinity(0, sorted(os.sched_getaffinity(0))[:4])
label,packages,selection=sys.argv[1],sys.argv[2].split(','),sys.argv[3] if len(sys.argv)>3 else ''
for pkg in packages:
 name=label+'-'+pkg.replace('/','_').replace('.','')
 cmd=['go','test','-p','1','-count=1','-json','-parallel=4','-timeout',('90s' if label.startswith('guards') else '11m'), './'+pkg]
 if selection:cmd+=['-run',selection]
 if len(sys.argv)>4 and sys.argv[4]:cmd+=['-skip',sys.argv[4]]
 start=time.monotonic();log=out/(name+'.jsonl')
 with log.open('w') as f:
  try:r=subprocess.run(['timeout','720',*cmd],stdout=f,stderr=subprocess.STDOUT,timeout=725);code=r.returncode
  except subprocess.TimeoutExpired:code=124
 seconds=round(time.monotonic()-start,3);events=[];failed=[]
 for line in log.read_text().splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if e.get('Action') in ['pass','fail','skip'] and e.get('Test'):events.append({k:e[k] for k in ['Action','Package','Test','Elapsed'] if k in e})
  if e.get('Action')=='fail' and e.get('Test'):failed.append(e['Test'])
 (out/(name+'-seconds.json.txt')).write_text(json.dumps(events,indent=2)+'\n')
 with (out/'suites.jsonl.txt').open('a') as f:f.write(json.dumps(dict(label=label,package=pkg,command=cmd,exit=code,seconds=seconds,failed=failed))+'\n')
 print(pkg,code,seconds,'failed',failed[:20],flush=True)
 if code:break
