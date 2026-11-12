import json,os,sys
from pathlib import Path
out=Path(__file__).parent;label=sys.argv[1]
tree=Path('/workspace/scratch/path-components-upstream' if label=='before' else '/workspace/scratch/path-components-after')
env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
env.update(LATENT_FULL='1',LATENT_ONLY_FILE=str(tree/'src/compiler/path.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(out/('unit-times-'+label+'.jsonl')))
(out/('census-env-'+label+'.json')).write_text(json.dumps({k:v for k,v in env.items() if k.startswith('LATENT_')},indent=2)+'\n')
os.execve(str(out/'latent'),[str(out/'latent'),str(tree/'src/compiler'),str(out/('census-'+label+'.jsonl'))],env)
