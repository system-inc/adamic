"""Compact, independently readable summary of the paired JSON artifacts."""
import json
from pathlib import Path
out=Path(__file__).resolve().parent
allrows={}
for label in ['A','B']:
 result={}
 for mode in ['entry','core']:
  p=out/f'{label}-{mode}-pair.json'
  if not p.exists():continue
  d=json.loads(p.read_text());summary={}
  for view in ['raw_findings','distinct_reads']:
   r=d[view]
   summary[view]={k:len(r[k]) for k in ['lowered','moved','new_findings','new_sites','unit_first_stop_delta','regressions']}
   summary[view]['totals']=r['totals']
  summary['units']=d['raw_findings']['unit_first_stop_delta']
  result[mode]=summary
 allrows[label]=result
(out/'summary.json').write_text(json.dumps(allrows,indent=2)+'\n')
for label,result in allrows.items():
 for mode,rows in result.items():
  print(label,mode)
  for v in ['raw_findings','distinct_reads']:print(v,{k:x for k,x in rows[v].items() if k!='totals'})
  for u in rows['units']:
   print(u['unit'].get('name'),u['unit']['where'],'OLD',u['base'] and u['base']['reason'],'NEW',u['head'] and u['head']['reason'])
