"""Pair sites by lexical declaration identity, kind, exact reason, phase, then source order."""
import json,collections
from pathlib import Path
out=Path(__file__).resolve().parent
old=json.loads((out/'previous-rows.json').read_text());new=json.loads((out/'rows.json').read_text())
def identity(data,f):
 units={u['id']:u for u in data['units']};u=units.get(f['site_unit_id']);return (u['source_key'] if u else f['site_unit_name'],f['kind'],f['reason'],f['phase'])
pending=collections.defaultdict(list)
for f in new['sites']:pending[identity(new,f)].append(f)
remain=[];gone=[]
for f in old['sites']:
 candidates=pending[identity(old,f)]
 if candidates:remain.append({'old':f,'new':candidates.pop(0)})
 else:gone.append(f)
added=sorted([f for fs in pending.values() for f in fs],key=lambda f:tuple(map(int,f['where'].rsplit(':',2)[1:])))
result={'method':'Lexical source_key, kind, exact reason and phase; repeated matching sites paired in source order. Locations are adapted coordinates in their own snapshots. New includes newly surveyed sites before getOwnValues.','gone':gone,'remain':remain,'new':added}
(out/'comparison.json').write_text(json.dumps(result,indent=2)+'\n');print('gone',len(gone),'remain',len(remain),'new',len(added))
