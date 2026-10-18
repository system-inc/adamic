import pathlib, subprocess, json, re, time
out=pathlib.Path(__file__).resolve().parent
paths=json.loads((out/'fixture-selection.json').read_text())
results=[]
groups={}
for path in paths: groups.setdefault(str(pathlib.Path(path).parent),[]).append(path)
batches=[names[i:i+8] for names in groups.values() for i in range(0,len(names),8)]
for i,group in enumerate(batches):
 prefix='/'.join('^'+re.escape(part)+'$' for part in pathlib.Path(group[0]).parent.parts)
 selector='^TestNativeAgreesWithNode_[0-9]+$/'+prefix+'/^('+ '|'.join(re.escape(pathlib.Path(p).name) for p in group)+')$'
 command=['go','test','./internal/oracle','-run',selector,'-count=1','-json','-timeout','90s']
 log=out/'logs'/('fixtures-%02d.jsonl'%i)
 start=time.monotonic()
 with log.open('w') as stream: result=subprocess.run(command,stdout=stream,stderr=subprocess.STDOUT,timeout=150)
 leaves=[]
 for line in log.read_text().splitlines():
  try:r=json.loads(line)
  except ValueError:continue
  if r.get('Test') and r.get('Action') in ['pass','fail','skip']:leaves.append(r)
 actual=[r for r in leaves if '/' in r['Test']]
 record=dict(command=command,paths=group,exit=result.returncode,seconds=time.monotonic()-start,leaves=leaves)
 results.append(record);(out/'fixture-results.json').write_text(json.dumps(results,indent=2)+'\n')
 print(i,result.returncode,len(actual),round(record['seconds'],3),flush=True)
 if result.returncode or len(actual)!=len(group):raise SystemExit('fixture selection failed or omitted a program')
 if any(r.get('Elapsed',0)>=60 for r in leaves):raise SystemExit('red test leaf exceeded 60 seconds; do not rerun')
