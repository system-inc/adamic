"""Losslessly retain raw logs (including their original whitespace) in Git."""
import gzip,hashlib,json
from pathlib import Path
out=Path(__file__).resolve().parent
records=[]
for p in sorted(out.glob('*.log')):
 data=p.read_bytes();compressed=gzip.compress(data,mtime=0)
 dest=p.with_suffix('.log.gz');dest.write_bytes(compressed)
 assert gzip.decompress(compressed)==data
 p.unlink()
 records.append(dict(original=p.name,compressed=dest.name,bytes=len(data),sha256=hashlib.sha256(data).hexdigest(),compressed_sha256=hashlib.sha256(compressed).hexdigest()))
(out/'compression.json').write_text(json.dumps(records,indent=2)+'\n')
