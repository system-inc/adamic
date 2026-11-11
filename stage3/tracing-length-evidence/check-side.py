from pathlib import Path
from run import run,raw,root
import os,sys
side=sys.argv[1];tree=Path('/workspace/scratch/tracing-length-'+side)
assert (tree/'patch-set.md').is_file(),'apply must finish first'
roots=sorted((tree/'src/compiler').rglob('*.ts'));assert len(roots)==79
run(side+'-types',[raw/'adamic','types',*roots])
env=dict(os.environ,LATENT_FULL='1',LATENT_ONLY_FILE=str(tree/'src/compiler/tracing.ts'),LATENT_ASSERT_NO_OUTPUT='1')
run(side+'-census',[raw/'latent',tree/'src/compiler',raw/(side+'-census.jsonl')],env=env)
