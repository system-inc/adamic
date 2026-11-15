from pathlib import Path
import os,subprocess,json
r=Path('/workspace/scratch/write-file-slot-run')
for label in ['main','a','b','hidden']:
 env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_ONLY_FILE=str(r/label/'src/compiler/sys.ts'),TMPDIR=str(r/'tmp'),GOTMPDIR=str(r/'tmp'),GOCACHE=str(r/'gocache'))
 with (r/(label+'-census.log')).open('w') as log:
  result=subprocess.run([str(r/'census'),str(r/label/'src/compiler'),str(r/(label+'.jsonl'))],env=env,stdout=log,stderr=subprocess.STDOUT)
 (r/(label+'-census-status.json')).write_text(json.dumps({'exit':result.returncode}))
 if result.returncode: raise SystemExit(result.returncode)
 print(label+' census complete',flush=True)
