"""Group measured records and retain exact per-root comparisons."""
import collections,hashlib,json,os
from pathlib import Path
base=Path(__file__).resolve().parent;mode=os.environ['CENSUS_MODULE_FORMAT'];out=base/mode
tree=Path('/workspace/scratch/next-stops-c4v7-6-adapted')
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

# Resolve only exact measured stop tuples from the pinned baseline metadata.
known={(g['where'],g['kind'],g['reason']):g for g in old}
for g in json.loads((base/'reviewed-stops.json').read_text()):known[(g['where'],g['kind'],g['reason'])]=g
evidence=[]
for g in result:
 key=(g['where'],g['kind'],g['reason'])
 assert key in known, 'New stop needs emitting-source review: '+str(key)
 for k in ['phase','owner']:g[k]=known[key][k]
 for k in ['emitter','owner_basis','expression_change_reach']:
  if k in known[key]:g[k]=known[key][k]
 e={k:g[k] for k in ['where','kind','reason','phase','owner','emitter','owner_basis','expression_change_reach'] if k in g}
 if g['where']!='-':
  file,line,col=g['where'].rsplit(':',2);lines=(tree/'src/compiler'/file).read_text().splitlines();n=int(line)
  e['adapted_line']=lines[n-1]
  e['adapted_excerpt']='\n'.join(str(i+1)+': '+lines[i] for i in range(max(0,n-4),min(len(lines),n+3)))
 evidence.append(e)
(out/'grouped-first-stops.json').write_text(json.dumps(result,indent=2)+'\n')
(out/'source-evidence.json').write_text(json.dumps(evidence,indent=2)+'\n')
