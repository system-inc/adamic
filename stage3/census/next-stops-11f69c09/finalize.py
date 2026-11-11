import sys,json,subprocess,hashlib,tarfile
from pathlib import Path
out=Path(__file__).resolve().parent
sys.path.insert(0,str(out));from run import run
sys.path.insert(0,'/workspace/adamic/stage3/meter');from report import READINESS_REASONS
rows=[json.loads(x) for x in (out/'census.jsonl').read_text().splitlines()]
roots=[r['roots'][0] for r in rows if r.get('reason') in READINESS_REASONS]
if not (out/'node-readiness-status.json').exists():run('node-readiness',[sys.executable,'stage3/meter/node_readiness.py','/workspace/scratch/next-stops-adapted',str(out/'node-readiness'),*roots])
run('artifact-classification',[sys.executable,str(out/'classify-artifacts.py')])
run('latent-totals',[sys.executable,str(out/'latent-totals.py')])
run('source-excerpts',[sys.executable,str(out/'source-excerpts.py')])
binaries={}
for name,filename in [('ordinary','next-stops-census'),('guarded_entry','next-stops-entry'),('latent','next-stops-latent')]:
 p=Path('/workspace/scratch')/filename
 binaries[name]={'path':str(p),'sha256':hashlib.sha256(p.read_bytes()).hexdigest()}
 with (out/(name+'-buildinfo.txt')).open('w') as f:
  child=subprocess.run(['go','version','-m',str(p)],stdout=f,stderr=f)
  print(name+' buildinfo child_exit='+str(child.returncode),flush=True);child.check_returncode()
(out/'binaries.json').write_text(json.dumps(binaries,indent=2)+'\n')
tree=Path('/workspace/scratch/next-stops-adapted')
(out/'apply-source-manifest.json').write_text(json.dumps({str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src/compiler').rglob('*')) if p.is_file()},indent=2)+'\n')
with tarfile.open(out/'guarded-entry-overlay.tar.gz','w:gz') as t:
 for d in ['next-stops-overlay','next-stops-entry-overlay']:t.add(Path('/workspace/scratch')/d,arcname=d)
run('write-report',[sys.executable,str(out/'write-report.py')])
(out/'PROGRESS.md').unlink(missing_ok=True)
run('validation',[sys.executable,str(out/'validate.py')])
