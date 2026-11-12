import json
from pathlib import Path
D=Path(__file__).resolve().parent/'raw';done={};stats={};errors=[]
for p in D.glob('group-*.json'):
 r=json.loads(p.read_text());done[p.stem.removeprefix('group-')]={'exit':r['child_exit'],'timed_out':r['timed_out'],'seconds':round(r['elapsed_seconds'],1)}
for p in D.glob('group-*-records.jsonl'):
 rows=[]
 for s in p.read_text().splitlines():
  try:rows.append(json.loads(s))
  except json.JSONDecodeError:pass
 stats[p.name.split('-records')[0].removeprefix('group-')]={'units':len(rows),'view_stops':sum(r.get('view_stop',False) for r in rows),'walk_calls':sum(r.get('real_host_walk_calls',0) for r in rows)};errors.extend(r for r in rows if r['kind']=='BoundaryProbeError')
print(json.dumps({'done':done,'stats':stats,'errors':errors}))
