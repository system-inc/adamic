#!/usr/bin/env python3
"""Pack raw receipts with both per-member and archive SHA-256 lists."""
import gzip
import hashlib
from pathlib import Path
import tarfile
import sys
source=Path(sys.argv[1]).resolve();target=Path(sys.argv[2]).resolve();target.parent.mkdir(parents=True,exist_ok=True)
files=sorted(f for f in source.rglob('*') if f.is_file())
with target.open('wb') as raw,gzip.GzipFile(fileobj=raw,mode='wb',mtime=0) as compressed,tarfile.open(fileobj=compressed,mode='w') as archive:
 for f in files:
  info=archive.gettarinfo(str(f),arcname=str(f.relative_to(source)));info.uid=info.gid=info.mtime=0;info.uname=info.gname=''
  with f.open('rb') as stream:archive.addfile(info,stream)
target.with_suffix('').with_suffix('.sha256').write_text(''.join(hashlib.sha256(f.read_bytes()).hexdigest()+'  '+str(f.relative_to(source))+'\n' for f in files))
Path(str(target)+'.sha256').write_text(hashlib.sha256(target.read_bytes()).hexdigest()+'  '+target.name+'\n')
print(str(target),len(files),target.stat().st_size)
