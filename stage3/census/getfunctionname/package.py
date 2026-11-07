"""Retain raw evidence after the accepted runtime inventory reconciles."""
from pathlib import Path
import json,hashlib,tarfile,shutil,subprocess
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/getfunctionname');raw=s/'raw';out=repo/'stage3/census/getfunctionname';e=s/'evidence'
load=lambda p:json.loads(p.read_text())
assert load(raw/'runtime-reconciled.json')['complete']
reuse=load(raw/'reuse-validation.json');manifest=reuse.pop('manifest');reuse['full_validation_sha256']=hashlib.sha256((raw/'reuse-validation.json').read_bytes()).hexdigest();reuse['input_manifest_sha256']=hashlib.sha256(json.dumps(manifest,sort_keys=True,separators=(',',':')).encode()).hexdigest();reuse['source_inputs']={k:v for k,v in manifest.items() if k.startswith('src/')};reuse['prior_archive_sha256']=hashlib.sha256((raw/'reused-javascript.tar.gz').read_bytes()).hexdigest();reuse['prior_archive_git_blob']=subprocess.check_output(['git','rev-parse','385ccad0bf6cfd800443566d947cc638a7e776aa:stage3/adapt/99-correctness-core-debug-messages/after-javascript.tar.gz'],cwd=repo,text=True).strip();(out/'reuse-validation-summary.json').write_text(json.dumps(reuse,indent=2)+'\n')
for name in ['references.json','declarations.json','pins.json','runtime-reconciled.json','coverage-control.json','positive-controls.json','instrumentation.json']:shutil.copy2(raw/name,out/name)
shutil.copy2(e/'runtime-counts.json',out/'runtime-counts.json')
for p in sorted(e.glob('lane-*.tar.gz')):shutil.copy2(p,out/p.name)
# Full reuse input map retained compressed, but the existing compiled archive and
# the generated control-only library are linked by hash instead of duplicated.
excluded={'reused-javascript.tar.gz','control-library.cjs'}
with tarfile.open(out/'raw-measurement.tar.gz','w:gz') as tf:
 for p in sorted(raw.rglob('*')):
  if p.is_file() and p.name not in excluded:tf.add(p,arcname='raw/'+str(p.relative_to(raw)))
 for p in sorted((s/'road-b').glob('*')):
  if p.is_file():tf.add(p,arcname='road-b/'+p.name)
with tarfile.open(out/'discarded-initial-run.tar.gz','w:gz') as tf:
 for p in sorted((s/'initial-evidence').glob('*')):
  if p.is_file():tf.add(p,arcname=p.name)
archives=[{'file':p.name,'bytes':p.stat().st_size,'sha256':hashlib.sha256(p.read_bytes()).hexdigest()} for p in sorted(out.glob('*.tar.gz'))];assert all(a['bytes']<20_000_000 for a in archives);(out/'ARCHIVES.json').write_text(json.dumps(archives,indent=2)+'\n')
print(json.dumps({'archives':len(archives),'archiveBytes':sum(a['bytes'] for a in archives),'largest':max(a['bytes'] for a in archives)}))
