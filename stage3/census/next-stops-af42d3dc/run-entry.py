import os,sys
from pathlib import Path
from supervise import run,scratch
repo=Path('/workspace/adamic');tree=Path('/workspace/after-wave-adapted');work=scratch/'artifacts/tsc';work.mkdir(parents=True,exist_ok=True)
for name,binary,filename in [('entry-census','census','census.jsonl'),('entry-latent','entry-census','latent.jsonl')]:
 env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')};env.update(LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(scratch/'entry-unit-times.jsonl'))
 result=run(name,[str(scratch/binary),str(tree/'src/tsc/tsc.ts'),str(work/filename)],env=env,watch=[work/filename,scratch/'entry-unit-times.jsonl'],limit=1800)
 if result['exit_code']:sys.exit(1)
