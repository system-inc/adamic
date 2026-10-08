import subprocess,json,pathlib
repo=pathlib.Path('/workspace/adamic'); out=pathlib.Path('/workspace/scratch/scanner-rehearsal-preflight');out.mkdir()
def run(args,label,cwd=repo):
 with (out/(label+'.stdout')).open('wb') as a,(out/(label+'.stderr')).open('wb') as b:
  return subprocess.run(args,cwd=cwd,stdout=a,stderr=b).returncode
def text(args,cwd=repo):return subprocess.check_output(args,cwd=cwd,text=True).strip()
refs=['origin/compiler/area-stack','origin/codex/records-maplike-next','origin/area/library','origin/codex/placeholder-nonnull','origin/codex/entries-provenance','origin/codex/assignment-proofs','origin/codex/scanner-cast-checks','origin/codex/stricter-options-next']
base=text(['git','rev-parse','origin/main']); tree=out/'tree'
assert run(['git','worktree','add','-b','scratch/scanner-rehearsal-preflight',str(tree),base],'worktree')==0
rows=[];accepted=[base]
for i,ref in enumerate(refs):
 sha=text(['git','rev-parse',ref]); code=run(['git','merge','--no-edit',sha],f'merge-{i}',tree)
 conflicts=text(['git','diff','--name-only','--diff-filter=U'],tree).splitlines() if code else []
 rows.append(dict(ref=ref,sha=sha,exit=code,conflicts=conflicts))
 if code:
  assert conflicts,rows[-1]
  assert run(['git','merge','--abort'],f'abort-{i}',tree)==0
 else:accepted.append(sha)
 print(ref,code,len(conflicts),flush=True)
(out/'stack.json').write_text(json.dumps(dict(base=base,merges=rows,accepted=accepted,missing=['library/area-on-next-2']),indent=2)+'\n')
env=__import__('os').environ.copy();env['STAGE3_CACHE']='/workspace/scratch/native3-cache';env['GOPROXY']='https://proxy.golang.org|direct'
with open('/tmp/scanner-rehearsal-run.log','wb') as log:
 code=subprocess.run(['bash','stage3/drivers/scanner/scratch-run.sh','/workspace/scratch/scanner-rehearsal-run',*accepted],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT).returncode
print('scratch-run exit',code,flush=True)
