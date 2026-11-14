"""Verify byte-for-byte adapted source identity and retain both manifests."""
import hashlib,json,subprocess
from pathlib import Path
BASE=Path('/workspace/scratch/c4v7-census');OUT=Path(__file__).resolve().parent
manifests={}
for label in ('main','head'):
 root=BASE/f'adapted-{label}'
 manifests[label]={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((root/'src').rglob('*')) if p.is_file()}
 files=[n for n in manifests[label] if n.startswith('src/compiler/') and n.endswith('.ts')];assert len(files)==79,len(files)
 (OUT/f'source-manifest-{label}.json').write_text(json.dumps(manifests[label],indent=2)+'\n')
a,b=manifests['main'],manifests['head'];differences=[n for n in sorted(a.keys()|b.keys()) if a.get(n)!=b.get(n)]
result=dict(src_files_main=len(a),src_files_head=len(b),compiler_ts_files=79,byte_identity=not differences,differences=differences,line_mapping='identity' if not differences else 'NOT IDENTITY')
(OUT/'source-identity.json').write_text(json.dumps(result,indent=2)+'\n');assert not differences,differences
# The baseline manifest covers compiler source and non-TS inputs, so compare its exact keys too.
old=json.loads(subprocess.check_output(['git','show','origin/typescript/next-stops-5c8338c4:stage3/census/next-stops-5c8338c4/adapted-source-manifest.json'],text=True))
assert all(b.get(n)==h for n,h in old.items()),'source differs from the baseline census'
result['baseline_manifest_identical']=True;(OUT/'source-identity.json').write_text(json.dumps(result,indent=2)+'\n')
ns=OUT.parent/'next-stops-c4v7'
compiler={n:h for n,h in b.items() if n.startswith('src/compiler/')}
for name in ['adapted-source-manifest.json','adapted-source-before-remaining.json']:(ns/name).write_text(json.dumps(compiler,indent=2)+'\n')
print(json.dumps(result,indent=2))
