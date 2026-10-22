import json,os,subprocess,time
from pathlib import Path
e=Path(__file__).resolve().parent
old=Path('review/compiler/loom-unions-3');checks=json.loads((old/'checks.json').read_text())
labels=['lower-focus','oracle-focus','fixtures','regex-adapter','stage1-gaps','flow-consumers','fresh-consumers','native-consumers','guards-ir','guards-runtime','guards-registration','guards-parallel','readers','vet']
selected=[]
for label in labels:
 selected.append(next(r for r in reversed(checks) if r['label']==label))
results=[]
for row in selected:
 cmd=row['command'];start=time.monotonic();log=e/(row['label']+'.log')
 with log.open('w') as out:
  try:code=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT,timeout=180).returncode
  except subprocess.TimeoutExpired:code=124
 events=[]
 for line in log.read_text().splitlines():
  try:r=json.loads(line)
  except json.JSONDecodeError:continue
  if r.get('Action') in ['pass','fail','skip'] and 'Test' in r:events.append(r)
 result=dict(label=row['label'],command=cmd,seconds=round(time.monotonic()-start,3),exit=code,events=events,overgrain=[r for r in events if r.get('Elapsed',0)>=60]);results.append(result);(e/'focused-results.json').write_text(json.dumps(results,indent=2)+'\n');print(row['label'],code,result['seconds'],len(events),flush=True)
