from pathlib import Path
import json,subprocess,time,re
out=Path(__file__).resolve().parent;results=json.loads((out/'results-initial.json').read_text())
jobs=[('counts',['go','test','./internal/oracle','-run','TestCountsAreRecorded','-count=1','-json','-timeout','600s','-parallel','4','-args','-update-counts'],660),('ir-prepare',['go','test','./internal/ir','-run','^TestCallTargetExportInputs','-count=1','-json','-timeout','90s','-parallel','4'],120)]
for pkg,limit in [('internal/lower',600),('internal/ir',180),('internal/flow',480),('internal/fresh',480),('internal/javascript',180),('internal/load',300),('stage3/project-loader/probe',90),('stage3/fixtures',600),('internal/native',840),('internal/oracle',1080)]:
 jobs.append((pkg.replace('/','_'),['go','test','./'+pkg+'/...','-count=1','-json','-timeout',str(limit)+'s','-parallel','4','-p','1'],limit+60))
jobs.append(('guards',['go','test','./cmd/adamic-gate','-run','^(TestEveryTestIsParallelOrSaysWhy|TestEveryDuplicatedProductDeclarationAgrees)$','-count=1','-json','-timeout','90s','-parallel','4'],120))
for label,command,limit in jobs:
 if label=='counts':continue # The one completed regeneration is retained.
 if label=='ir-prepare':command[4]='^TestCallTargetExportInputs(Flow|Oracle)$'
 print('START '+label,flush=True);start=time.monotonic();log=out/'logs'/(label+'.jsonl')
 with log.open('w') as f:
  try:run=subprocess.run(command,stdout=f,stderr=subprocess.STDOUT,timeout=limit);code=run.returncode
  except subprocess.TimeoutExpired:code=124
 events=[]
 for line in log.read_text(errors='replace').splitlines():
  try:x=json.loads(line)
  except:continue
  if x.get('Test') and x['Action'] in ['pass','fail','skip']:events.append({'test':x['Test'],'package':x.get('Package'),'action':x['Action'],'seconds':x.get('Elapsed',0)})
 row={'label':label,'command':command,'exit':code,'seconds':round(time.monotonic()-start,3),'passed':sum(x['action']=='pass' for x in events),'skipped':sum(x['action']=='skip' for x in events),'failures':[x for x in events if x['action']=='fail'],'over_60s':[x for x in events if x['action']!='skip' and x['seconds']>=60]};results.append(row);(out/'results.json').write_text(json.dumps(results,indent=2)+'\n');(out/(label+'-tests.json')).write_text(json.dumps(events,separators=(',',':'))+'\n');print(json.dumps(row),flush=True)
