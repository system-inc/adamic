import json,pathlib,subprocess,time
out=pathlib.Path(__file__).resolve().parent
selections=json.loads((out/'focused-selection.json').read_text());results=[]
for package,names in selections.items():
 groups=[names[i:i+4] for i in range(0,len(names),4)] if names else [[]]
 for index,group in enumerate(groups):
  selector='^('+'|'.join(group)+')$' if group else '^$'
  command=['go','test',package,'-run',selector,'-count=1','-json','-timeout','90s']
  path=out/'logs'/(package.replace('/','_')+'-%02d.jsonl'%index)
  start=time.monotonic()
  with path.open('w') as log:result=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=180)
  events=[]
  for line in path.read_text().splitlines():
   try:r=json.loads(line)
   except ValueError:continue
   if r.get('Test') and r.get('Action') in ['pass','fail','skip']:events.append(r)
  record=dict(package=package,tests=group,command=command,seconds=time.monotonic()-start,exit=result.returncode,leaves=events,log=str(path))
  results.append(record);(out/'focused-results.json').write_text(json.dumps(results,indent=2)+'\n')
  print(package,index,result.returncode,round(record['seconds'],3),flush=True)
  if result.returncode:raise SystemExit(1)
  if any(r.get('Elapsed',0)>=60 for r in events):raise SystemExit('red leaf exceeded 60 seconds; do not rerun')
