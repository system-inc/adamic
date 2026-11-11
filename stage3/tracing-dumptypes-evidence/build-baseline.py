from pathlib import Path
import subprocess,sys
S=Path(__file__).resolve().parent;B=Path('/workspace/scratch/dp24r5m-base');R=Path('/workspace/adamic')
(B/'node_modules').symlink_to(R/'stage3/typescript/node_modules',target_is_directory=True)
p=subprocess.run(['python3',S/'run.py','baseline-build','npm','run','build'],cwd=B);print('baseline build wrapper exit:',p.returncode,flush=True);sys.exit(p.returncode)
