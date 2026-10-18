"""Hard-link identical immutable export archives; keep product paths and bytes."""
from pathlib import Path
import hashlib, os, json, time
root=Path('/home/agent/.cache/adamic-build/local');began=time.monotonic();seen={};rows=[];saved=0
for directory in sorted(root.iterdir()):
 if not (directory/'listing.json').is_file():continue
 for p in (directory/'exports').glob('*.a'):
  size=p.stat().st_size
  with p.open('rb') as f:digest=hashlib.file_digest(f,'sha256').hexdigest()
  key=(size,digest)
  if key in seen:
   first=seen[key]
   if first.stat().st_ino==p.stat().st_ino:continue
   temporary=p.with_suffix('.dedup');os.link(first,temporary);os.replace(temporary,p);saved+=size;rows.append(str(p.relative_to(root)))
  else:seen[key]=p
out=Path(__file__).parent/'cache-deduplication-final.json';out.write_text(json.dumps({'files':rows,'saved_bytes':saved,'seconds':time.monotonic()-began},indent=2)+'\n');print('deduplicated',len(rows),'archives; bytes',saved)
