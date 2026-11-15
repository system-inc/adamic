from pathlib import Path
import subprocess,os,time,json
r=Path('/workspace/scratch/write-file-construct-run');s=Path('/workspace/scratch/write-file-slot-aprime/runtime-keys.cjs').read_text().replace('/workspace/scratch/write-file-slot-aprime',str(r)).replace("r+'/baseline-lane/adapted-tree", "r+'/control-adapted").replace("r+'/candidate/", "r+'/lane/adapted-tree/");(r/'runtime-keys.cjs').write_text(s)
while not (r/'lane/execution.json').exists() and not (r/'lane.exit').exists():time.sleep(5)
assert json.loads((r/'lane/execution.json').read_text())['apply_exit']==0
with (r/'runtime-keys.log').open('w') as log:p=subprocess.run(['node',str(r/'runtime-keys.cjs')],stdout=log,stderr=subprocess.STDOUT)
(r/'runtime-keys.exit').write_text(str(p.returncode)+'\n');assert p.returncode==0
