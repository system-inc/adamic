"""Copy evidence only, retaining compressed bytes and a hash manifest."""
import gzip,hashlib,json,shutil
from pathlib import Path
out=Path('/workspace/scratch/995spev')
dest=out/'evidence-branch/stage3/census/commonjs-bundle-evidence'
dest.mkdir(parents=True,exist_ok=True)
manifest={}
for source in sorted(out.iterdir()):
 if not source.is_file() or source.suffix not in ['.json','.jsonl','.log','.txt','.md','.py','.go']:continue
 if source.name in ['pack-status.json','pack.log']:continue
 raw=source.read_bytes();compress=source.suffix=='.log' or source.name.endswith('-sys.jsonl')
 name=source.name+('.txt' if source.suffix=='.go' else '')+('.gz' if compress else '')
 data=gzip.compress(raw,mtime=0) if compress else raw
 if compress:assert gzip.decompress(data)==raw
 (dest/name).write_bytes(data)
 manifest[name]=dict(sha256=hashlib.sha256(data).hexdigest(),bytes=len(data),uncompressed_sha256=hashlib.sha256(raw).hexdigest())
(dest/'artifact-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print('evidence_files='+str(len(manifest))+' bytes='+str(sum(row['bytes'] for row in manifest.values())))
