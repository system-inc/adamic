import hashlib,json
from pathlib import Path
out=Path(__file__).parent;a=Path('/workspace/scratch/path-components-upstream');b=Path('/workspace/scratch/path-components-after')
hash=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
files=lambda root:sorted(p.relative_to(root).as_posix() for p in root.rglob('*') if p.is_file())
assert files(a/'src')==files(b/'src')
changes=[f for f in files(a/'src') if (a/'src'/f).read_bytes()!=(b/'src'/f).read_bytes()]
assert changes==['compiler/path.ts'],changes
old=b'export type PathPathComponents = Path[] & { __pathComponensBrand: undefined; };'
new=b'export interface PathPathComponents<T extends Path = Path> extends Array<T> { __pathComponensBrand: undefined; }'
x=(a/'src/compiler/path.ts').read_bytes();y=(b/'src/compiler/path.ts').read_bytes();assert x.count(old)==1;assert x.replace(old,new)==y
names=['_tsc.js','_tsserver.js','_typingsInstaller.js','tsc.js','tsserver.js','tsserverlibrary.js','typescript.js','typingsInstaller.js','watchGuard.js','run.js']
actual=[f for f in files(a/'built/local') if f.endswith(('.js','.cjs','.mjs'))]
assert sorted(names)==actual,actual
assert actual==[f for f in files(b/'built/local') if f.endswith(('.js','.cjs','.mjs'))]
js=[]
for name in names:
 p=a/'built/local'/name;q=b/'built/local'/name;assert p.read_bytes()==q.read_bytes(),name
 js.append(dict(file=name,bytes=p.stat().st_size,sha256=hash(p)))
api=[]
for directory in ['lib','built/local']:
 for name in ['typescript.d.ts','tsserverlibrary.d.ts']:
  p=a/directory/name;q=b/directory/name;assert p.read_bytes()==q.read_bytes(),str(p)
  api.append(dict(file=directory+'/'+name,bytes=p.stat().st_size,sha256=hash(p)))
stocka=json.loads((out/'stock-before.json').read_text());stockb=json.loads((out/'stock-after.json').read_text());assert stocka==stockb
result=dict(changedSources=changes,onlyReviewedDeclarationChanged=True,javascript=js,publicDeclarations=api,stockDiagnostics=len(stocka['diagnostics']),roots=79)
(out/'comparison.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
