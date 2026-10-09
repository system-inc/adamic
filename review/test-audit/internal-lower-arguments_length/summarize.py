import json,pathlib,statistics
P=pathlib.Path(__file__).parent
rows=json.loads((P/'rows.json').read_text()); selected={r['test'] for r in rows}
def events(log):
 out=[]
 for line in (P/'logs'/log).read_text().splitlines():
  try: e=json.loads(line)
  except ValueError: continue
  out.append(e)
 return out
runs=json.loads((P/'matrix-runs.json').read_text()) if (P/'matrix-runs.json').exists() else {}
for id,run in runs.items():
 fails=sorted({e['Test'].split('/')[0] for e in events(run['log']) if e['Action']=='fail' and 'Test' in e})
 if 'individual' in run:
  fails=[x for x in fails if x not in selected]+[name for name,r in run['individual'].items() if r['exit']==1]
 print(id,'unit:',','.join(x for x in fails if x in selected),'all:',','.join(fails))
