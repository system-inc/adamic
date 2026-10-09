from pathlib import Path
import subprocess,time,json,statistics,os
root=Path(__file__).resolve().parents[3]
out=Path(__file__).resolve().parent
names=[x.strip() for x in Path('/tmp/u007-test-list.log').read_text().splitlines() if x.startswith('Test')]
records=[]
for name in names:
 for trial in range(1,4):
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic/','-run','^'+name+'$']
  log=out/(name+'-timing-'+str(trial)+'.log')
  start=time.monotonic()
  with log.open('w') as stream: result=subprocess.run(cmd,cwd=root,stdout=stream,stderr=subprocess.STDOUT)
  events=[]
  for line in log.read_text().splitlines():
   try: events.append(json.loads(line))
   except ValueError: pass
  pkg=next((x for x in reversed(events) if x.get('Action') in ('pass','fail') and 'Test' not in x),{})
  record=dict(test=name,trial=trial,command=' '.join(cmd),wall_seconds=time.monotonic()-start,binary_seconds=pkg.get('Elapsed'),exit=result.returncode,log=str(log.relative_to(root)),events=[x for x in events if x.get('Action') in ('pass','fail','skip') and '/' not in x.get('Test','')])
  records.append(record)
  (out/'timings.json').write_text(json.dumps(records,indent=2)+'\n')
  print(name,trial,record['binary_seconds'],record['exit'],flush=True)
  if result.returncode: raise SystemExit('red timing run')
