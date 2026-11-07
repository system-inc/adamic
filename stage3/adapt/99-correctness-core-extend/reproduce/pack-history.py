import tarfile,hashlib,shutil
from pathlib import Path
s=Path('/workspace/scratch/core-extend');old=s/'f027-evidence';raw=s/'history-evidence';out=Path('/workspace/adamic/stage3/adapt/99-correctness-core-extend')
raw.mkdir(exist_ok=True)
for p in old.iterdir():
 if p.is_file() and p.suffix in ['.py','.cjs','.json','.jsonl','.log','.txt','.md','.diff']:shutil.copyfile(p,raw/p.name)
for p in old.iterdir():
 if p.is_dir() and (p.name.startswith('piece-') or p.name.startswith('excluded-') or p.name in ['retained-artifacts','overlay']):shutil.copytree(p,raw/p.name,symlinks=True,ignore=shutil.ignore_patterns('adapted-tree','node_modules','.git'))
(raw/'BASE.txt').write_text('Historical frozen base f0271c6a7. Superseded by recut on 698965fe66aed945d18efa320a594a8a8b4df2f8. No historical count is included in current-main totals.\n')
(raw/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(raw))+'\n' for p in sorted(raw.rglob('*')) if p.is_file()))
a=out/'historical-f027.tar.gz'
with tarfile.open(a,'w:gz') as tar:tar.add(raw,arcname='historical-f027')
(out/'historical-f027.sha256').write_text(hashlib.sha256(a.read_bytes()).hexdigest()+'  '+a.name+'\n');print(a.stat().st_size,flush=True)
