"""Group both freshly measured revisions; compare their exact first-stop tuples."""
import collections,hashlib,json
from pathlib import Path
base=Path(__file__).resolve().parent
tree=Path('/workspace/scratch/next-stops-37d96306/adapted')
norm=lambda s:s.split('/src/compiler/')[-1]
read=lambda p:json.loads(p.read_text())
allrows={}; allgroups={}
known={(g['where'],g['kind'],g['reason']):g for g in read(base/'reviewed-stops.json')} if (base/'reviewed-stops.json').exists() else {}
for revision in ['base','head']:
 out=base/revision/'CommonJSBundle'; rows=[]; groups=collections.defaultdict(list)
 for i in range(1,80):
  pair=[json.loads(l) for l in (out/f'root-{i:02}.jsonl').read_text().splitlines()]
  assert len(pair)==2 and {k:v for k,v in pair[0].items() if k!='seconds'}=={k:v for k,v in pair[1].items() if k!='seconds'}
  rows.append(pair[0])
 for row in rows:groups[(norm(row.get('where','-')),row['kind'],row.get('reason','-'))].append(norm(row['roots'][0]))
 result=[];evidence=[]
 for key,roots in sorted(groups.items(),key=lambda x:(-len(x[1]),x[0])):
  g=dict(where=key[0],kind=key[1],reason=key[2],roots=roots,count=len(roots))
  g.update({k:v for k,v in known.get(key,{}).items() if k not in ('roots','count')})
  result.append(g)
  e={k:v for k,v in g.items() if k not in ('roots','count')}
  if key[0]!='-':
   file,line,col=key[0].rsplit(':',2);lines=(tree/'src/compiler'/file).read_text().splitlines();n=int(line)
   e['adapted_line']=lines[n-1];e['adapted_excerpt']='\n'.join(str(i+1)+': '+lines[i] for i in range(max(0,n-4),min(len(lines),n+3)))
  evidence.append(e)
 (out/'grouped-first-stops.json').write_text(json.dumps(result,indent=2)+'\n')
 (out/'source-evidence.json').write_text(json.dumps(evidence,indent=2)+'\n')
 (out/'census.jsonl').write_text(''.join(json.dumps(r)+'\n' for r in rows))
 allrows[revision]=rows;allgroups[revision]=result
 for g in result:print(revision,g['count'],g['where'],g['kind'],g['reason'])
baseline={root:[g['where'],g['kind'],g['reason']] for g in allgroups['base'] for root in g['roots']}
comparisons=[]
for row,oldrow in zip(allrows['head'],allrows['base']):
 root=norm(row['roots'][0]);after=[norm(row.get('where','-')),row['kind'],row.get('reason','-')]
 comparisons.append(dict(root=root,before=baseline[root],after=after,unchanged=baseline[root]==after,observation_unchanged={k:v for k,v in row.items() if k!='seconds'}=={k:v for k,v in oldrow.items() if k!='seconds'}))
for name,data in [('comparison.json',comparisons),('baseline.json',baseline)]:
 (base/name).write_text(json.dumps(data,indent=2)+'\n')
manifest={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src/compiler').rglob('*')) if p.is_file()}
(base/'adapted-source-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
