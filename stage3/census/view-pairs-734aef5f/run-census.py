"""Bounded entry-first and per-file full coverage; publish every finished piece."""
import concurrent.futures,gzip,importlib.util,json,os,subprocess,sys,time
from pathlib import Path
sys.dont_write_bytecode=True
here=Path(__file__).resolve().parent;root=here.parents[2];scratch=Path('/workspace/scratch/vqzty8t');source=Path('/workspace/view-pairs-adapted')
spec=importlib.util.spec_from_file_location('supervise',here/'supervise.py');m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
def run(name,cmd,**kw):return m.run(name,cmd,cwd=str(root),**kw)
def checkpoint(name):
 for step,cmd in [('report',[sys.executable,str(here/'report.py')]),('add',['git','add','--',str(here.relative_to(root))]),('commit',['git','commit','-m','view-pairs census: '+name]),('push',['git','push','origin','HEAD:refs/heads/typescript/view-pairs-census'])]:
  r=run(name+'-'+step,cmd)
  if r['exit_code']:raise RuntimeError(r)
def measure(job):
 name=job['name'];env={k:v for k,v in os.environ.items() if not k.startswith(('LATENT_','BOUNDARY_','ADAMIC_VIEW_PAIRS_','RECORD_RECOVERY_'))}
 env.update(LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(scratch/(name+'-times.jsonl')),BOUNDARY_STREAM=str(scratch/(name+'-boundary.jsonl')),ADAMIC_VIEW_PAIRS_LOG=str(scratch/(name+'-pairs.jsonl')),ADAMIC_VIEW_PAIRS_UNITS=str(scratch/(name+'-units.jsonl')),GOMEMLIMIT='6GiB' if not job.get('parallel') else '4GiB')
 if job['mode']=='full':env.update(LATENT_FULL='1',LATENT_ONLY_FILE=job['file'])
 (scratch/(name+'-environment.json')).write_text(json.dumps({k:v for k,v in env.items() if k.startswith(('LATENT_','BOUNDARY_','ADAMIC_VIEW_PAIRS_')) or k=='GOMEMLIMIT'},indent=2)+'\n')
 return run(name,[str(scratch/'census'),str(source/'src/tsc/tsc.ts'),str(scratch/(name+'-latent.jsonl'))],env=env,watch=[env['ADAMIC_VIEW_PAIRS_LOG'],env['ADAMIC_VIEW_PAIRS_UNITS'],env['BOUNDARY_STREAM']],limit=1490)
# Preparation is an already-running supervised child, not another apply.
while not (scratch/'apply-status.json').exists():time.sleep(1)
assert json.loads((scratch/'apply-status.json').read_text())['exit_code']==0
assert json.loads((scratch/'build-status.json').read_text())['exit_code']==0
first={'name':'first','mode':'first-error'}
if not (scratch/'first-status.json').exists():
 checkpoint('setup');measure(first);checkpoint('first')
header=json.loads((scratch/'first-latent.jsonl').open().readline());files=header['files'];assert len(files)==81
compiler=source/'src/compiler'
priority=[str(compiler/n) for n in ['program.ts','builder.ts','builderPublic.ts','builderState.ts','builderStatePublic.ts','resolutionCache.ts','executeCommandLine.ts','sys.ts']]
ordered=priority+sorted([f for f in files if f not in priority],key=lambda f:(-Path(f).stat().st_size,f))
jobs=[dict(name=f'full-{i:03d}',file=f,mode='full',parallel=i>=len(priority) and Path(f).stat().st_size<=300000) for i,f in enumerate(ordered)]
plan={'source_base':'734aef5f3','probe':'86157bea8','entry':str(source/'src/tsc/tsc.ts'),'limit_seconds':1490,'jobs':[first,*jobs],'policy':'Entry first error, then one full-mode child per resolved file; priority and large files alone, at most two small files concurrently. Push every finished piece.'}
(scratch/'plan.json').write_text(json.dumps(plan,indent=2)+'\n')
def alone(j):
 if (scratch/(j['name']+'-status.json')).exists():return
 measure(j);checkpoint(j['name']+' '+Path(j['file']).name)
pending=[]
for job in jobs:
 if (scratch/(job['name']+'-status.json')).exists():continue
 if not job['parallel']:
  for old in pending:alone(old)
  pending=[];alone(job)
 else:pending.append(job)
 if len(pending)==2:
  with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
   futures={pool.submit(measure,j):j for j in pending}
   for future in concurrent.futures.as_completed(futures):
    future.result();j=futures[future];checkpoint(j['name']+' '+Path(j['file']).name)
  pending=[]
for job in pending:alone(job)
(scratch/'complete.json').write_text(json.dumps({'entry_pass_finished':True,'full_files_attempted':81})+'\n')
checkpoint('final');print('runner child_exit=0',flush=True)
