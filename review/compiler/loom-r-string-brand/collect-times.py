import json
from pathlib import Path
out=Path(__file__).resolve().parent;checks=json.loads((out/'checks.json').read_text());events=[];mutants=[]
for p in sorted(out.glob('*.log')):
 rows=[];errors=[]
 for line in p.read_text().splitlines():
  try:r=json.loads(line)
  except ValueError:continue
  if not isinstance(r,dict):continue
  if 'Test' in r and r.get('Action') in ['pass','fail']:
   e=dict(log=p.name,test=r['Test'],seconds=r.get('Elapsed',0),result=r['Action']);rows.append(e);events.append(e)
  if r.get('OutputType')=='error':errors.append(r.get('Output',''))
 if p.name.startswith('mutant-'):mutants.append(dict(topic=p.stem.removeprefix('mutant-'),catchers=[r for r in rows if r['result']=='fail'],errors=errors))
assert len(mutants)==5 and all(r['catchers'] for r in mutants)
(out/'test-times.json').write_text(json.dumps(events,indent=2));(out/'mutants/results.json').write_text(json.dumps(mutants,indent=2));(out/'test-times.tsv').write_text('log\ttest\tseconds\tresult\n'+''.join(f"{r['log']}\t{r['test']}\t{r['seconds']}\t{r['result']}\n" for r in events))

assert all(e['seconds'] < 60 for e in events)
print('All observed tests below 60 seconds; maximum:',max(e['seconds'] for e in events))
