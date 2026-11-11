from pathlib import Path
import subprocess,sys,os
S=Path(__file__).resolve().parent;side=sys.argv[1];tree=Path('/workspace/scratch/dp24r5m-'+side)
env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(tree/'src/compiler/tracing.ts'),LATENT_TIMES=str(S/(side+'-census-times.jsonl')))
p=subprocess.run(['python3',S/'run.py',side+'-census-run',S/'census',tree/'src/compiler',S/(side+'-census.jsonl')],env=env);print('census wrapper exit:',p.returncode,flush=True);sys.exit(p.returncode)
