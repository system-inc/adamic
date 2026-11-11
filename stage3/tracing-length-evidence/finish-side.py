from pathlib import Path
from run import run,raw,root
import sys,time
side=sys.argv[1];tree=Path('/workspace/scratch/tracing-length-'+side)
while not (raw/(side+'-apply.json')).is_file():time.sleep(1)
import json
assert json.loads((raw/(side+'-apply.json')).read_text())['exit']==0,'apply must exit successfully first'
run(side+'-build-wrapper',['python3',raw/'build-side.py',side])
run(side+'-checks-wrapper',['python3',raw/'check-side.py',side])
if side=='after':run('mutants-wrapper',['python3',raw/'mutants.py'])
