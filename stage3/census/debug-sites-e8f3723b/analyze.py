import json,re,collections
from pathlib import Path
out=Path(__file__).resolve().parent
ast=json.loads((out/'ast-map.json').read_text())
def lc(w):return tuple(map(int,re.search(r':(\d+):(\d+)$',w).groups()))
def owner(w):
 p=lc(w); candidates=[n for n in ast if (n['line'],n['column'])<=p<(n['end_line'],n['end_column'])]
 return min(candidates,key=lambda n:n['end']-n['start']) if candidates else None
records={}
for mode in ['first-error','full']:
 path=out/(mode+'.jsonl')
 if not path.exists():continue
 stream=[json.loads(l) for l in path.read_text().splitlines()]; files=[r for r in stream if 'file' in r]
 assert len(files)==1 and files[0]['file'].endswith('/debug.ts')
 records[mode]={'measurement':stream[0],'record':files[0]}
groups={}
for mode,data in records.items():
 for f in data['record']['findings']:
  if f['kind']=='Boundary':continue
  key=tuple(f[k] for k in ['where','kind','reason','phase'])
  if key not in groups:groups[key]=dict(f,observations=[],owner=owner(f['where']))
  groups[key]['observations'].append({'mode':mode,'unit':f['unit']})
sites=sorted(groups.values(),key=lambda f:(lc(f['where']),f['phase'],f['kind'],f['reason']))
for i,s in enumerate(sites,1):s['id']='F'+str(i)
(out/'rows.json').write_text(json.dumps({'records':records,'sites':sites},indent=2)+'\n')
for mode,data in records.items():print(mode,'units',len(data['record']['units']),'diagnostics',len(data['measurement']['diagnostics']),'findings',len(data['record']['findings']))
for s in sites:print(s['id'],s['where'].split('/debug.ts:')[-1],s['owner']['name'] if s['owner'] else '?',s['kind'],s['phase'],s['reason'])
