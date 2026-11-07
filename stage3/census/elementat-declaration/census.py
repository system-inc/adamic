import os,json,time,sys
from pathlib import Path
sys.path.insert(0,'/workspace/scratch/elementat-declaration')
from run import run
s=Path('/workspace/scratch/elementat-declaration')
while not (s/'candidate/src/compiler/core.ts').exists() or not any(json.loads(l)['name']=='build-core-latent' for l in (s/'commands.jsonl').read_text().splitlines()):time.sleep(1)
for variant in ['baseline','candidate']:
 tree=s/variant
 env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
 env.update(LATENT_FULL='1',LATENT_ONLY_FILE=str(tree/'src/compiler/core.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(s/(variant+'-times.jsonl')),GOMAXPROCS='2')
 row=run('core-census-'+variant,[str(s/'bin/core-latent'),str(tree/'src/compiler'),str(s/(variant+'-census.jsonl'))],env=env,limit=1400)
 assert row['exit']==0,row
