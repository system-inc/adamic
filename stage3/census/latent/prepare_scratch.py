"""Reproduce the resolved comparisons on new never-pushed scratch branches."""
import json, pathlib, subprocess, sys
repository=pathlib.Path(sys.argv[1]).resolve(); scratch=pathlib.Path(sys.argv[2]).resolve()
assert str(scratch).startswith('/tmp/'), 'scratch trees only'
scratch.mkdir(parents=True,exist_ok=True)
prefix=sys.argv[3] if len(sys.argv)>3 else 'scratch/latent-reproduce-'
assert prefix.startswith('scratch/latent-')
root=pathlib.Path(__file__).resolve().parent
provenance=json.loads((root/'REPORT.json').read_text())
main=provenance['runs'][0]['main']
preparation=['origin/codex/stage3-base','origin/codex/tsc-census','origin/codex/stage3-type-imports','origin/codex/stage3-optional-declarations']
# Use the recorded preparation commits too, so reproduction cannot drift with origin refs.
preparation_shas=provenance['preparation_shas']
rows=[]
def command(args,tree,log):
 with log.open('w') as output:return subprocess.run(args,cwd=tree,stdout=output,stderr=subprocess.STDOUT).returncode
for spec in provenance['runs']:
 name=spec['name'];tree=scratch/name
 assert command(['git','worktree','add','-b',prefix+name,str(tree),main],repository,scratch/(name+'-worktree.log'))==0
 refs=list(preparation_shas.values())+list(spec['features'].values())
 for index,ref in enumerate(refs):
  status=command(['git','merge','--no-edit',ref],tree,scratch/(name+f'-merge-{index}.log'))
  if status:
   conflicts=subprocess.check_output(['git','diff','--name-only','--diff-filter=U'],cwd=tree,text=True)
   assert conflicts, 'merge failed without conflicts'
   assert command(['python3',str(root/'resolve_scratch.py'),str(tree)],tree,scratch/(name+f'-resolve-{index}.log'))==0
   assert command(['git','commit','-m','Resolve census scratch integration'],tree,scratch/(name+f'-commit-{index}.log'))==0
 rows.append(dict(name=name,tree=str(tree),features=list(spec['features']),preparation_shas=preparation_shas))
 (scratch/'worktrees.json').write_text(json.dumps(rows,indent=2)+'\n')
 print(name,'resolved',flush=True)
