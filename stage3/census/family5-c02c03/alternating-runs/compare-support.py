from pathlib import Path
import json,hashlib
W=Path('/tmp/family5-alternating');a=Path('/workspace/family5-main-lane-alone/adapted-tree');b=Path('/workspace/family5-c02-lane/adapted-tree')
files=['Herebyfile.mjs','package.json','package-lock.json','scripts/build/tests.mjs','scripts/build/utils.mjs','scripts/build/projects.mjs']
rows=[]
for f in files:
 x=(a/f).read_bytes();y=(b/f).read_bytes();rows.append({'file':f,'byte_identical':x==y,'bytes':len(x),'main_sha256':hashlib.sha256(x).hexdigest(),'c02_sha256':hashlib.sha256(y).hexdigest()})
# Cover externally-required runner dependencies, including executable native modules.
def inv(root):return {str(p.relative_to(root/'node_modules')):p for p in (root/'node_modules').rglob('*') if p.is_file() and (p.suffix in ['.js','.cjs','.mjs','.json','.node','.wasm'] or '/.bin/' in str(p))}
x=inv(a);y=inv(b);d=[]
for f in sorted(x.keys()|y.keys()):
 p=x.get(f);q=y.get(f);u=p.read_bytes() if p else b'';v=q.read_bytes() if q else b''
 if not p or not q or u!=v:d.append({'file':f,'main_bytes':len(u) if p else None,'c02_bytes':len(v) if q else None,'first_difference_offset_0based':next((i for i,(c,e) in enumerate(zip(u,v)) if c!=e),min(len(u),len(v)))})
r={'build_and_launch_inputs':rows,'external_dependency_files_compared':len(x.keys()|y.keys()),'external_dependency_differences':d}
(W/'support-comparison.json').write_text(json.dumps(r,indent=2)+'\n');print('Launch/build support files identical:',all(i['byte_identical'] for i in rows));print('External executable/config dependency files:',r['external_dependency_files_compared'],'differences:',len(d))
for z in d[:10]:print(z)
