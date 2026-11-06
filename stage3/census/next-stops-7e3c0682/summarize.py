from pathlib import Path
from collections import defaultdict
import json
repo = Path('/workspace/adamic')
out = repo/'stage3/census/next-stops-7e3c0682'
rows = [row for p in sorted(out.glob('group-*-rows.json')) for row in json.loads(p.read_text())]
assert len(rows)==79 and len({row['roots'][0] for row in rows})==79

def norm(s):
 return s.split('/src/compiler/',1)[-1] if s else '-'
def key(row):
 return norm(row.get('where','')),row['kind'],row.get('reason','-')
groups=defaultdict(list)
for row in rows: groups[key(row)].append(norm(row['roots'][0]))
old=json.loads((repo/'stage3/adapt/99-correctness-axolotl-w/census-results.json').read_text())['rows']
baseline={norm(row['roots'][0]):key(row) for row in old}
assert set(baseline)=={norm(row['roots'][0]) for row in rows}
ordered=sorted(groups,key=lambda k:(-len(groups[k]),k))
summary=[dict(where=k[0],kind=k[1],reason=k[2],count=len(groups[k]),roots=sorted(groups[k])) for k in ordered]
(out/'grouped-first-stops.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary,indent=2))
comparison=defaultdict(list)
for row in rows:
 root=norm(row['roots'][0]); previous=baseline[root]; current=key(row)
 comparison[previous,current].append(root)
(out/'comparison.json').write_text(json.dumps([{'before':a,'after':b,'status':'unchanged' if a==b else 'moved','roots':sorted(names)} for (a,b),names in comparison.items()],indent=2)+'\n')
