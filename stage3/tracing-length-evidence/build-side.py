from pathlib import Path
from run import run,raw,root
import sys,os
side=sys.argv[1];tree=Path('/workspace/scratch/tracing-length-'+side)
assert (tree/'patch-set.md').is_file(),'apply must finish first'
link=tree/'node_modules'
if not link.exists():link.symlink_to(root/'stage3/typescript/node_modules',target_is_directory=True)
run(side+'-build',['npm','run','build'],cwd=tree)
