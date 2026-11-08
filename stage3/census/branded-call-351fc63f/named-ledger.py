"""Attribute each named source call to its full-mode failed statement, if any."""
import json
from pathlib import Path
out=Path(__file__).resolve().parent

def ready(name):
 p=out/(name+'.status.json')
 return p.exists() and json.loads(p.read_text())['exit']==0

def offset(where):
 file,line,column=where.rsplit(':',2);line,column=int(line),int(column)
 lines=Path(file).read_bytes().decode('utf8').splitlines(keepends=True)
 prefix='';units=0
 for ch in lines[line-1]:
  if units>=column-1:break
  prefix+=ch;units+=len(ch.encode('utf-16-le'))//2
 assert units==column-1
 return sum(len(s.encode('utf8')) for s in lines[:line-1])+len(prefix.encode('utf8'))

def observation(name,site):
 if not ready(name):return dict(state='pending',run=name)
 rows=[json.loads(s) for s in (out/(name+'.jsonl')).read_text().splitlines()]
 record=rows[1];u=record['units'][0];pos=offset(site)
 candidates=[f for f in record['findings'] if f['kind']=='Boundary' and f.get('start',0)<=pos<f.get('end',0)]
 if candidates:
  b=min(candidates,key=lambda f:f['end']-f['start'])
  matches=[f for f in record['findings'] if f['kind']!='Boundary' and f['text']==b['text']]
  assert matches,(name,b)
  f=matches[0]
  state='unit prologue' if b['reason'].startswith('function signature or prologue') else 'enclosing statement'
  return dict(state=state,run=name,unit=u,finding=f,boundary=b)
 return dict(state='not reached' if u['status']!='deferred_generic_instantiation' else 'deferred generic',run=name,unit=u,first_finding=next((f for f in record['findings'] if f['kind']!='Boundary'),None))

def esc(x):return str(x).replace('|','&#124;').replace('\n',' ')
def text(o):
 if o['state']=='pending':return 'pending'
 f=o.get('finding') or o.get('first_finding')
 why=f"{f['kind']}: {f['reason']}" if f else 'no finding recorded'
 return o['state']+': '+why

allrows=[];lines=['# Named caller observations','',
 'These use the standard full-mode replay selector. A failed statement can contain more than one call; the ledger identifies the enclosing statement stop, not independent execution of every argument. Unit-prologue failures and unreached/deferred bodies are stated explicitly.','']
for label in ['A','B']:
 data=json.loads((out/f'{label}-sites.json').read_text())['checker']
 unique={(r['unit'] if label=='A' else r['observing_unit']):r for r in reversed(data)}
 indexes={u:i for i,u in enumerate(sorted(unique),1)}
 lines += [f'## Pairing {label}','','| Requested checker line | Observed site | Full-mode unit | Parent | Head |','| ---: | --- | --- | --- | --- |']
 for row in data:
  if label=='A' and row['requested_line']==10911:
   root=json.loads((out/'A-observing-sites.json').read_text())['checker'];r=next(x for x in root if x['requested_line']==10911)
   unit=r['observing_unit'];suffix='serializer'
  else:
   unit=row['unit'] if label=='A' else row['observing_unit'];suffix='focus-'+str(indexes[unit])
  before=observation(f'{label}-parent-{suffix}',row['site']);after=observation(f'{label}-head-{suffix}',row['site'])
  result=dict(pairing=label,requested_line=row['requested_line'],site=row['site'],observing_unit=unit,owners=row['owners'],parent=before,head=after)
  allrows.append(result)
  unit_name=(before.get('unit') or after.get('unit') or {}).get('name',row['owners'][-1])
  values=[row['requested_line'],row['site'].split('/src/',1)[-1],unit_name+' ('+unit.split('/src/',1)[-1]+')',text(before),text(after)]
  lines.append('| '+' | '.join(map(esc,values))+' |')
 lines+=['']
(out/'named-sites.json').write_text(json.dumps(allrows,indent=2)+'\n')
(out/'named-sites.md').write_text('\n'.join(lines)+'\n')
