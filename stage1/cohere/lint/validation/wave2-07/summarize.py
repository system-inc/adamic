import json,pathlib,datetime,collections
root=pathlib.Path('/tmp/wave2-07-gate');meta=json.loads((root/'metadata.json').read_text());results=[]
def date(s):return datetime.datetime.fromisoformat(s.replace('Z','+00:00'))
for log in sorted(root.glob('stage1-*.jsonl')):
 events=[]
 for line in log.open():
  try:events.append(json.loads(line))
  except:pass
 terminals=[e for e in events if e['Action'] in ['pass','fail','skip'] and 'Test' in e]
 tops=[e for e in terminals if '/' not in e['Test']];first=next((e for e in events if e['Action']=='start'),None)
 cont=next((e for e in events if e['Action']=='cont' and '/' not in e.get('Test','/')),None)
 counts=collections.Counter(e['Action'] for e in terminals);topcounts=collections.Counter(e['Action'] for e in tops)
 row={'log':str(log),'package':events[0]['Package'] if events else None,'counts':dict(counts),'top_counts':dict(topcounts),'top_tests':sorted([{'test':e['Test'],'result':e['Action'],'seconds':e.get('Elapsed',0)} for e in tops],key=lambda t:-t['seconds']),'skips':[e['Test'] for e in terminals if e['Action']=='skip'],'failures':[e['Test'] for e in terminals if e['Action']=='fail'],'serial_seconds':(date(cont['Time'])-date(first['Time'])).total_seconds() if first and cont else None,'mutant_catches':[]}
 row['started_top_tests']=sorted({e['Test'] for e in events if e['Action']=='run' and '/' not in e.get('Test','/')})
 row['unfinished_top_tests']=sorted(set(row['started_top_tests'])-{e['Test'] for e in tops})
 if cont and first:
  serial=[e for e in tops if date(e['Time'])<=date(cont['Time'])]
  if serial:
   last=max(serial,key=lambda e:date(e['Time']));row['last_serial_test']=last['Test'];row['last_serial_finish_seconds']=(date(last['Time'])-date(first['Time'])).total_seconds()
 for e in events:
  s=e.get('Output','')
  if ('caught' in s or 'canary' in s or 'survived' in s) and e.get('Test'):row['mutant_catches'].append({'test':e['Test'],'message':s.strip()})
 results.append(row)
(root/'summary.json').write_text(json.dumps({'metadata':meta,'results':results},indent=2))
for r in results:print(r['package'],r['counts'],'serial',r['serial_seconds'],'fail',r['failures'],'skip',r['skips'])
