from pathlib import Path
from run import run,raw,root
import json,shutil,os
run('prepare-count',['python3',raw/'prepare-count.py'])
base=raw/'before-counted';(base/'built/local').mkdir(parents=True,exist_ok=True)
for p in Path('/workspace/scratch/tracing-length-before/built/local').glob('*.js'):shutil.copy2(p,base/'built/local'/p.name)
run('before-count-instrument',['node',raw/'instrument.cjs',base])
for side,tree in [('before',base),('after',raw/'counted-product/tree')]:
 directory=raw/(side+'-control-counts');directory.mkdir()
 env=dict(os.environ,NODE_OPTIONS='--require='+str(raw/'preload.cjs'),TRACING_LENGTH_COUNTER_DIR=str(directory))
 run(side+'-counter-control',['node',raw/'witness.cjs',tree,tree],env=env)
run('whole-lane',['go','test','-overlay',raw/'lane-overlay.json','./stage3/lane','-run','^TestShard_','-count=1','-parallel=2','-json','-timeout=90m'])
run('lane-reconciliation',['python3',raw/'lane-summary.py'])
