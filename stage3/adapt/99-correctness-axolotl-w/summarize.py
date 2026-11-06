"""Summarize only independent compiler .ts roots, retaining raw census rows."""
import collections,json,sys
from pathlib import Path
rows=[json.loads(line) for line in Path(sys.argv[1]).read_text().splitlines()]
roots=[r for r in rows if len(r['roots'])==1 and r['roots'][0].endswith('.ts')]
assert len(roots)==79,len(roots)
groups=collections.defaultdict(list)
for r in roots:
 where=r.get('where','').split('/src/compiler/')[-1]
 groups[(r['kind'],where,r.get('reason',''))].append(Path(r['roots'][0]).name)
result={'roots':len(roots),'groups':[dict(count=len(names),kind=k[0],where=k[1],reason=k[2],roots=names) for k,names in sorted(groups.items(),key=lambda item:-len(item[1]))],'rows':roots}
Path(sys.argv[2]).write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result['groups'],indent=2))
