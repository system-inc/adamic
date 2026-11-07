import os,subprocess,signal,shutil,json,time
from pathlib import Path
s=Path('/workspace/scratch/core-extend');repo=Path('/workspace/adamic');old=s/'f027-evidence';old.mkdir()
rows=[r.strip().split(None,2) for r in subprocess.check_output(['ps','-eo','pid,ppid,args'],text=True).splitlines()[1:]];roots={int(p) for p,pp,args in rows if args in ['python3 /workspace/scratch/core-extend/count-pieces.py','python3 /workspace/scratch/core-extend/pack-pieces.py']};ids=set(roots)
while True:
 expanded=ids|{int(p) for p,pp,_ in rows if int(pp) in ids}
 if expanded==ids:break
 ids=expanded
for pid in sorted(ids,reverse=True):
 try:os.kill(pid,signal.SIGKILL)
 except ProcessLookupError:pass
for p in list(s.iterdir()):
 if p==old or p.name in ['recut-prepare.py','recut-prepare.log']:continue
 shutil.move(str(p),str(old/p.name))
# Keep helper programs available; raw earlier evidence remains under f027-evidence.
for p in [*old.glob('*.py'),*old.glob('*.cjs')]:shutil.copyfile(p,s/p.name)
head=subprocess.check_output(['git','rev-parse','origin/main'],cwd=repo,text=True).strip();(s/'recut-base.json').write_text(json.dumps({'base':head,'priorBase':'f0271c6a7','reason':'origin/main advanced with getOwnValues/record-table amendments'},indent=2)+'\n')
with (s/'recut-worktree.log').open('w') as log:subprocess.run(['git','worktree','add','--detach',str(s/'main'),head],cwd=repo,stdout=log,stderr=subprocess.STDOUT,check=True)
(s/'main/stage3/api/node_modules').symlink_to(repo/'stage3/api/node_modules',target_is_directory=True)
# All commands on the new tip use its adaptation stack, parser and lane scripts.
for name in ['apply.py','apply.sh','census','lane','oracle']:
 pass
print('recut base',head,flush=True)
