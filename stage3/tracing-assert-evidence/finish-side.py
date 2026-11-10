from pathlib import Path
import sys,time,subprocess,os
from run import run,raw,root
side=sys.argv[1];tree=Path('/workspace/scratch/baobab-'+side)
needle='python3 /workspace/adamic/stage3/apply.py '+str(tree)
while needle in subprocess.check_output(['ps','-eo','args='],text=True).splitlines():time.sleep(1)
assert (tree/'patch-set.md').is_file(),'apply did not complete source adaptations'
roots=sorted((tree/'src/compiler').rglob('*.ts'));assert len(roots)==79
run(side+'-types',[raw/'adamic','types',*roots])
env=dict(os.environ,BAOBAB_CENSUS_ROOT=str(tree/'src/compiler/core.ts'))
run(side+'-root-census',[raw/'root-census',tree/'src/compiler',raw/(side+'-root-census.jsonl')],env=env)
modules=tree/'node_modules'
if not modules.exists():modules.symlink_to(root/'stage3/typescript/node_modules',target_is_directory=True)
run(side+'-build',['npm','run','build'],cwd=tree)
