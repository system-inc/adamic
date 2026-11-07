import hashlib,json,tarfile,shutil
from pathlib import Path
s=Path('/workspace/scratch/core-extend');out=Path('/workspace/adamic/stage3/adapt/99-correctness-core-extend');raw=s/'proof-evidence'
if raw.exists():shutil.rmtree(raw)
raw.mkdir()
exclude={'count-summary.json','count-pieces.log','count-pause.json','restart-count.log','count-reconciled.json'}
for p in sorted(s.iterdir()):
 if p.is_file() and p.suffix in ['.json','.jsonl','.log','.txt','.cjs','.py','.diff','.md'] and p.name not in exclude:
  shutil.copyfile(p,raw/p.name)
for name in ['overlay','whole-control','excluded-view-setup-attempt','excluded-prototype-prefix']:
 if (s/name).exists():shutil.copytree(s/name,raw/name,symlinks=True,ignore=shutil.ignore_patterns('adapted-tree','node_modules','.git'))
for side in ['before','after']:
 shutil.copyfile(s/side/'patch-set.md',raw/('patch-set-'+side+'.md'))
(raw/'javascript.diff').write_text('');(raw/'public-declarations.diff').write_text('')
(raw/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(raw))+'\n' for p in sorted(raw.rglob('*')) if p.is_file()))
archive=out/'core-proof.tar.gz'
with tarfile.open(archive,'w:gz') as tar:tar.add(raw,arcname='proof')
(out/'core-proof.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
for p in [*s.glob('*.py'),*s.glob('*.cjs')]:shutil.copyfile(p,out/'reproduce'/p.name)
print('proof archive',archive.stat().st_size,flush=True)
