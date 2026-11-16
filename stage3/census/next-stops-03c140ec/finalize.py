"""Validate compressed logs, record checkpoints and hash all evidence."""
import gzip,hashlib,json,subprocess
from pathlib import Path
base=Path(__file__).resolve().parent
assert json.loads((base/'validation.json').read_text())['raw_observations']==316
for name,info in json.loads((base/'compression.json').read_text()).items():
 p=base/(name+'.gz');assert p.exists()
 assert hashlib.sha256(p.read_bytes()).hexdigest()==info['compressed_sha256']
 assert hashlib.sha256(gzip.decompress(p.read_bytes())).hexdigest()==info['sha256']
p=json.loads((base/'provenance.json').read_text())
checkpoints=[dict(sha=line.split(' ',1)[0],subject=line.split(' ',1)[1]) for line in subprocess.check_output(['git','log','--reverse','--format=%H %s',p['sha']+'..HEAD'],text=True).splitlines() if 'checkpoint' in line]
(base/'checkpoints.json').write_text(json.dumps(checkpoints,indent=2)+'\n')
manifest={str(p.relative_to(base)):dict(bytes=p.stat().st_size,sha256=hashlib.sha256(p.read_bytes()).hexdigest()) for p in sorted(base.rglob('*')) if p.is_file() and p.name!='artifact-manifest.json' and '__pycache__' not in p.parts}
assert all(d['bytes']<=20_000_000 for d in manifest.values())
(base/'artifact-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print('finalized',len(manifest),'artifacts')
