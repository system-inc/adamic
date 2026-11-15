from pathlib import Path
import subprocess,os
r=Path('/workspace/scratch/write-file-slot-run')
for label in ['a','b']:
 with (r/('production-'+label+'.log')).open('w') as log:
  subprocess.run([str(r/'production-census'),str(r/label/'src/tsc/tsc.ts'),str(r/('production-'+label+'.jsonl'))],env=dict(os.environ,CENSUS_MODULE_FORMAT='CommonJSBundle'),stdout=log,stderr=subprocess.STDOUT,check=True)
