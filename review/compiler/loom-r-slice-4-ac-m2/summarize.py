from pathlib import Path
import json,gzip
out=Path(__file__).resolve().parent;events=[];logs=[]
for path in sorted(list((out/'logs').glob('*.jsonl'))+list(out.glob('consumer-*.jsonl'))):
 if 'mutant' in path.name:continue
 rows=[]
 for l in path.read_text(errors='replace').splitlines():
  try:r=json.loads(l)
  except:continue
  rows.append(r)
  if r.get('Test') and r['Action'] in ['pass','fail','skip']:
   events.append(dict(test=r['Test'],package=r.get('Package'),action=r['Action'],seconds=r.get('Elapsed',0),time=r['Time'],log=str(path.relative_to(out))))
 logs.append(dict(log=str(path.relative_to(out)),packages=[dict(package=r.get('Package'),action=r['Action'],seconds=r.get('Elapsed')) for r in rows if not r.get('Test') and r['Action'] in ['pass','fail']],failed_tests=[r['Test'] for r in rows if r.get('Test') and r['Action']=='fail']))
latest={}
for r in sorted(events,key=lambda x:x['time']):latest[r['package']+':'+r['test']]=r
failures=[r for r in latest.values() if r['action']=='fail'];slow=[r for r in events if r['action']!='skip' and r['seconds']>=60];result=dict(logs=logs,latest=list(latest.values()),unresolved_failures=failures,over_60s=slow,max_seconds=max([r['seconds'] for r in events] or [0]));(out/'validation-summary.json').write_text(json.dumps(result,separators=(',',':'))+'\n');print('latest tests',len(latest),'unresolved',failures,'over60',slow,'max',result['max_seconds'])
