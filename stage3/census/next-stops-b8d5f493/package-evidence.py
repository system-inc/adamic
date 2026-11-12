"""Compress large evidence files and hash the final directory contents."""
import gzip,hashlib,json,shutil
from pathlib import Path
out=Path(__file__).resolve().parent
compressed=[]
for p in sorted(out.rglob('*')):
 if '__pycache__' in p.parts or not p.is_file() or p.suffix=='.gz' or p.stat().st_size<=1_000_000:continue
 dest=p.with_name(p.name+'.gz')
 with p.open('rb') as source,dest.open('wb') as raw:
  with gzip.GzipFile(filename='',mode='wb',fileobj=raw,mtime=0) as target:shutil.copyfileobj(source,target)
 compressed.append({'file':str(p.relative_to(out)),'bytes':p.stat().st_size,'gzip':str(dest.relative_to(out))})
 p.unlink()
(out/'compression.json').write_text(json.dumps(compressed,indent=2)+'\n')
manifest=out/'SHA256SUMS'
manifest.write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(out))+'\n' for p in sorted(out.rglob('*')) if p.is_file() and p!=manifest and '__pycache__' not in p.parts))
print(json.dumps({'compressed':compressed,'files':len(manifest.read_text().splitlines())}))
