import json,time,subprocess
from pathlib import Path
out=Path(__file__).resolve().parent
for side in ['base','head']:
 p=out/(side+'-entry-run-relations.jsonl')
 if not p.exists():print(side,'not started');continue
 rows=[]
 for line in p.read_text().splitlines():
  try:rows.append(json.loads(line))
  except json.JSONDecodeError:pass
 units=[r for r in rows if r['where']!='program']
 marked=[(i,r) for i,r in enumerate(units,1) if r['answer']]
 print(side,'probed='+str(len(units)),'last='+str(units[-1]['where'] if units else 'load'),'first_marked_attempt='+str(marked[0][0] if marked else 'none'),'last_update_age='+str(round(time.time()-p.stat().st_mtime,1)))
for p in sorted(out.glob('*entry-run.status.json')):print(p.name,p.read_text())
