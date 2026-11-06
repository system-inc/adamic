#!/usr/bin/env python3
"""Pack selected regular evidence files and list every member and archive SHA-256."""
import argparse
import gzip
import hashlib
import io
from pathlib import Path
import tarfile
import os

parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('archive',type=Path)
parser.add_argument('sha256',type=Path)
parser.add_argument('root',type=Path)
parser.add_argument('paths',nargs='+')
parser.add_argument('--prefix',default='raw')
args=parser.parse_args()
root=args.root.resolve();selected=set()
for name in args.paths:
 path=(root/name).resolve();assert path.is_relative_to(root)
 for file in path.rglob('*') if path.is_dir() else [path]:
  if file.is_file() and not file.is_symlink():selected.add(file)
assert selected
args.archive.parent.mkdir(parents=True,exist_ok=True)
temporary=args.archive.with_name(args.archive.name+'.tmp');hashes={}
with temporary.open('wb') as raw,gzip.GzipFile(fileobj=raw,mode='wb',filename='',mtime=0) as compressed,tarfile.open(fileobj=compressed,mode='w|') as tar:
 for file in sorted(selected):
  data=file.read_bytes();name=args.prefix+'/'+file.relative_to(root).as_posix();info=tarfile.TarInfo(name);info.size=len(data);info.mode=0o644;info.mtime=0
  tar.addfile(info,io.BytesIO(data));hashes[name]=hashlib.sha256(data).hexdigest()
os.replace(temporary,args.archive)
with tarfile.open(args.archive,'r:gz') as tar:
 members=tar.getmembers();assert len(members)==len(hashes)
 for member in members:assert hashlib.sha256(tar.extractfile(member).read()).hexdigest()==hashes[member.name]
archive_hash=hashlib.sha256(args.archive.read_bytes()).hexdigest()
args.sha256.write_text(archive_hash+'  '+args.archive.name+'\n'+''.join(hashes[name]+'  '+name+'\n' for name in sorted(hashes)))
print('Packed',len(hashes),'files;',args.archive.name,archive_hash)
