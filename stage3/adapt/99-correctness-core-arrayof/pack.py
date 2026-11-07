from pathlib import Path
import hashlib,io,json,sys,tarfile
s=Path('/workspace/scratch/arrayof');output=Path(sys.argv[1]);patterns=sys.argv[2:];files=set()
for pattern in patterns:
 for p in s.glob(pattern):
  candidates=p.rglob('*') if p.is_dir() else [p]
  for f in candidates:
   rel=f.relative_to(s)
   if any(x in rel.parts for x in ['adapted-tree','node_modules','.git','__pycache__']):continue
   if f.is_file() and not f.is_symlink():
    assert not f.name.endswith('.go'),str(f)
    files.add(f)
files=sorted(files);manifest=''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(s))+'\n' for p in files).encode();output.parent.mkdir(parents=True,exist_ok=True)
with tarfile.open(output,'w:gz') as tar:
 for p in files:tar.add(p,arcname=str(p.relative_to(s)),recursive=False)
 info=tarfile.TarInfo('SHA256SUMS');info.size=len(manifest);tar.addfile(info,io.BytesIO(manifest))
with tarfile.open(output,'r:gz') as tar:
 for line in tar.extractfile('SHA256SUMS').read().decode().splitlines():
  expected,_,name=line.partition('  ');assert hashlib.sha256(tar.extractfile(name).read()).hexdigest()==expected,name
sha=hashlib.sha256(output.read_bytes()).hexdigest();output.with_suffix(output.suffix+'.sha256').write_text(sha+'  '+output.name+'\n');print(json.dumps(dict(archive=output.name,files=len(files),bytes=output.stat().st_size,sha256=sha,internalManifestVerified=True)),flush=True)
