from pathlib import Path
import subprocess,shutil,json,hashlib,gzip
s=Path('/workspace/scratch/diag-read-view-2');r=s/'repo';e=Path('/workspace/scratch/diag-read-view-1/evidence');d=e/'stage3/census/diag-read-view-2';d.mkdir(parents=True,exist_ok=True)
sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=r,text=True).strip()
for f in subprocess.check_output(['git','diff-tree','--no-commit-id','--name-only','-r','HEAD'],cwd=r,text=True).splitlines():
 p=d/'implementation'/f;p.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(r/f,p)
for name in ['prepare.py','add-helpers.cjs','run-proof.py','pair.py','controls.cjs','stock.cjs','javascript.cjs','checkpoint.py']:
 if (s/name).exists():shutil.copyfile(s/name,d/name)
for p in (s/'raw').iterdir():
 if p.is_file() and not p.name.endswith('.jsonl'):shutil.copyfile(p,d/p.name)
for name in ['submodules.log','api-install.log']:
 if (s/name).exists():shutil.copyfile(s/name,d/name)
for p in (s/'raw').glob('*.jsonl'):
 try:complete=json.loads(p.read_text().splitlines()[-1]).get('status')=='relation_report'
 except:complete=False
 if complete:
  with gzip.open(d/(p.name+'.gz'),'wb') as out:out.write(p.read_bytes())
status={'implementation':sha,'base':'8e3a784aa5a601f57342c07b68f145fe56252c21','upstream':'050880ce59e30b356b686bd3144efe24f875ebc8','status':'PROOF_PENDING','note':'v2 applies all eight reader parameter substitutions. User accepts exactly the pre-existing renamed pair; no other new or changed findings are accepted. No Loom submission.','completedResults':{p.name:json.loads(p.read_text()) for p in (s/'raw').glob('*.result.json')}}
if (s/'raw/FINAL.json').exists():status.update(json.loads((s/'raw/FINAL.json').read_text()))
(d/'STATUS.md').write_text(json.dumps(status,indent=2)+'\n')
(d/'manifest.json').write_text(json.dumps({str(p.relative_to(d)):hashlib.sha256(p.read_bytes()).hexdigest() for p in d.rglob('*') if p.is_file() and p.name!='manifest.json'},indent=2)+'\n')
for cmd in [['git','add','-f','stage3/census/diag-read-view-2'],['git','commit','-m','evidence: checkpoint complete diagnostic reader proof'],['git','push','origin','HEAD:refs/heads/typescript/diag-read-view-evidence']]:subprocess.run(cmd,cwd=e,check=True)
