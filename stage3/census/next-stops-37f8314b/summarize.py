from pathlib import Path
from collections import defaultdict
import json
out=Path(__file__).resolve().parent
rows=[r for p in sorted(out.glob('group-*-rows.json')) for r in json.loads(p.read_text())]
assert len(rows)==79 and len({r['roots'][0] for r in rows})==79
norm=lambda s:s.split('/src/compiler/',1)[-1] if s else '-'
key=lambda r:(norm(r.get('where','')),r['kind'],r.get('reason','-'))
groups=defaultdict(list)
for r in rows:groups[key(r)].append(norm(r['roots'][0]))
summary=[dict(where=k[0],kind=k[1],reason=k[2],count=len(v),roots=sorted(v)) for k,v in sorted(groups.items(),key=lambda item:(-len(item[1]),item[0]))]
(out/'grouped-first-stops.json').write_text(json.dumps(summary,indent=2)+'\n')
old=json.loads((out/'previous-grouped-first-stops.json').read_text())
baseline={root:(g['where'],g['kind'],g['reason']) for g in old for root in g['roots']}
assert set(baseline)=={norm(r['roots'][0]) for r in rows}
comparison=defaultdict(list)
for r in rows:comparison[baseline[norm(r['roots'][0])],key(r)].append(norm(r['roots'][0]))
(out/'comparison.json').write_text(json.dumps([dict(before=a,after=b,status='unchanged' if a==b else 'moved',roots=sorted(v)) for (a,b),v in comparison.items()],indent=2)+'\n')
print(json.dumps([{k:v for k,v in g.items() if k!='roots'} for g in summary],indent=2))
