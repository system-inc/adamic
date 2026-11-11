from pathlib import Path
from run import run,raw,root
import time,json,os
while not (raw/'lane-summary.json').is_file():
 for n in ['whole-lane','lane-reconciliation']:
  p=raw/(n+'.json')
  if p.exists():assert json.loads(p.read_text())['exit']==0,n+' failed'
 time.sleep(1)
while not (raw/'guarded-full-apply.json').is_file():time.sleep(1)
assert json.loads((raw/'guarded-full-apply.json').read_text())['exit']==0
run('guarded-full-source-comparison',['python3',raw/'guarded-full-compare.py'])
run('final-docs',['python3',raw/'final-content.py'])
while not (raw/'frozen-final-apply.json').is_file():time.sleep(1)
assert json.loads((raw/'frozen-final-apply.json').read_text())['exit']==0
import hashlib
for name,digest in json.loads((raw/'frozen-adapter-inputs.json').read_text()).items():assert hashlib.sha256((root/'stage3/adapt/99-correctness-tracing-length'/name).read_bytes()).hexdigest()==digest,name
a=Path('/workspace/scratch/tracing-length-after');f=Path('/workspace/scratch/tracing-length-final-frozen')
files=sorted(str(p.relative_to(a)) for p in (a/'src').rglob('*') if p.is_file())
assert files==sorted(str(p.relative_to(f)) for p in (f/'src').rglob('*') if p.is_file());assert all((a/n).read_bytes()==(f/n).read_bytes() for n in files)
run('final-status',['git','status','--porcelain','--untracked-files=all','--ignore-submodules'])
for line in (raw/'final-status.log').read_text().splitlines():
 name=line[3:];assert name=='stage3/upstream/LEDGER.md' or name.startswith('stage3/adapt/99-correctness-tracing-length/'),name
run('final-add',['git','add','stage3/adapt/99-correctness-tracing-length','stage3/upstream/LEDGER.md'])
run('final-diff-check',['git','diff','--cached','--check'])
run('final-staged-files',['git','diff','--cached','--name-only'])
names=(raw/'final-staged-files.log').read_text().splitlines();assert len(names)==4,names
assert all(n=='stage3/upstream/LEDGER.md' or n.startswith('stage3/adapt/99-correctness-tracing-length/') for n in names)
(raw/'final-validation-report.json').write_text(json.dumps({'finalApplyExit':0,'sourceFilesIdenticalToBuiltAndCountedCandidate':len(files),'stagedFiles':names,'diffCheckExit':0,'readyForSingleFeatureCommit':True},indent=2)+'\n')
print('final adaptation applies cleanly, matches the validated runtime source, and is staged for review',flush=True)
