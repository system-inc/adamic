#!/usr/bin/env python3
import difflib, hashlib, json, pathlib, subprocess, sys
before,after,out=map(pathlib.Path,sys.argv[1:])
out.mkdir(parents=True,exist_ok=True)
rel=pathlib.Path('src/compiler/transformers/declarations.ts')
def inventory(tree):
    return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (tree/'src').rglob('*') if p.is_file()}
a=inventory(before);b=inventory(after)
assert a.keys()==b.keys()
changed=[p for p in a if a[p]!=b[p]]
assert changed==[str(rel)],changed
old=(before/rel).read_text();new=(after/rel).read_text()
(out/'source.diff').write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='before/'+str(rel),tofile='after/'+str(rel))))
# The full apply snapshot immediately before 99e must have exactly control's src bytes.
snapshot='refs/stage3/99-correctness-axolotl-e-before'
for name,digest in a.items():
    data=subprocess.check_output(['git','-C',str(after),'show',snapshot+':'+name])
    assert hashlib.sha256(data).hexdigest()==digest,name
patch=subprocess.check_output(['git','-C',str(after),'diff','--numstat',snapshot,'refs/stage3/99-correctness-axolotl-e-after'],text=True)
assert patch.strip()=='4\t2\t'+str(rel),patch
result=dict(files=len(a),changed=changed,controlEqualsPre99eSnapshot=True,incrementalNumstat=patch.strip(),before=a[str(rel)],after=b[str(rel)])
(out/'scope.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))
