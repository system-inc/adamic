from pathlib import Path
import subprocess,json,hashlib,shutil
s=Path('/workspace/scratch/diag-read-view-3');r=s/'repo';e=Path('/workspace/scratch/diag-read-view-1/evidence');d=e/'stage3/census/diag-read-view-3';d.mkdir(parents=True,exist_ok=True)
for name in ['prove.py','replay.cjs','checkpoint.py','api-install.log','submodules.log']:
 if (s/name).exists():shutil.copyfile(s/name,d/name)
for f in subprocess.check_output(['git','diff-tree','--no-commit-id','--name-only','-r','HEAD'],cwd=r,text=True).splitlines():
 p=d/'implementation'/f;p.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(r/f,p)
(d/'record-only.patch').write_bytes(subprocess.check_output(['git','diff','d7a6aa49976828aa9627b64c137234b718c9e503','HEAD'],cwd=r))
for p in (s/'raw').rglob('*'):
 if p.is_file():
  target=d/p.relative_to(s/'raw');target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(p,target)
status={'status':'PROOF_PENDING','implementation':subprocess.check_output(['git','rev-parse','HEAD'],cwd=r,text=True).strip(),'base':'8e3a784aa5a601f57342c07b68f145fe56252c21','v2':'d7a6aa49976828aa9627b64c137234b718c9e503','note':'Only the requested ledger append and README wording differ from v2. Zero-write replay, fresh apply and the whole lane are rerun. No Loom submission.','results':{p.name:json.loads(p.read_text()) for p in (s/'raw').glob('*.result.json')}}
if (s/'raw/FINAL.json').exists():status.update(json.loads((s/'raw/FINAL.json').read_text()))
(d/'STATUS.md').write_text(json.dumps(status,indent=2)+'\n')
(d/'manifest.json').write_text(json.dumps({str(p.relative_to(d)):hashlib.sha256(p.read_bytes()).hexdigest() for p in d.rglob('*') if p.is_file() and p.name!='manifest.json'},indent=2)+'\n')
for cmd in [['git','add','-f','stage3/census/diag-read-view-3'],['git','commit','-m','evidence: diagnostic read-view record fixes and reruns'],['git','push','origin','HEAD:refs/heads/typescript/diag-read-view-evidence']]:subprocess.run(cmd,cwd=e,check=True)
