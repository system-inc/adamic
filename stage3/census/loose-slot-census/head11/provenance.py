from pathlib import Path
import json,hashlib,subprocess,gzip
out=Path(__file__).resolve().parent;s=Path('/workspace/loose-slot-scratch/eleven');r={}
for which in ['base','head']:
 repo=s/which;tree=s/('adapted-'+which)/'src/compiler';files={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest()for p in sorted(tree.rglob('*.ts'))};assert len(files)==79
 r[which]={'sha':subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip(),'files':files,'submodules':subprocess.check_output(['git','submodule','status','--recursive'],cwd=repo,text=True)}
 r[which]['binary_sha256']=hashlib.sha256((s/(which+'-census')).read_bytes()).hexdigest()
r['inputs_identical']=r['head']['files']==r['base']['files'];old=json.loads((out.parent/'evidence/provenance.json').read_text());r['inputs_identical_to_minus7']=r['head']['files']==old['head']['files'];assert r['inputs_identical']and r['inputs_identical_to_minus7']
r['type_binary_sha256']=hashlib.sha256((s/'type-count').read_bytes()).hexdigest();r['probe_source_sha256']=hashlib.sha256((out/'type_count.go.txt').read_bytes()).hexdigest();r['typescript_source']=subprocess.check_output(['git','rev-parse','HEAD'],cwd=s.parent/'typescript',text=True).strip()
(out/'evidence/provenance.json').write_text(json.dumps(r,indent=2)+'\n')
diff=subprocess.check_output(['git','diff','8843e0ab^','8843e0ab'],cwd=s/'head');(out/'evidence/head-parent.diff.gz').write_bytes(gzip.compress(diff,mtime=0))
for p in (out/'evidence').rglob('*.log'):
 compressed=p.with_name(p.name+'.gz')
 if p.stat().st_size:compressed.write_bytes(gzip.compress(p.read_bytes(),mtime=0))
 elif compressed.exists():compressed.unlink()
print('provenance verified')
