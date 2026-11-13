"""Archive a completed lane's raw records without caches or transient source views."""
import hashlib,json,pathlib,sys,tarfile
source=pathlib.Path(sys.argv[1]);target=pathlib.Path(sys.argv[2])
with tarfile.open(target,'w:gz',compresslevel=6) as archive:
 for file in sorted(source.rglob('*')):
  relative=file.relative_to(source)
  if file.is_symlink() or not file.is_file() or any(p in {'caches','adapted-tree','__pycache__'} for p in relative.parts):continue
  name=str(relative)
  if name.endswith('.go'):name+='.txt'
  archive.add(file,arcname=name,recursive=False)
record=dict(file=target.name,bytes=target.stat().st_size,sha256=hashlib.sha256(target.read_bytes()).hexdigest())
target.with_suffix(target.suffix+'.sha256.json').write_text(json.dumps(record,indent=2)+'\n')
print(json.dumps(record))
