#!/usr/bin/env python3
"""Compare all adapted worktree content (directories, symlink targets, file bytes), excluding .git metadata."""
import os,sys,json,hashlib
from pathlib import Path
old,new,out=map(Path,sys.argv[1:4])
def manifest(root):
 records={};size=0
 for base,dirs,files in os.walk(root,followlinks=False):
  dirs[:]=sorted(d for d in dirs if not (Path(base)==root and d=='.git'))
  for name in sorted(dirs+files):
   p=Path(base)/name;rel=p.relative_to(root).as_posix()
   if rel=='.git':continue
   if p.is_symlink():records[rel]={'kind':'symlink','target':os.readlink(p)}
   elif p.is_dir():records[rel]={'kind':'directory'}
   elif p.is_file():
    n=p.stat().st_size
    with p.open('rb') as f:d=hashlib.file_digest(f,'sha256').hexdigest()
    records[rel]={'kind':'file','bytes':n,'sha256':d};size+=n
 digest=hashlib.sha256(''.join(json.dumps([k,records[k]],sort_keys=True,separators=(',',':'))+'\n' for k in sorted(records)).encode()).hexdigest()
 return records,{'entries':len(records),'regular_files':sum(r['kind']=='file' for r in records.values()),'regular_bytes':size,'manifest_sha256':digest}
a,am=manifest(old);b,bm=manifest(new)
changes=[{'path':p,'old':a.get(p),'new':b.get(p)} for p in sorted(set(a)|set(b)) if a.get(p)!=b.get(p)]
result={'old_tree':str(old),'new_tree':str(new),'definition':'All worktree directory names, symlink targets, and regular-file bytes; .git metadata excluded. Timestamps and permission bits are not byte content.','byte_identical':not changes,'old':am,'new':bm,'changed_entries':changes}
out.write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({'byte_identical':not changes,'old_files':am['regular_files'],'new_files':bm['regular_files'],'changed_paths':[c['path'] for c in changes]}))
