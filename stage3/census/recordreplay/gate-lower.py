from pathlib import Path
import time,subprocess as s,os
D=Path('/workspace/scratch/recordreplay')
while 'gate stock and declaration build complete' not in (D/'gate.log').read_text():time.sleep(2)
for label in ['control','candidate']:
 tree=D/label
 s.run(['python3',str(D/'run.py'),label+'-types',str(D/'adamic'),'types',str(tree/'src/compiler/sys.ts')],cwd=D/'head')
 env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_ONLY_FILE=str(tree/'src/compiler/sys.ts'))
 s.run(['python3',str(D/'run.py'),label+'-sys-full',str(D/'latent-head'),str(tree/'src/compiler/sys.ts'),str(D/(label+'-sys-full.jsonl'))],cwd=D/'head',env=env)
print('gate lower observations complete',flush=True)
