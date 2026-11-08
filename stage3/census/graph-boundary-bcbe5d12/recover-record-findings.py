"""Recover only missing per-unit finding payloads; never replace census counts."""
import importlib.util,json,os,sys
from pathlib import Path
sys.dont_write_bytecode=True
here=Path(__file__).resolve().parent;scratch=Path('/workspace/scratch/jawqwv7')
spec=importlib.util.spec_from_file_location('supervise',scratch/'supervise.py');m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
plan=json.loads((here/'record-recovery-plan.json').read_text())
for job in sorted(plan,key=lambda x:(not x['file'].endswith('/checker.ts'),not x['full_mode'],x['name'])):
 if len(sys.argv)>1 and job['name'] not in sys.argv[1:]:continue
 if (scratch/(job['name']+'-status.json')).exists():continue
 env={k:v for k,v in os.environ.items() if not k.startswith(('LATENT_','BOUNDARY_'))}
 env.update(LATENT_ONLY_FILE=job['file'],LATENT_ASSERT_NO_OUTPUT='1',GOMEMLIMIT='4GiB',RECORD_RECOVERY_UNITS=job['selection'],BOUNDARY_STREAM=str(scratch/(job['name']+'-findings.jsonl')),LATENT_TIMES=str(scratch/(job['name']+'-times.jsonl')))
 if job['full_mode']:env['LATENT_FULL']='1'
 (scratch/(job['name']+'-environment.json')).write_text(json.dumps({k:v for k,v in env.items() if k.startswith(('LATENT_','BOUNDARY_','RECORD_')) or k=='GOMEMLIMIT'},indent=2)+'\n')
 m.run(job['name'],[job.get('binary',str(scratch/'record-recovery-census')),'/workspace/graph-boundary-adapted/src/tsc/tsc.ts',str(scratch/(job['name']+'-latent.jsonl'))],env=env,watch=[env['BOUNDARY_STREAM']],limit=1490)
(scratch/'record-recovery-completion.json').write_text(json.dumps({'planned_jobs':len(plan),'finished_jobs':sum((scratch/(j['name']+'-status.json')).exists() for j in plan)},indent=2)+'\n')
print('recovery runner child_exit=0',flush=True)
