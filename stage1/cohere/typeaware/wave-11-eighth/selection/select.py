import subprocess,json,re,hashlib
from pathlib import Path
root=Path('/workspace/adamic');directory=root/'stage1/cohere/typeaware'
def git(*args):return subprocess.check_output(['git',*args],cwd=root,text=True)
old=json.loads((directory/'validation-wave-11-sixth/sixth-selection.json').read_text()); ranking=[row for row in old['ranking'] if row['rule']!='findings'];names={row['rule'] for row in ranking};excluded=set(old['baseline_ports'])
refs=git('for-each-ref','--format=%(refname)','refs/remotes/origin').splitlines();trees={}
for ref in refs:trees.setdefault(git('rev-parse',ref+'^{tree}').strip(),ref)
claims=[];seen=set()
for tree,ref in trees.items():
 for row in git('ls-tree','-r',tree,'stage1/cohere/typeaware/claims/').splitlines():
  metadata,path=row.split('\t',1);blob=metadata.split()[2]
  if blob in seen:continue
  seen.add(blob)
  if not path.endswith(('.md','.json','.txt')):continue
  source=git('show',blob)
  if path.endswith('.json'):
   document=json.loads(source)
   reservation={key:document[key] for key in ('claimed','selected','rules','reservations') if key in document}
   if not reservation:continue
   source=json.dumps(reservation)
  found=sorted(name for name in names if re.search(r'(?<![a-zA-Z0-9_/@-])'+re.escape(name)+r'(?![a-zA-Z0-9_/@-])',source))
  excluded.update(found);claims.append(dict(ref=ref,path=path,blob=blob,names=found))
available=[row['rule'] for row in ranking if row['rule'] not in excluded]
output=dict(origin_refs=len(refs),distinct_trees=len(trees),main=git('rev-parse','origin/main').strip(),bridge=git('rev-parse','origin/codex/tsgo-c-library').strip(),claims=claims,ranking=ranking,baseline_ports=old['baseline_ports'],available=available)
Path('/workspace/wave-11-logs/parked-selection.json').write_text(json.dumps(output,indent=2)+'\n');print(json.dumps(dict(origin_refs=len(refs),distinct_trees=len(trees),claim_blobs=len(claims),available=available),indent=2))
