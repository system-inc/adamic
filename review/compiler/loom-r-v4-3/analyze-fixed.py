import collections,csv,json,re
from pathlib import Path
e=Path(__file__).resolve().parent/"fixed";alltests={};summary={}
for side in ['base','v4']:
 p=e/(side+'-package.jsonl')
 if not p.exists():continue
 tests={};terminal=None
 for line in p.read_text(errors='replace').splitlines():
  try:r=json.loads(line)
  except ValueError:continue
  if name:=r.get('Test'):
   t=tests.setdefault(name,dict(test=name,output=[],start=None,cont=None,status='unfinished',seconds=None))
   if r['Action']=='run':t['start']=r['Time']
   if r['Action']=='cont':t['cont']=r['Time']
   if r['Action']=='output':t['output'].append(r.get('Output',''))
   if r['Action'] in ['pass','fail','skip']:t['status']=r['Action'];t['seconds']=r.get('Elapsed')
  elif r['Action'] in ['pass','fail']:terminal=r
 for name,t in tests.items():
  output=''.join(t['output']);t['output']=output;t['budget_overrun']=t['seconds'] is not None and t['seconds']>=60
  t['shard']=bool(re.fullmatch('TestPortMatchesGoCohere_[0-9]{4}',name))
  t['kind']='pass' if t['status']=='pass' else 'skip' if t['status']=='skip' else 'unfinished'
  if t['status']=='fail':
   t['kind']='timeout' if re.search('exceeded.*deadline|timed out|deadline exceeded',output,re.I) else 'assertion/setup'
  t['sides']=dict(re.findall(r'(Go|Node|release|JavaScript backend|ASan/UBSan/LeakSanitizer) ([0-9.]+)s',output))
  if t['sides']:t['sides']={k:float(v) for k,v in t['sides'].items()}
 alltests[side]=tests
 problems=[t for t in tests.values() if t['status']=='fail' or t['budget_overrun']]
 summary[side]=dict(status=terminal['Action'] if terminal else 'running',package_seconds=terminal.get('Elapsed') if terminal else None,completed=collections.Counter(t['status'] for t in tests.values()),problems=[{k:v for k,v in t.items() if k!='output'} for t in problems])
 (e/(side+'-tests.json')).write_text(json.dumps(tests,indent=2)+'\n')
(e/'package-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
with (e/'shard-times.csv').open('w') as out:
 w=csv.writer(out);w.writerow(['test','base_status','base_seconds','v4_status','v4_seconds','crossed_60_on_v4'])
 for name in sorted(set.union(*(set(ts) for ts in alltests.values())) if alltests else []):
  b=alltests.get('base',{}).get(name,{});v=alltests.get('v4',{}).get(name,{});w.writerow([name,b.get('status'),b.get('seconds'),v.get('status'),v.get('seconds'),b.get('seconds') is not None and b['seconds']<60 and v.get('seconds') is not None and v['seconds']>=60])
print(json.dumps(summary,indent=2))
