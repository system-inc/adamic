#!/usr/bin/env python3
"""Pack the continuation's prechecks without adding raw replay or latent files."""
from pathlib import Path
import hashlib
import json
import shutil
import subprocess
import sys
import tarfile
repo,scratch=map(lambda s:Path(s).resolve(),sys.argv[1:])
unit=repo/'stage3/adapt/32-indexed-reads-program';out=unit/'proof/maybebind/continuation';out.mkdir(parents=True,exist_ok=True)
def digest(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def pack(root,name):
    files=sorted(p for p in root.rglob('*') if p.is_file())
    with tarfile.open(out/(name+'.tar.gz'),'w:gz') as archive:
        for p in files:archive.add(p,arcname=p.relative_to(root),recursive=False)
    (out/(name+'-files.sha256')).write_text(''.join(digest(p)+'  '+str(p.relative_to(root))+'\n' for p in files))
    (out/(name+'.sha256')).write_text(digest(out/(name+'.tar.gz'))+'  '+name+'.tar.gz\n')
replay=scratch/'32r-replay';status=json.loads((replay/'status.json').read_text());assert status['complete'] and status['differences']==[]
pack(replay,'replay')
summary={n:json.loads((replay/n).read_text()) for n in ['before.json','after.json','status.json','coverage.json']}
(out/'replay-records.json').write_text(json.dumps(summary,indent=2)+'\n');shutil.copyfile(replay/'records.diff',out/'replay.diff')
assert (out/'replay.diff').stat().st_size==0
pack(scratch/'32r-latent','latent')
shutil.copyfile(scratch/'32r-cw051-summary.json',out/'cw051.json')
logs=scratch/'32r-pack-logs';logs.mkdir(exist_ok=True)
for p in scratch.glob('32r-*'):
    if p.is_file() and p.suffix in ['.log','.exit','.json']:shutil.copyfile(p,logs/p.name)
for p in [scratch/'32r-latent-build/provenance.json',scratch/'32r-latent-build/selector.patch']:
    shutil.copyfile(p,logs/p.name)
pack(logs,'precheck-logs')
functional=['adapt.cjs','maybebind.cjs','maybebind-sites.json','public-host-guards.cjs','public-host-guards.json']
for name in functional:
    old=subprocess.check_output(['git','show','997331ef1:'+str((unit/name).relative_to(repo))],cwd=repo)
    assert old==(unit/name).read_bytes(),name
source=lambda root:{str(p.relative_to(root/'src')):digest(p) for p in (root/'src').rglob('*') if p.is_file()}
unchanged={}
for side,old in [('before',Path('/workspace/32-before-final')),('after',Path('/workspace/32-after-final'))]:
    unchanged[side]=source(scratch/('32r-'+side))==source(old)
    assert unchanged[side]
provenance=dict(main='12fec995dcadc7f1ae94e4ba63835884451378ae',fold='997331ef1ec86480f47f3710fd43e1150407dafc',adapter99x='5fa773b799e417c568ae62a4003a801333b87714',foldBaseRetainedInMain=True,ancestryCommand='git merge-base --is-ancestor 322f80565 12fec995',functionalInputs={n:digest(unit/n) for n in functional},functionalInputsUnchanged=True,completeCompilerSourceTreesIdenticalToPriorProof=unchanged,ordinaryCompilerSha256=digest(scratch/'32r-adamic'),types=json.loads((scratch/'32r-types.json').read_text()),fullApplyBeforeExit=int((scratch/'32r-apply-before.exit').read_text()),fullApplyAfterExit=int((scratch/'32r-apply-after.exit').read_text()))
assert provenance['fullApplyBeforeExit']==provenance['fullApplyAfterExit']==0
artifacts=list((scratch/'32r-oracle-cache').glob('*/ready.json'));assert len(artifacts)==1
artifact=artifacts[0].parent
assert source(artifact/'tree')==source(scratch/'32r-after')
provenance['oracle']=dict(artifact=str(artifact),ready=json.loads(artifacts[0].read_text()),allCompilerSourcesEqual=True,preparation=json.loads((scratch/'32r-oracle-prepare/metrics.json').read_text()))
assert provenance['oracle']['preparation']['exit']==0
(out/'provenance.json').write_text(json.dumps(provenance,indent=2)+'\n')
print(json.dumps(dict(status='pass',replayAdapters=status['adapters_recorded'],cw051='watch.ts:855:19',sourceTreesUnchanged=unchanged)))
