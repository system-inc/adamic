import json,difflib,hashlib
from pathlib import Path
base=Path('/workspace/scratch/brands2/control-tree');after=Path('/workspace/scratch/brands3/retyped-tree');out=Path('/workspace/adamic/stage3/adapt/40-explicit-any/evidence/brands-undefined-3')
h=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
js=lambda root:{str(p.relative_to(root)):h(p) for p in (root/'built').rglob('*') if p.is_file() and p.suffix in ['.js','.mjs','.cjs']}
a,b=js(base),js(after);assert a==b,'JavaScript file names or bytes changed'
rows=[]
for label,file in [('built','built/local/typescript.d.ts'),('reference','tests/baselines/reference/api/typescript.d.ts')]:
 old=(base/file).read_text();new=(after/file).read_text();diff=list(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='control/'+file,tofile='retyped/'+file));removed=[s[1:] for s in diff if s.startswith('-') and not s.startswith('---')];added=[s[1:] for s in diff if s.startswith('+') and not s.startswith('+++')]
 assert len(removed)==len(added)==27
 assert [s.replace(': void',': undefined') for s in removed]==added
 (out/(label+'-api.diff')).write_text(''.join(diff));rows.append({'artifact':file,'changed_lines':27,'only':'void to undefined','before_sha256':h(base/file),'after_sha256':h(after/file)})
(out/'api-runtime.json').write_text(json.dumps({'javascript_files':len(a),'javascript_bytes_identical':True,'apis':rows},indent=2)+'\n')
(out/'javascript-hashes.json').write_text(json.dumps(a,indent=2)+'\n')
print(json.dumps({'javascript_files':len(a),'public_api_lines':27}))
