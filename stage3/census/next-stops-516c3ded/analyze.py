"""Group measured records and retain exact per-root comparisons."""
import collections,hashlib,json,os
from pathlib import Path
base=Path(__file__).resolve().parent;mode=os.environ['CENSUS_MODULE_FORMAT'];out=base/mode
tree=Path('/workspace/scratch/next-stops-516c3ded-adapted')
norm=lambda s:s.split('/src/compiler/')[-1]
old=json.loads((base/('previous-'+mode+'-grouped.json')).read_text())
baseline={root:(g['where'],g['kind'],g['reason']) for g in old for root in g['roots']}
rows=[]
for i in range(1,80):
 pair=[json.loads(l) for l in (out/f'root-{i:02}.jsonl').read_text().splitlines()]
 assert len(pair)==2 and {k:v for k,v in pair[0].items() if k!='seconds'}=={k:v for k,v in pair[1].items() if k!='seconds'}
 rows.append(pair[0])
groups=collections.defaultdict(list);comparisons=[]
for row in rows:
 root=norm(row['roots'][0]);after=(norm(row.get('where','-')),row['kind'],row.get('reason','-'))
 groups[after].append(root);comparisons.append(dict(root=root,before=baseline[root],after=after,unchanged=baseline[root]==after))
result=[dict(where=k[0],kind=k[1],reason=k[2],roots=v,count=len(v)) for k,v in sorted(groups.items(),key=lambda x:(-len(x[1]),x[0]))]
for name,data in [('grouped-first-stops.json',result),('comparison.json',comparisons),('baseline.json',baseline)]:
 (out/name).write_text(json.dumps(data,indent=2)+'\n')
(out/'census.jsonl').write_text(''.join(json.dumps(r)+'\n' for r in rows))
(base/'adapted-source-manifest.json').write_text(json.dumps({str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src/compiler').rglob('*')) if p.is_file()},indent=2)+'\n')
for g in result:print(g['count'],g['where'],g['kind'],g['reason'])
