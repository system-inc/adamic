"""Group ordinary records and compare exact tuples to the recorded baseline."""
import collections,hashlib,json
from pathlib import Path
out=Path(__file__).resolve().parent
tree=Path('/workspace/scratch/next-stops-f44b5c27-adapted')
rows=[row for n in range(1,5) for row in json.loads((out/f'group-{n}-rows.json').read_text())]
assert len(rows)==79
assert len({r['roots'][0] for r in rows})==79
normalize=lambda s:s.split('/src/compiler/')[-1]
old=json.loads((out/'previous-grouped-first-stops.json').read_text())
baseline={root:(r['where'],r['kind'],r['reason']) for r in old for root in r['roots']}
groups=collections.defaultdict(list)
for row in rows:groups[(normalize(row.get('where','-')),row['kind'],row.get('reason','-'))].append(normalize(row['roots'][0]))
result=[]
for (where,kind,reason),roots in sorted(groups.items(),key=lambda x:(-len(x[1]),x[0])):
 prior=[baseline[r] for r in roots]
 unchanged=sum(p==(where,kind,reason) for p in prior)
 known=(where,kind,reason) in baseline.values()
 result.append(dict(where=where,kind=kind,reason=reason,roots=roots,count=len(roots),unchanged=unchanged,moved=(len(roots)-unchanged if known else 0),new=(0 if known else len(roots)),comparison='unchanged' if unchanged==len(roots) else 'moved' if known else 'new'))
(out/'grouped-first-stops.json').write_text(json.dumps(result,indent=2)+'\n')
comparisons=[]
for r in rows:
 root=normalize(r['roots'][0]);before=baseline[root];after=(normalize(r.get('where','-')),r['kind'],r.get('reason','-'))
 classification='unchanged' if before==after else 'moved' if after in baseline.values() else 'new'
 comparisons.append(dict(root=root,before=before,after=after,classification=classification))
(out/'comparison.json').write_text(json.dumps(comparisons,indent=2)+'\n')
(out/'census.jsonl').write_text(''.join(json.dumps(r)+'\n' for r in rows))
(out/'adapted-source-manifest.json').write_text(json.dumps({str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src/compiler').rglob('*')) if p.is_file()},indent=2)+'\n')
for r in result:print(r['count'],r['where'],r['kind'],r['comparison'],r['reason'])
