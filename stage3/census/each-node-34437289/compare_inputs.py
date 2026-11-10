from control import *
import hashlib,difflib
left=root/'adapted-parent'; right=root/'adapted-head'
manifest={}; differences=[]
for p in sorted((left/'src').rglob('*')):
 if not p.is_file(): continue
 rel=p.relative_to(left); q=right/rel
 a=p.read_bytes(); b=q.read_bytes()
 manifest[str(rel)]={'parent':hashlib.sha256(a).hexdigest(),'head':hashlib.sha256(b).hexdigest()}
 if a!=b:
  differences.append(str(rel))
  (root/(p.name+'.source-diff.txt')).write_text(''.join(difflib.unified_diff(a.decode().splitlines(True),b.decode().splitlines(True),fromfile='parent/'+str(rel),tofile='head/'+str(rel))))
(root/'source-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
(root/'source-differences.json').write_text(json.dumps(differences,indent=2)+'\n')
print(differences)
