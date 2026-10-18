"""Inventory every current test and retain exact final status and budget evidence."""
from pathlib import Path
import json
out=Path(__file__).parent
names=['core-final','javascript-fuzz-final','reader-guard-final','native-final','oracle-final','native-wasi-after-suite','normalization-product-recovered','moves-product-recovered','cycleledger-final']
paths=[out/'logs'/(name+'.jsonl')for name in names]+list((out/'logs').glob('batch-*.jsonl'))
events=[]
for path in paths:
 for line in path.read_text().splitlines():
  try:event=json.loads(line)
  except json.JSONDecodeError:continue
  if event.get('Action')in['pass','fail','skip']and 'Test'in event and '/'not in event['Test']:
   event['log']=str(path.relative_to(out));events.append(event)
latest={}
for event in sorted(events,key=lambda event:event.get('Time','')):latest[event['Package'],event['Test']]=event
inventory=json.loads((out/'test-inventory.json').read_text());report={}
for package,tests in inventory.items():
 missing=[name for name in tests if(package,name)not in latest]
 failed=[{'test':name,'log':latest[package,name]['log']}for name in tests if(package,name)in latest and latest[package,name]['Action']=='fail']
 durations=[{'test':name,'seconds':latest[package,name].get('Elapsed',0),'log':latest[package,name]['log']}for name in tests if(package,name)in latest]
 over=[event for event in durations if event['seconds']>=60]
 report[package]={'total':len(tests),'passed':sum(latest.get((package,name),{}).get('Action')=='pass'for name in tests),'skipped':sum(latest.get((package,name),{}).get('Action')=='skip'for name in tests),'missing':missing,'failed':failed,'over60':over,'tests':[{'name':name,'status':latest[package,name]['Action'],'seconds':latest[package,name].get('Elapsed',0),'log':latest[package,name]['log']}for name in tests if(package,name)in latest]}
 assert not missing and not failed,(package,missing,failed)
 assert all(event['test']=='TestMovesAgreesWithNode_asan'for event in over),over
(out/'validation.json').write_text(json.dumps(report,indent=2)+'\n')
print('all',sum(row['total']for row in report.values()),'tests accounted for; no missing or failed tests')
print('budget exceptions:',[event for row in report.values()for event in row['over60']])
