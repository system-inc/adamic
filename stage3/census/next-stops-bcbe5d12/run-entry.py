import json,os,sys,time
from pathlib import Path
from supervise import run,scratch
out=Path(__file__).resolve().parent;tree=Path('/workspace/bcbe5d12-adapted');work=scratch/'artifacts/tsc';work.mkdir(parents=True,exist_ok=True)
for status in ['entry-build','apply-checkpoint']:
 while not (scratch/(status+'-status.json')).exists():time.sleep(1)
 assert json.loads((scratch/(status+'-status.json')).read_text())['exit_code']==0
for name,binary,filename in [('entry-census','census','census.jsonl'),('entry-latent','entry-census','latent.jsonl')]:
 env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')};env.update(LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(scratch/'entry-unit-times.jsonl'))
 result=run(name,[str(scratch/binary),str(tree/'src/tsc/tsc.ts'),str(work/filename)],env=env,watch=[work/filename,scratch/'entry-unit-times.jsonl'],limit=1800)
 if result['exit_code']:sys.exit(result['exit_code'] if result['exit_code']>0 else 128-result['exit_code'])
