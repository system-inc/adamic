from pathlib import Path
import time,json,os
from run import run,raw,root
while not (raw/'whole-lane-exit.json').exists():time.sleep(1)
assert json.loads((raw/'whole-lane-exit.json').read_text())['exit']==0,'whole lane failed; finalization stopped'
run('lane-reconciliation',['python3','-B',raw/'lane-summary.py'])
run('final-content',['python3','-B',raw/'finalize-content.py'])
run('final-stage',['git','add','stage3/adapt/99-correctness-tracing-assert','stage3/upstream/LEDGER.md'])
run('final-diff-check',['git','diff','--cached','--check'])
env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-6.0.3')
run('final-apply',['bash','stage3/apply.sh','/workspace/scratch/baobab-final'],env=env)
a=Path('/workspace/scratch/baobab-after');f=Path('/workspace/scratch/baobab-final')
files=sorted(str(p.relative_to(a)) for p in (a/'src').rglob('*') if p.is_file());assert all((a/name).read_bytes()==(f/name).read_bytes() for name in files)
(raw/'final-source-validation.json').write_text(json.dumps({'sourceFiles':len(files),'allBytesIdenticalToValidatedCandidate':True,'finalApplyExit':0},indent=2)+'\n')
print('final saved adaptation content applies cleanly and preserves the validated source',flush=True)
