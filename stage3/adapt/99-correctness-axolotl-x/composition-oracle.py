#!/usr/bin/env python3
"""Run the unchanged whole lane and every standard shard for two fixed compositions."""
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import time
p=Path(__file__).resolve().parent;repo=p.parents[2];scratch=Path('/workspace/scratch/bind2');out=p/'composition/oracle';out.mkdir(parents=True,exist_ok=True)
def wait(path):
 while not path.exists():time.sleep(5)
 return json.loads(path.read_text())
def run(command,cwd):subprocess.run(command,cwd=cwd,check=True)
def measure(label,command,cwd):
 destination=scratch/label
 if not (destination/'metrics.json').exists():run(['python3',str(p/'measure_piece.py'),str(destination),'--',*map(str,command)],cwd)
 assert wait(destination/'metrics.json')['exit']==0
 return destination
def pack(raw,target):
 run(['python3',str(p/'composition-pack.py'),str(raw),str(target)],repo)
 return [target,target.with_suffix('').with_suffix('.sha256'),Path(str(target)+'.sha256')]
def publish(message,*paths):run(['python3',str(p/'publish_paths.py'),message,*[str(x.relative_to(repo)) for x in paths]],repo)
def digest(root):return {str(f.relative_to(root)):hashlib.sha256(f.read_bytes()).hexdigest() for f in root.rglob('*') if f.is_file()}
prepared={}
# Preparation can run while non-oracle checks finish; timing-sensitive tests wait below.
for label in ['x','32x']:
 assert wait(p/'composition/matrix'/(label+'.json'))['status']=='pass'
 frozen=Path('/workspace/bind2-'+label+'-repo');cache=scratch/('oracle-cache-'+label)
 measured=measure('oracle-prepare-'+label,['python3','stage3/lane/prepare_shards.py',cache],frozen)
 artifacts=list(cache.glob('*/ready.json'));assert len(artifacts)==1;artifact=artifacts[0].parent;prepared[label]=artifact
 assert digest(artifact/'tree/src')==digest(Path('/workspace/bind2-'+label+'-tree/src'))
 for name in ['adapt.cjs','engine.cjs','inventory.json','owners.json','ownership.json']:assert (p/name).read_bytes()==(frozen/'stage3/adapt/99-correctness-axolotl-x'/name).read_bytes()
 raw=scratch/('oracle-preparation-receipt-'+label);raw.mkdir(exist_ok=True)
 for name in ['ready.json','shards-ready.json','apply.log','install.log','build.log','harness.log']:shutil.copyfile(artifact/name,raw/name)
 shutil.copyfile(measured/'metrics.json',raw/'metrics.json')
 result=dict(status='pass',source_files_identical=len(digest(artifact/'tree/src')),frozen_commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=frozen,text=True).strip(),artifact=artifact.name,metrics=wait(measured/'metrics.json'))
 target=out/label;target.mkdir(exist_ok=True);summary=target/'preparation.json';summary.write_text(json.dumps(result,indent=2)+'\n')
 publish('99x '+label+' oracle preparation and source identity pass',summary,*pack(raw,target/'preparation.tar.gz'))
# All full applies, compiler builds and latent measurements finish before timeout-sensitive tests.
assert wait(p/'composition/matrix/both.json')['status']=='pass'
assert wait(p/'composition/latent/summary.json')['status']=='pass'
for label in ['x','32x']:
 frozen=Path('/workspace/bind2-'+label+'-repo');artifact=prepared[label];target=out/label
 whole=scratch/('oracle-whole-'+label);measured=measure('oracle-whole-piece-'+label,['stage3/lane/run.sh',whole,'--artifact',artifact],frozen)
 verdict=wait(whole/'report.json');assert verdict['status']=='pass'
 raw=scratch/('oracle-whole-receipt-'+label);raw.mkdir(exist_ok=True)
 for f in whole.rglob('*'):
  if not f.is_file() or 'adapted-tree' in f.relative_to(whole).parts:continue
  dest=raw/f.relative_to(whole);dest.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(f,dest)
 shutil.copyfile(measured/'metrics.json',raw/'piece-metrics.json')
 summary=target/'whole.json';summary.write_text(json.dumps(dict(verdict=verdict,metrics=wait(measured/'metrics.json')),indent=2)+'\n')
 publish('99x '+label+' whole lane PASS',summary,*pack(raw,target/'whole.tar.gz'))
 results=scratch/('oracle-shards-'+label)
 for index in range(10):
  measured=measure('oracle-shards-'+label+'-'+str(index),['python3',p/'run_oracle_piece.py',frozen,artifact,results,str(index)],repo)
  piece=wait(results/('piece-'+str(index))/'results.json');assert not piece['failed']
  raw=scratch/('oracle-piece-receipt-'+label+'-'+str(index));raw.mkdir(exist_ok=True)
  shutil.copyfile(measured/'metrics.json',raw/'metrics.json');shutil.copyfile(measured/'process.log',raw/'process.log');shutil.copyfile(results/('piece-'+str(index))/'results.json',raw/'results.json')
  for row in piece['units']:
   assert row['exit']==0 and row['report']['status']=='pass'
   for f in (results/'units'/row['name']).rglob('*'):
    if not f.is_file() or 'adapted-tree' in f.relative_to(results/'units'/row['name']).parts:continue
    dest=raw/'shards'/row['name']/f.relative_to(results/'units'/row['name']);dest.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(f,dest)
  summary=target/('piece-'+str(index)+'.json');summary.write_text(json.dumps(dict(status='pass',units=len(piece['units']),metrics=wait(measured/'metrics.json')),indent=2)+'\n')
  publish('99x '+label+' oracle shard piece '+str(index)+' passes',summary,*pack(raw,target/('piece-'+str(index)+'.tar.gz')))
 union=target/'union.json';run(['python3',str(p/'check_oracle_union.py'),str(frozen),str(results),str(whole),str(union)],repo)
 publish('99x '+label+' all 398 shards exactly match the whole lane',union)
(scratch/'oracle-done.json').write_text(json.dumps(dict(status='pass',variants=['x','32x']))+'\n')
