"""Hash the validated CommonJSBundle evidence without running another format."""
import gzip,hashlib,json
from pathlib import Path
base=Path(__file__).resolve().parent
v=json.loads((base/'CommonJSBundle/validation.json').read_text())
assert v['roots']==79 and v['observations']==158 and v['duplicates_agree']
for name,info in json.loads((base/'compression.json').read_text()).items():
 p=base/(name+'.gz')
 if p.exists():assert hashlib.sha256(gzip.decompress(p.read_bytes())).hexdigest()==info['sha256']
manifest={str(p.relative_to(base)):dict(bytes=p.stat().st_size,sha256=hashlib.sha256(p.read_bytes()).hexdigest()) for p in sorted(base.rglob('*')) if p.is_file() and p.name!='artifact-manifest.json' and '__pycache__' not in p.parts}
assert all(d['bytes']<=20_000_000 for d in manifest.values())
(base/'artifact-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print('CommonJSBundle artifact manifest verified')
