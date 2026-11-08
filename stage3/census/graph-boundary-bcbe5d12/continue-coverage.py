"""Bound each missing file separately; publish completed pieces in completion order."""
import concurrent.futures,importlib.util,json,os,sys,time
from pathlib import Path
sys.dont_write_bytecode=True
here=Path(__file__).resolve().parent;root=here.parents[2];scratch=Path('/workspace/scratch/jawqwv7')
spec=importlib.util.spec_from_file_location('supervise',here/'supervise.py');m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
compiler=Path('/workspace/graph-boundary-adapted/src/compiler')
missing=[f for g in json.loads((scratch/'pieces-stopped.json').read_text())['unstarted'] for f in g]
assert len(missing)==51 and len(set(missing))==51
priority=[str(compiler/f) for f in ['program.ts','builder.ts','builderPublic.ts','builderState.ts','builderStatePublic.ts','resolutionCache.ts','executeCommandLine.ts']]
files=priority+[f for f in missing if f not in priority]
jobs=[dict(name=f'full-20-{i:03d}',file=f,priority=i<len(priority)) for i,f in enumerate(files)]
plan=dict(parent='772a979ea56a4738b1255cffa4da15cbb303ca5f',requested_missing=missing,jobs=jobs,limit_seconds=1490,go_memory_limit='4GiB',policy='Priority files and files over 300000 bytes alone; remaining small files at most two concurrent. Commit and push each completed piece before proceeding.')
(scratch/'coverage-followup.json').write_text(json.dumps(plan,indent=2)+'\n')
def run(name,cmd,**kw):return m.run(name,cmd,cwd=str(root),**kw)
def measure(job):
 name=job['name'];env=os.environ.copy()
 for key in list(env):
  if key.startswith('LATENT_') or key.startswith('BOUNDARY_'):env.pop(key)
 env.update(LATENT_FULL='1',LATENT_ONLY_FILE=job['file'],LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(scratch/(name+'-times.jsonl')),BOUNDARY_STREAM=str(scratch/(name+'-boundary.jsonl')),GOMEMLIMIT='4GiB')
 (scratch/(name+'-environment.json')).write_text(json.dumps({k:v for k,v in env.items() if k.startswith(('LATENT_','BOUNDARY_')) or k=='GOMEMLIMIT'},indent=2)+'\n')
 return run(name,[str(scratch/'census'),'/workspace/graph-boundary-adapted/src/tsc/tsc.ts',str(scratch/(name+'-latent.jsonl'))],env=env,watch=[env['LATENT_TIMES'],env['BOUNDARY_STREAM']],limit=1490)
def checkpoint(job):
 name=job['name']
 for step,cmd in [('snapshot',[sys.executable,str(here/'snapshot.py')]),('report',[sys.executable,str(here/'report.py')]),('audit',[sys.executable,str(here/'audit.py')]),('snapshot-final',[sys.executable,str(here/'snapshot.py')]),('add',['git','add','--',str(here.relative_to(root))]),('commit',['git','commit','-m','census coverage: '+Path(job['file']).name]),('push',['git','push','origin','HEAD:refs/heads/typescript/graph-boundary-census'])]:
  result=run(name+'-'+step,cmd)
  if result['exit_code']!=0:raise RuntimeError(result)
def alone(job):measure(job);checkpoint(job)
for job in jobs[:len(priority)]:alone(job)
# Small-file concurrency is process orchestration only. Large files run alone,
# and every completed child is reaped before its own report/push is performed.
pending=[]
for job in jobs[len(priority):]:
 if Path(job['file']).stat().st_size>300000:
  for small in pending:alone(small)
  pending=[];alone(job)
 else:pending.append(job)
 if len(pending)==2:
  with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
   futures={pool.submit(measure,j):j for j in pending}
   for future in concurrent.futures.as_completed(futures):future.result();checkpoint(futures[future])
  pending=[]
for job in pending:alone(job)
(scratch/'coverage-followup-complete.json').write_text(json.dumps(dict(files=len(jobs),requested_missing=len(missing),completed=True))+'\n')
print('coverage runner child_exit=0',flush=True)
