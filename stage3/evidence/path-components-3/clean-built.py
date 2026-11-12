import shutil,sys
from pathlib import Path
tree=Path(sys.argv[1]);assert str(tree) in ['/workspace/scratch/path-components-upstream','/workspace/scratch/path-components-after']
shutil.rmtree(tree/'built');print('removed generated built outputs only; source diagnostics preserved')
