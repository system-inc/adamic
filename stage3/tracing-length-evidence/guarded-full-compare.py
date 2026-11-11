from pathlib import Path
from run import raw,root,run
import json,hashlib
assert json.loads((raw/'guarded-full-apply.json').read_text())['exit']==0
b=Path('/workspace/scratch/tracing-length-after');a=Path('/workspace/scratch/tracing-length-guarded-full')
files=sorted(str(p.relative_to(b)) for p in (b/'src').rglob('*') if p.is_file())
assert files==sorted(str(p.relative_to(a)) for p in (a/'src').rglob('*') if p.is_file());assert all((b/n).read_bytes()==(a/n).read_bytes() for n in files)
for n in ['typescript.d.ts','typescript.internal.d.ts']:
 assert (b/'built/local'/n).read_bytes()==(Path('/workspace/scratch/tracing-length-before/built/local')/n).read_bytes()
checkpoint=json.loads((raw/'guarded-checkpoint.json').read_text())['commit']
for name in ['adapt.cjs','pins.json']:
 relative='stage3/adapt/99-correctness-tracing-length/'+name
 run('checkpoint-blob-'+name,['git','show',checkpoint+':'+relative])
 assert (raw/('checkpoint-blob-'+name+'.log')).read_bytes()==(root/relative).read_bytes()
(raw/'guarded-full-source-report.json').write_text(json.dumps({'fullApplyExit':0,'allSourceFilesIdenticalToBuiltAndCountedCandidate':len(files),'normalDeclarationBundlesByteIdentical':['typescript.d.ts','typescript.internal.d.ts'],'implementationAndPinsUnchangedSinceCheckpoint':checkpoint},indent=2)+'\n')
print('full apply source matches the built and counted guarded candidate exactly')
