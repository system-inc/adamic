import subprocess,json,hashlib,shutil
from pathlib import Path
root=Path('/workspace/adamic');w=Path('/workspace/scratch/abrq61p-revision');e=root/'stage3/census/deferred-reach/evidence/revision';tool=(w/'tool-sha.txt').read_text().strip();old='8fea9f1018dfb124b0c77b4bac702e06f968680f'
results={}
for path in [root/'stage3/census/latent/make_overlay.py',*sorted((root/'stage3/census/latent').glob('*.go.txt'))]:
 key=str(path.relative_to(root));values=[]
 for sha in [old,tool]:
  p=subprocess.run(['git','show',sha+':'+key],cwd=root,capture_output=True)
  print('git show compiled source child exit:',p.returncode,flush=True);p.check_returncode();values.append(p.stdout)
 assert values[0]==values[1],key
 results[key]=hashlib.sha256(values[0]).hexdigest()
(e/'compiled-observer-identity.json').write_text(json.dumps(dict(previous_tool=old,revised_tool=tool,byte_identical=True,sources=results),indent=2)+'\n')
for name in ['report-audit.log','report-update.log']:shutil.copyfile(w/name,e/name)
shutil.copyfile(w/'finalize_report.py',e/'finalize_report.py')
(e/'SHA256.json').write_text(json.dumps({p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(e.iterdir()) if p.name!='SHA256.json'},indent=2)+'\n')
print('PASS: compiled observation sources unchanged; existing census and meter proofs retained')
