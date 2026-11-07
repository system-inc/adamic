import json,os,sys
from pathlib import Path
sys.path.insert(0,'/workspace/adamic/stage3/census/next-stops-af42d3dc')
from supervise import run,scratch
repo=Path('/workspace/adamic');out=Path(__file__).resolve().parent;tree=Path('/workspace/after-wave-adapted')
env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
env.update(LATENT_FULL='1',LATENT_ONLY_FILE=str(tree/'src/compiler/core.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(scratch/'unit-times.jsonl'))
result=run('core-census',[str(scratch/'latent'),str(tree/'src/compiler'),str(scratch/'raw.jsonl')],env=env,watch=[scratch/'raw.jsonl',scratch/'unit-times.jsonl'],limit=3600)
(out/'census-result.json').write_text(json.dumps(result,indent=2)+'\n')
sys.exit(0 if result['exit_code']==0 else 1)
