#!/usr/bin/env python3
"""Publish preparation, whole lane and each bounded shard piece for the new scope."""
import gzip
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import time
p=Path(__file__).resolve().parent;repo=p.parents[2];frozen=Path('/workspace/bind-owned-oracle-repo');scratch=Path('/workspace/scratch');out=p/'owned/oracle';out.mkdir(parents=True,exist_ok=True)
def publish(message,*paths):subprocess.run(['python3',str(p/'publish_paths.py'),message,*[str(x.relative_to(repo)) for x in paths]],cwd=repo,check=True)
def wait_file(path):
    while not path.exists():time.sleep(5)
    return json.loads(path.read_text())
metrics=wait_file(scratch/'bind-owned-oracle-prepare/metrics.json');assert metrics['exit']==0,metrics
ready=list((scratch/'bind-owned-oracle-cache').glob('*/ready.json'));assert len(ready)==1
artifact=ready[0].parent
source=Path('/workspace/bind-owned-after/src');actual=artifact/'tree/src'
digest=lambda root:{str(f.relative_to(root)):hashlib.sha256(f.read_bytes()).hexdigest() for f in root.rglob('*') if f.is_file()}
assert digest(source)==digest(actual)
code={}
for name in ['adapt.cjs','engine.cjs','owners.json','inventory.json','ownership.json']:
    assert (p/name).read_bytes()==(frozen/'stage3/adapt/99-correctness-axolotl-x'/name).read_bytes();code[name]=hashlib.sha256((p/name).read_bytes()).hexdigest()
(out/'source-identity.json').write_text(json.dumps(dict(source_files=len(digest(source)),source_identical=True,adapter_sha256=code,frozen_commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=frozen,text=True).strip()),indent=2)+'\n')
shutil.copyfile(ready[0],out/'artifact-ready.json');shutil.copyfile(scratch/'bind-owned-oracle-prepare/metrics.json',out/'prepare-metrics.json')
for label in ['apply','install','build','harness']:
    with gzip.open(out/(label+'.log.gz'),'wb') as arc:arc.write((artifact/(label+'.log')).read_bytes())
publish('Ownership revision oracle preparation and source identity',out,p/'run_owned_oracle.py',p/'publish_paths.py',p/'archive_oracle_piece.py')
# Keep timing-sensitive tests separate from census work.
wait_file(scratch/'bind-owned-latent-after-5/metrics.json')
while True:
    try:
        rows=[json.loads(s) for s in (scratch/'bind-owned-census-after.jsonl').read_text().splitlines()]
        if rows and len(rows[-1].get('roots',[]))==79:break
    except (FileNotFoundError,json.JSONDecodeError):pass
    time.sleep(5)
whole=scratch/'bind-owned-lane-whole';piece=scratch/'bind-owned-lane-whole-piece'
if not (piece/'metrics.json').exists():
    subprocess.run(['python3',str(p/'measure_piece.py'),str(piece),'--','stage3/lane/run.sh',str(whole),'--artifact',str(artifact)],cwd=frozen,check=True)
assert json.loads((piece/'metrics.json').read_text())['exit']==0
for f in whole.rglob('*'):
    if not f.is_file() or 'adapted-tree' in f.relative_to(whole).parts:continue
    target=out/'whole'/f.relative_to(whole);target.parent.mkdir(parents=True,exist_ok=True)
    if f.suffix in ['.log','.jsonl']:
        with gzip.open(str(target)+'.gz','wb') as arc:arc.write(f.read_bytes())
    else:shutil.copyfile(f,target)
shutil.copyfile(piece/'metrics.json',out/'whole/piece-metrics.json')
assert json.loads((whole/'report.json').read_text())['status']=='pass'
publish('Ownership revision whole lane PASS with exact main counts',out/'whole')
results=scratch/'bind-owned-oracle-shards'
for index in range(10):
    destination=scratch/('bind-owned-oracle-shards-'+str(index))
    if not (destination/'metrics.json').exists():
        subprocess.run(['python3',str(p/'measure_piece.py'),str(destination),'--','python3',str(p/'run_oracle_piece.py'),str(frozen),str(artifact),str(results),str(index)],cwd=repo,check=True)
    assert json.loads((destination/'metrics.json').read_text())['exit']==0
    subprocess.run(['python3',str(p/'archive_oracle_piece.py'),str(results),str(index),'--owned'],cwd=repo,check=True)
    publish(f'Ownership revision oracle shard piece {index} passes',out/'pieces'/str(index),out/'shards')
subprocess.run(['python3',str(p/'check_oracle_union.py'),str(frozen),str(results),str(whole),str(out/'union.json')],cwd=repo,check=True)
publish('Verify all 398 shards equal the whole lane by identities and API bytes',out/'union.json')
(scratch/'bind-owned-oracle-done.json').write_text(json.dumps(dict(pass_=True))+'\n')
