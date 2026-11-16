"""Group paired ordinary measurements against the archived fb8db081 tables."""
import collections,hashlib,json,os
from pathlib import Path
base=Path(__file__).resolve().parent
tree=Path('/workspace/scratch/next-stops-82c78b6b/adapted')
norm=lambda s:s.split('/src/compiler/')[-1]
read=lambda p:json.loads(p.read_text())
baselines={};known={}
for mode in ['CommonJSBundle','EcmaScript']:
 old=read(base/'base'/mode/'grouped-first-stops.json')
 baselines[mode]={r:[g['where'],g['kind'],g['reason']] for g in old for r in g['roots']}
 known.update({(g['where'],g['kind'],g['reason']):g for g in old})
(base/'baseline.json').write_text(json.dumps(baselines,indent=2)+'\n')
if (base/'reviewed-stops.json').exists():
 known.update({(g['where'],g['kind'],g['reason']):g for g in read(base/'reviewed-stops.json')})
comparisons=[]
for mode in os.environ.get('CENSUS_ANALYZE_FORMATS','CommonJSBundle EcmaScript').split():
 baseline=baselines[mode]
 out=base/'head'/mode; rows=[];groups=collections.defaultdict(list)
 for i in range(1,80):
  pair=[json.loads(l) for l in (out/f'root-{i:02}.jsonl').read_text().splitlines()]
  assert len(pair)==2 and {k:v for k,v in pair[0].items() if k!='seconds'}=={k:v for k,v in pair[1].items() if k!='seconds'}
  rows.append(pair[0])
 for r in rows: groups[(norm(r.get('where','-')),r['kind'],r.get('reason','-'))].append(norm(r['roots'][0]))
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
 for r in rows:
  root=norm(r['roots'][0]);after=[norm(r.get('where','-')),r['kind'],r.get('reason','-')]
  comparisons.append(dict(module_format=mode,root=root,before=baseline[root],after=after,unchanged=baseline[root]==after))
 for g in result: print(mode,g['count'],g['where'],g['kind'],g['reason'])
(base/'comparison.json').write_text(json.dumps(comparisons,indent=2)+'\n')
manifest={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src/compiler').rglob('*')) if p.is_file()}
(base/'adapted-source-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
