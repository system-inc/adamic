import hashlib,io,pathlib,sys,tarfile
root=pathlib.Path(sys.argv[1]);target=pathlib.Path(sys.argv[2]);snapshots={}
for p in root.iterdir():
 if p.is_file() and p.name!='latent':snapshots[p.name]=p.read_bytes()
for name in ['census-before','census-after','census-mutant','count','seed-count','mutant-seed-count','lane-count','physical-count','after-lane','replay','checks','overlay']:
 directory=root/name
 if not directory.exists():continue
 for p in sorted(directory.rglob('*')):
  relative=p.relative_to(root)
  if any(part in ['adapted-tree','tree','results','replay-before','replay-after','__pycache__','node_modules','.git'] for part in relative.parts):continue
  if p.is_file() and not p.is_symlink():
   assert p.suffix!='.go',p
   snapshots[str(relative)]=p.read_bytes()
manifest=''.join(hashlib.sha256(data).hexdigest()+'  '+name+'\n' for name,data in sorted(snapshots.items())).encode()
snapshots['SHA256SUMS']=manifest
with tarfile.open(target,'w:gz') as archive:
 for name,data in sorted(snapshots.items()):
  info=tarfile.TarInfo('queue-box/'+name);info.size=len(data);info.mode=0o644;archive.addfile(info,io.BytesIO(data))
target.with_suffix('').with_suffix('.sha256').write_text(hashlib.sha256(target.read_bytes()).hexdigest()+'  '+target.name+'\n')
print({'files':len(snapshots),'bytes':target.stat().st_size,'sha256':hashlib.sha256(target.read_bytes()).hexdigest()})
