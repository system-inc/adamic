from pathlib import Path
import hashlib,json
out=Path(__file__).parent;a=Path('/workspace/scratch/path-components-upstream');b=Path('/workspace/scratch/path-components-after')
files=lambda root:sorted(p.relative_to(root).as_posix() for p in root.rglob('*') if p.is_file())
names=files(a/'src');assert names==files(b/'src')
changed=[f for f in names if (a/'src'/f).read_bytes()!=(b/'src'/f).read_bytes()]
assert changed==['compiler/path.ts'],changed
r=next(r for r in json.loads(Path('/workspace/adamic/stage3/adapt/40-explicit-any/class-rules.json').read_text())['brands'] if r['id']==103)
x=(a/r['file']).read_bytes();y=(b/r['file']).read_bytes();assert x.count(r['before'].encode())==1;assert x.replace(r['before'].encode(),r['after'].encode())==y
result=dict(files=len(names),changed=changed,onlyReviewedDeclarationChanged=True,before=hashlib.sha256(x).hexdigest(),after=hashlib.sha256(y).hexdigest())
(out/'source-scope.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
