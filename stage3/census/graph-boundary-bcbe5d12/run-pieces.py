"""One entry-program pass or bounded full-mode file pieces, checkpointed each time."""
import importlib.util,json,os,subprocess,sys,time
from pathlib import Path
sys.dont_write_bytecode=True
here=Path(__file__).resolve().parent;root=here.parents[2];scratch=Path('/workspace/scratch/jawqwv7')
spec=importlib.util.spec_from_file_location('supervise',here/'supervise.py');m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
def run(name,cmd,**kw):return m.run(name,cmd,cwd=str(root),**kw)
def checkpoint(label):
 for name,cmd in [('snapshot',[sys.executable,str(here/'snapshot.py')]),('report',[sys.executable,str(here/'report.py')]),('snapshot-final',[sys.executable,str(here/'snapshot.py')])]:
  r=run(label+'-'+name,cmd)
  if r['exit_code']!=0:raise RuntimeError(r)
 rel=str(here.relative_to(root))
 for name,cmd in [('add',['git','add','--',rel]),('commit',['git','commit','-m','census checkpoint: '+label]),('push',['git','push','origin','HEAD:refs/heads/typescript/graph-boundary-census-checkpoint'])]:
  r=run(label+'-'+name,cmd)
  if r['exit_code']!=0:raise RuntimeError(r)
def census(name,full=False,files=None,limit=1790):
 env=os.environ.copy();env['BOUNDARY_STREAM']=str(scratch/(name+'-boundary.jsonl'));env['LATENT_TIMES']=str(scratch/(name+'-times.jsonl'));env['LATENT_ASSERT_NO_OUTPUT']='1'
 for key in ['LATENT_FULL','LATENT_SPECULATIVE','LATENT_ONLY_FILE','LATENT_PIECE_FILES','LATENT_PIECE_INDEX']:env.pop(key,None)
 if full:env['LATENT_FULL']='1';env['LATENT_ONLY_FILE']='\n'.join(files);env['GOMEMLIMIT']='10GiB'
 (scratch/(name+'-environment.json')).write_text(json.dumps({k:v for k,v in env.items() if k.startswith('LATENT_') or k in ('GOMEMLIMIT','BOUNDARY_STREAM')},indent=2)+'\n')
 cmd=[str(scratch/'census'),'/workspace/graph-boundary-adapted/src/tsc/tsc.ts',str(scratch/(name+'-latent.jsonl'))]
 return run(name,cmd,env=env,watch=[env['BOUNDARY_STREAM'],env['LATENT_TIMES']],limit=limit)
mode=sys.argv[1]
if mode=='apply':
 checkpoint('apply')
elif mode=='first':
 census('first');checkpoint('first')
elif mode=='full':
 header=json.loads((scratch/'first-latent.jsonl').read_text().splitlines()[0]);files=header['files']
 groups=[];current=[];size=0
 for file in files:
  n=Path(file).stat().st_size
  if n>600000:groups.append([file]);continue
  if size+n>600000 and current:groups.append(current);current=[];size=0
  current.append(file);size+=n
 if current:groups.append(current)
 # Start with the small accepted helper to validate full mode, then checker.
 smoke=next(f for f in files if f.endswith('/compiler/hostErrors.ts'))
 groups=[[f for f in g if f!=smoke] for g in groups]
 groups=[g for g in groups if g]
 groups.sort(key=lambda g:sum(Path(f).stat().st_size for f in g))
 checker_group=next(i for i,g in enumerate(groups) if any(f.endswith('/compiler/checker.ts') for f in g))
 checker=groups.pop(checker_group)
 groups=[[smoke],checker,*groups]
 assert set(files)=={f for g in groups for f in g}
 (scratch/'pieces.json').write_text(json.dumps(groups,indent=2)+'\n')
 deadline=time.monotonic()+60*60
 for i,group in enumerate(groups):
  left=deadline-time.monotonic()
  if left<60:
   (scratch/'pieces-stopped.json').write_text(json.dumps(dict(reason='60-minute full-piece budget exhausted',unstarted=groups[i:]),indent=2)+'\n');break
  name=f'full-{i:02d}'
  census(name,True,group,min(1790,left));checkpoint(name)
else:raise ValueError(mode)
