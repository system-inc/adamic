#!/usr/bin/env python3
import gzip,json,re
from pathlib import Path
out=Path(__file__).resolve().parent
rs={s:[json.loads(x) for x in gzip.open(out/f'{s}.jsonl.gz','rt')] for s in 'AB'}
sites={s:{tuple(f[k] for k in ('kind','where','reason','text')) for r in records[1:] for f in (r.get('findings') or [])} for s,records in rs.items()}
rows=[]
for reason in sorted({t[2] for ss in sites.values() for t in ss if t[0] in ('NotYet','Refused') and re.search(r'view|branded supertype|predicate',t[2],re.I)}):
 for kind in ('NotYet','Refused'):
  a={t for t in sites['A'] if t[0]==kind and t[2]==reason};b={t for t in sites['B'] if t[0]==kind and t[2]==reason}
  if not(a or b):continue
  rows.append({'kind':kind,'reason':reason,'A':len(a),'B':len(b),'A_examples':sorted({t[1].replace('/workspace/scratch/tree/','') for t in a})[:5],'B_examples':sorted({t[1].replace('/workspace/scratch/tree/','') for t in b})[:5]})
(out/'view-predicate-rows.json').write_text(json.dumps(rows,indent=2)+'\n')
lines=['# All view, branded-supertype and predicate reason rows','','Includes unchanged rows. Counts use the meter distinct-site key.','','| Kind and reason | A | B | A example locations | B example locations |','|---|---:|---:|---|---|']
for r in rows:
 vals=[r['kind']+': '+r['reason'],str(r['A']),str(r['B']),'; '.join(r['A_examples']),'; '.join(r['B_examples'])]
 lines.append('| '+' | '.join(v.replace('|','&#124;').replace('\n',' ') for v in vals)+' |')
(out/'view-predicate-rows.md').write_text('\n'.join(lines)+'\n')
