from pathlib import Path
import subprocess,shutil,json,hashlib,gzip
s=Path('/workspace/scratch/diag-read-view-1');r=s/'repo';e=s/'evidence';d=e/'stage3/census/diag-read-view';d.mkdir(parents=True,exist_ok=True)
sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=r,text=True).strip()
files=subprocess.check_output(['git','diff-tree','--no-commit-id','--name-only','-r','HEAD'],cwd=r,text=True).splitlines()
for f in files:
 p=d/'implementation'/f;p.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(r/f,p)
for name in ['run-proof.py','pair.py','controls.cjs','stock.cjs','javascript.cjs','make-pins.cjs','implement.py','route-owners.py','amend-95.py','amend-import-owner.py','stop-type-facts.cjs','write-stop.py','checkpoint.py']:
 shutil.copyfile(s/name,d/name)
for p in (s/'raw').iterdir():
 if p.is_file() and not p.name.endswith('.jsonl'):shutil.copyfile(p,d/p.name)
for name in ['submodules.log','api-install.log','clone.log','main-apply.log','overlay.log','latent-build.log','adamic-build.log','main-npm-ci.log']:
 if (s/name).exists():shutil.copyfile(s/name,d/name)
# Record only complete census streams under their final names.
for p in (s/'raw').glob('*.jsonl'):
 try:
  rows=p.read_text().splitlines();complete=json.loads(rows[-1]).get('status')=='relation_report'
 except:complete=False
 if complete:
  with gzip.open(d/(p.name+'.gz'),'wb') as out:out.write(p.read_bytes())
status={'implementation':sha,'base':'8e3a784aa5a601f57342c07b68f145fe56252c21','upstream':'050880ce59e30b356b686bd3144efe24f875ebc8','note':'Proof checkpoint; only recorded completed results are claims. Census/apply/lane acceptance remains pending until final verdict. No Loom submission.','completed':[p.name for p in d.glob('*.result.json')]}
if (s/'raw/STOP.json').exists():
 status['status']='STOPPED_UNACCEPTED'
 status['note']='Stopped at the strict full-census changed-finding gate. No successful final apply or whole lane is claimed. No Loom submission.'
 status['stop']=json.loads((s/'raw/STOP.json').read_text())
(d/'STATUS.md').write_text(json.dumps(status,indent=2)+'\n')
(d/'ownership-grep.log').write_text(subprocess.run(['rg','-l','getDiagnosticsHelper|sortAndDeduplicateDiagnostics|compareDiagnostics|compareRelatedInformation|compareMessageText|diagnosticsEqualityComparer|getDiagnosticFilePath|getDiagnosticCode|getDiagnosticMessage|DiagnosticWithLocation|"owner": "createProgram"','stage3/adapt','--glob','!**/evidence/**','--glob','!**/fold-evidence/**'],cwd=r,text=True,stdout=subprocess.PIPE).stdout)
(d/'manifest.json').write_text(json.dumps({str(p.relative_to(d)):hashlib.sha256(p.read_bytes()).hexdigest() for p in d.rglob('*') if p.is_file() and p.name!='manifest.json'},indent=2)+'\n')
for cmd in [['git','add','-f','stage3/census/diag-read-view'],['git','commit','-m','evidence: stop diagnostic read-view at optional-widening census' if (s/'raw/STOP.json').exists() else 'evidence: checkpoint diagnostic amendment proofs'],['git','push','origin','HEAD:refs/heads/typescript/diag-read-view-evidence']]:subprocess.run(cmd,cwd=e,check=True)
