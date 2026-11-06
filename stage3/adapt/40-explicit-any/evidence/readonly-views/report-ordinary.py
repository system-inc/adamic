"""Compare ordinary production compiler first stops for exactly 79 TypeScript roots."""
import collections,json,sys
from pathlib import Path
work=Path(sys.argv[1]);out=Path(__file__).resolve().parent
records={}
for label,name,tree in [('before','ordinary-before.jsonl','/workspace/scratch/ro-before'),('final','ordinary-final.jsonl','/workspace/scratch/ro-final')]:
 rows=[json.loads(x) for x in (work/name).read_text().splitlines()]
 selected=[r for r in rows if len(r['roots'])==1 and r['roots'][0].endswith('.ts')]
 assert len(selected)==79,(label,len(selected))
 def normalize(v):
  if isinstance(v,str):return v.replace(tree+'/','')
  if isinstance(v,list):return [normalize(x) for x in v]
  if isinstance(v,dict):return {k:normalize(x) for k,x in v.items() if k!='seconds'}
  return v
 records[label]={normalize(r['roots'][0]):normalize(r) for r in selected}
assert records['before'].keys()==records['final'].keys()
changed=[{'root':root,'before':a,'final':records['final'][root]} for root,a in records['before'].items() if a!=records['final'][root]]
result={'roots_each':79,'normalization':'remove tree prefix from all embedded strings; omit timing only','kinds':{label:dict(collections.Counter(r['kind'] for r in rows.values())) for label,rows in records.items()},'changed_first_stops':changed,'records':records}
(out/'ordinary-comparison.json').write_text(json.dumps(result,indent=2)+'\n')
lines=['# Ordinary compiler census','','All 79 compiler roots were measured independently with the ordinary production compiler. The driver also emits three non-TypeScript file records and one whole-program record; those are preserved raw and excluded from this 79-root comparison.','','| Kind | Before | Final |','| --- | ---: | ---: |']
for kind in sorted(set(result['kinds']['before'])|set(result['kinds']['final'])):lines.append(f"| {kind} | {result['kinds']['before'].get(kind,0)} | {result['kinds']['final'].get(kind,0)} |")
lines+=['',f'Changed first stops: {len(changed)}.']
for row in changed:lines+=['',json.dumps(row,indent=2)]
lines+=['','Complete per-root records, including exact locations, reasons and fixes, are in ordinary-comparison.json. Only tree prefixes and timing fields are normalized. The compiler code is unmodified from base 63e14a0b2b61cf251ad74292e42d5fd09af804b3.']
(out/'ORDINARY-CENSUS.md').write_text('\n'.join(lines)+'\n')
print(json.dumps({'roots_each':79,'kinds':result['kinds'],'changed_first_stops':changed},indent=2))
