from control import *
import hashlib
exclude={'checkpoint.lock','debug-report-draft.md','final-report.md'}
raw={}
for p in root.glob('*.jsonl'):
 data=p.read_bytes(); raw[p.name]={'uncompressed_sha256':hashlib.sha256(data).hexdigest(),'uncompressed_bytes':len(data),'delivered_as':p.name+('.gz' if len(data)>1000000 else '')}
(root/'raw-stream-hashes.json').write_text(json.dumps(raw,indent=2)+'\n')
for p in dest.iterdir():
 if p.is_file(): p.unlink()
for p in root.iterdir():
 if not p.is_file() or p.name in exclude: continue
 data=p.read_bytes(); name=p.name
 if name.endswith(".go.work"): name=name.removesuffix(".go.work")+"-workspace.go.txt"
 if len(data)>1000000: data=gzip.compress(data,mtime=0); name+='.gz'
 (dest/name).write_bytes(data)
(dest/'REPORT.md').write_bytes((root/'final-report.md').read_bytes())
for p in dest.rglob('*'):
 if p.is_file():
  assert not p.name.endswith('.go'),str(p)
  assert p.stat().st_size<=1000000 or p.name.endswith('.gz'),str(p)
sha=[]
for p in sorted(dest.rglob('*')):
 if p.is_file(): sha.append(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(dest)))
(dest/'SHA256SUMS').write_text('\n'.join(sha)+'\n')
print('artifacts',len(sha),'bytes',sum(p.stat().st_size for p in dest.rglob('*') if p.is_file()))
