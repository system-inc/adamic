"""Quote every recorded site in the requested core declaration group."""
import json
from pathlib import Path
out=Path(__file__).resolve().parent

def esc(s):return str(s).replace('|','&#124;').replace('\n',' ')
def rows(p):return [json.loads(s) for s in p.read_text().splitlines()]
lines=['# Core declaration site ledger','']
for label in ['A','B']:
 if not (out/f'{label}-head-core.jsonl').exists():continue
 sites=json.loads((out/f'{label}-sites.json').read_text())['core']
 lines += [f'## Pairing {label}','', 'All kinds and phases are retained here, including recorder Boundary events. Repeated contexts share one row; raw streams retain them all.','', '| Owner | Location | Kind | Phase | Exact reason | Parent | Head |','| --- | --- | --- | --- | --- | ---: | ---: |']
 found={}
 statuses=[]
 for side in ['parent','head']:
  records=rows(out/f'{label}-{side}-core.jsonl')
  for record in records[1:]:
   for u in record.get('units',[]):
    if u.get('name') in ['orPredicate2','orPredicate3','or']:statuses.append(dict(side=side,**u))
   for f in record.get('findings',[]):
    if '/src/compiler/core.ts:' not in f['where']:continue
    line=int(f['where'].rsplit(':',2)[1]);owners=[s for s in sites if s['line']<=line<=s['endLine']]
    if not owners:continue
    owner=min(owners,key=lambda s:s['end']-s['start'])['name']
    key=(owner,f['where'].split('/src/',1)[-1],f['kind'],f['phase'],f['reason'])
    found.setdefault(key,dict(parent=0,head=0))[side]+=1
 for key,counts in sorted(found.items(),key=lambda i:(int(i[0][1].rsplit(':',2)[1]),int(i[0][1].rsplit(':',1)[1]),i[0][2:])):
  lines.append('| '+' | '.join(map(esc,[*key,counts['parent'],counts['head']]))+' |')
 if not found:lines+=['| No recorded sites | | | | | 0 | 0 |']
 lines+=['','Declaration unit statuses:','', '| Side | Name | Unit | Status | Instantiated |','| --- | --- | --- | --- | --- |']
 for u in statuses:lines.append('| '+' | '.join(map(esc,[u['side'],u['name'],u['where'].split('/src/',1)[-1],u['status'],u['instantiated']]))+' |')
 lines+=['']
(out/'core-sites.md').write_text('\n'.join(lines)+'\n')
