import sys,json,hashlib,difflib
from pathlib import Path
base,head,out=map(Path,sys.argv[1:]);rows=[];missing=[];new=[]
def sha(b):return hashlib.sha256(b).hexdigest()
for root in ['built/local','lib']:
 for p in sorted((head/root).rglob('*.js')):
  rel=p.relative_to(head);q=base/rel
  if not q.exists():new.append(str(rel));continue
  a=q.read_bytes();b=p.read_bytes();offset=next((i for i,(x,y) in enumerate(zip(a,b)) if x!=y),min(len(a),len(b))) if a!=b else None
  rows.append({'file':str(rel),'before_bytes':len(a),'after_bytes':len(b),'before_sha256':sha(a),'after_sha256':sha(b),'byte_identical':a==b,'first_difference_offset':offset})
  if a!=b:
   (out.parent/(str(rel).replace('/','-')+'.diff')).write_text(''.join(difflib.unified_diff(a.decode(errors='replace').splitlines(keepends=True),b.decode(errors='replace').splitlines(keepends=True),fromfile='base/'+str(rel),tofile='candidate/'+str(rel))))
 for p in sorted((base/root).rglob('*.js')):
  if not (head/p.relative_to(base)).exists():missing.append(str(p.relative_to(base)))
r={'compared':len(rows),'different':sum(not x['byte_identical'] for x in rows),'baseline_only_not_built_by_local_task':missing,'candidate_only':new,'rows':rows};out.write_text(json.dumps(r,indent=2)+'\n');print('Emitted JS compared='+str(len(rows))+' different='+str(r['different']));sys.exit(bool(r['different'] or new))
