import subprocess,json,re,hashlib
from pathlib import Path
def git(*x):return subprocess.check_output(['git',*x],text=True)
root=Path('/workspace/adamic/stage1/cohere/typeaware');old=json.loads((root/'validation-wave-11-fifth/fifth-selection.json').read_text());names={x['rule'] for x in old['ranking']};excluded=set(old['baseline_ports']);refs=git('for-each-ref','--format=%(refname)','refs/remotes/origin').splitlines();trees={}
for r in refs:trees.setdefault(git('rev-parse',r+'^{tree}').strip(),r)
claims=[];seen=set()
for tree,ref in trees.items():
 for path in git('ls-tree','-r','--name-only',tree,'stage1/cohere/typeaware/claims/').splitlines():
  if not path.endswith('.md'):continue
  s=git('show',tree+':'+path)
  if path.endswith('/wave-11.md'):s=s.split('## Sixth batch claim')[0]
  digest=hashlib.sha256(s.encode()).hexdigest()
  if (path,digest) in seen:continue
  seen.add((path,digest));found=sorted(n for n in names if re.search('(?<![a-zA-Z0-9_/@-])'+re.escape(n)+'(?![a-zA-Z0-9_/@-])',s));excluded.update(found)
  claims.append({'branch':ref,'path':path,'sha256':digest,'names':found})
ranking=[dict(x,excluded=x['rule'] in excluded) for x in old['ranking']];selected=[x['rule'] for x in ranking if not x['excluded']][:3]
assert selected==['react-hooks/set-state-in-effect','react-hooks/set-state-in-render','react-hooks/static-components'],selected
out=dict(origin_refs=len(refs),distinct_trees=len(trees),main=git('rev-parse','origin/main').strip(),bridge=git('rev-parse','origin/codex/tsgo-c-library').strip(),baseline_ports=old['baseline_ports'],claims=claims,ranking=ranking,selected=selected,note='Own claim trimmed before sixth batch to reconstruct reservation snapshot. All named reservations excluded, including skipped and released entries. Main and bridge matches are inventories, not ports.')
Path('/workspace/wave-11-logs/sixth-selection.json').write_text(json.dumps(out,indent=2)+chr(10));print(len(refs),len(trees),len(claims),selected)
