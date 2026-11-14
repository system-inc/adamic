"""Retargeted next-stop recipe, one root child at a time, push each group."""
from pathlib import Path
import subprocess as s,os,json,gzip,shutil,time
D=Path('/workspace/scratch/recordreplay');R=D/'base';O=R/'stage3/census/recordreplay';S=O/'next-stops'
def checkpoint(message):
 s.run(['git','add',str(O.relative_to(R))],cwd=R,check=True)
 dirty=s.run(['git','diff','--cached','--quiet'],cwd=R)
 if dirty.returncode:
  s.run(['git','commit','-m',message],cwd=R,check=True)
  s.run(['git','push','origin','HEAD:refs/heads/typescript/recordreplay-census'],cwd=R,check=True)
def norm(v):
 if isinstance(v,str):return v.replace(str(D/'control')+'/', '').replace(str(D/'candidate2')+'/', '')
 if isinstance(v,list):return [norm(x) for x in v]
 if isinstance(v,dict):return {k:norm(x) for k,x in v.items() if k not in ['seconds','roots']}
 return v
def compare(pair,mode):
 root=O/'ordinary';before=root/(pair+'-main')/mode;after=root/(pair+'-head')/mode
 rows=[]
 for i in range(1,80):
  a=before/f'root-{i:02}.jsonl';b=after/f'root-{i:02}.jsonl'
  if not a.exists() or not b.exists():continue
  aa=json.loads(a.read_text().splitlines()[0]);bb=json.loads(b.read_text().splitlines()[0])
  rows.append({'root':aa['roots'][0].split('/src/compiler/')[-1],'before':norm(aa),'after':norm(bb),'unchanged':norm(aa)==norm(bb)})
 p=root/(pair+'-'+mode+'-paired.json');p.write_text(json.dumps({'rows':rows,'complete':len(rows)==79,'moved':[r for r in rows if not r['unchanged']]},indent=2)+'\n')
for pair,tree in [('A',D/'control'),('B',D/'candidate2')]:
 for side,binary in [('main',D/'census-base'),('head',D/'census-head')]:
  label=pair+'-'+side
  for mode in ['CommonJSBundle','EcmaScript']:
   dest=O/'ordinary'/label/mode;dest.mkdir(parents=True,exist_ok=True)
   env=dict(os.environ,CENSUS_MODULE_FORMAT=mode,CENSUS_REPO=str(R),CENSUS_TREE=str(tree),CENSUS_OUTPUT=str(dest),CENSUS_BINARY=str(binary),CENSUS_LABEL=label+'-'+mode,RUN_LIMIT='1700')
   for first in range(1,80,10):
    last=min(first+9,79)
    for i in range(first,last+1):
     target=dest/f'root-{i:02}.jsonl'
     if target.exists():continue
     print('START',label,mode,i,flush=True)
     s.run(['python3',str(S/'run-one.py'),str(i)],cwd=R,env=env,check=True)
     resource=D/(label+'-'+mode+f'-ordinary-root-{i:02}.json')
     shutil.copyfile(resource,dest/f'root-{i:02}-resource.json')
     log=resource.with_suffix('.log');(dest/f'root-{i:02}.log.gz').write_bytes(gzip.compress(log.read_bytes(),mtime=0))
     # The scheduler log is only the resource line; the child's actual log is retained above.
     (dest/f'root-{i:02}.log').unlink(missing_ok=True)
    compare(pair,mode)
    checkpoint(f'stage3: recordreplay {label} {mode} roots {first}-{last}')
    print('CHECKPOINT',label,mode,first,last,flush=True)
print('ordinary tables complete',flush=True)
