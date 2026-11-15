from pathlib import Path
import hashlib,json,time
s=Path('/workspace/loose-slot-scratch');r=s/'head';d=r/'stage3/census/loose-slot-census'
for which in ['head','main']:
 log=s/('apply-'+which+'.log');expected=str(s/('adapted-'+which))
 while not log.exists() or log.read_text().splitlines()[-1]!=expected:time.sleep(5)
 (s/('apply-'+which+'.done')).write_text('completed successfully\n')
 print(which,'apply complete',flush=True)
manifest=json.loads((d/'evidence/provenance.json').read_text())
for which in ['main','head']:
 tree=s/('adapted-'+which)/'src/compiler'
 manifest[which]['files']={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(tree.rglob('*.ts'))}
 assert len(manifest[which]['files'])==79,len(manifest[which]['files'])
manifest['adapted_identical']=manifest['main']['files']==manifest['head']['files'];assert manifest['adapted_identical']
(d/'evidence/provenance.json').write_text(json.dumps(manifest,indent=2)+'\n')
print('79 roots, identical',flush=True)
