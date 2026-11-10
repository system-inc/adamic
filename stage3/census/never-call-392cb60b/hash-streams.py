from pathlib import Path
import gzip,hashlib,json
out=Path(__file__).resolve().parent;rows=[]
for p in sorted(out.glob('*.jsonl*')):
 if p.name.endswith('.pending'):continue
 h=hashlib.sha256();size=0
 with (gzip.open(p,'rb') if p.suffix=='.gz' else p.open('rb')) as f:
  while chunk:=f.read(1024*1024):h.update(chunk);size+=len(chunk)
 rows.append(dict(file=p.name,stored_bytes=p.stat().st_size,uncompressed_bytes=size,stored_sha256=hashlib.sha256(p.read_bytes()).hexdigest(),uncompressed_sha256=h.hexdigest()))
(out/'stream-hashes.json').write_text(json.dumps(rows,indent=2)+'\n')
print('stream hash child exit code: 0')
