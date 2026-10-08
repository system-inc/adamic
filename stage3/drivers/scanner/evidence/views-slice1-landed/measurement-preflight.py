import subprocess,json,pathlib
repo=pathlib.Path('/workspace/adamic');out=pathlib.Path('/workspace/scratch/scanner-views-landed-candidates');out.mkdir()
def text(args,cwd=repo):return subprocess.check_output(args,cwd=cwd,text=True).strip()
def run(args,label,cwd=repo):
 with (out/(label+'.stdout')).open('wb') as a,(out/(label+'.stderr')).open('wb') as b:return subprocess.run(args,cwd=cwd,stdout=a,stderr=b).returncode
base=text(['git','rev-parse','origin/main']);tree=out/'tree';assert run(['git','worktree','add','-b','scratch/scanner-views-landed-candidates',str(tree),base],'worktree')==0
rows=[];accepted=[base]
for i,ref in enumerate(['origin/codex/records-maplike-next','origin/codex/maplike-records','origin/library/area-on-next']):
 sha=text(['git','rev-parse',ref]);code=run(['git','merge','--no-edit',sha],f'merge-{i}',tree);conflicts=text(['git','diff','--name-only','--diff-filter=U'],tree).splitlines() if code else []
 rows.append(dict(ref=ref,sha=sha,exit=code,conflicts=conflicts))
 if code:
  assert conflicts;assert run(['git','merge','--abort'],f'abort-{i}',tree)==0
 else:accepted.append(sha)
 print(ref,code,len(conflicts),flush=True)
(out/'stack.json').write_text(json.dumps(dict(base=base,merges=rows,accepted=accepted,selection='Tried requested records-maplike-next and newer descendant maplike-records (contains 8c013f1c, not rebased onto 031a1259) before library.'),indent=2)+'\n')
