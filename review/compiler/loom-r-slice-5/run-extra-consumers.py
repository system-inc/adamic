import pathlib,json,subprocess,time
out=pathlib.Path(__file__).resolve().parent
results=json.loads((out/'extra-consumer-results.json').read_text())
for i,row in enumerate(json.loads((out/'extra-consumer-selection.json').read_text())):
 if any(r['package']==row['package'] and r['exit']==0 for r in results):continue
 selector='^('+'|'.join(row['tests'])+')$' if row['tests'] else '^$'
 command=['go','test',row['package'],'-run',selector,'-count=1','-json','-timeout','90s'];start=time.monotonic();log=out/'logs'/('extra-consumer-%02d.jsonl'%i)
 if log.exists(): log.rename(log.with_suffix('.pre-owner.jsonl'))
 try:
  with log.open('w') as f:r=subprocess.run(command,stdout=f,stderr=subprocess.STDOUT,timeout=150)
  code=r.returncode
 except subprocess.TimeoutExpired:code=124
 events=[]
 for line in log.read_text().splitlines():
  try:event=json.loads(line)
  except ValueError:continue
  if event.get('Test') and event.get('Action') in ['pass','fail','skip']:events.append(event)
 result=dict(package=row['package'],command=command,exit=code,seconds=time.monotonic()-start,tests=events);results.append(result);(out/'extra-consumer-results.json').write_text(json.dumps(results,indent=2)+'\n');print(i,row['package'],code,round(result['seconds'],2),flush=True)
 if code:continue
 if row['tests'] and not set(row['tests']).issubset({e['Test'] for e in events}):raise SystemExit('omitted test')
 leaves=[e for e in events if not any(x['Test'].startswith(e['Test']+'/') for x in events)]
 if any(e.get('Elapsed',0)>=60 for e in leaves): print('red leaf; will not rerun',flush=True)
