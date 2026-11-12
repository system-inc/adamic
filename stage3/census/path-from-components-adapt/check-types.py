import subprocess,sys
from pathlib import Path
label=sys.argv[1]; tree=Path('/workspace/scratch/path-adapt-'+label);out=Path('/workspace/scratch/path-adapt-proof')
roots=sorted(str(p) for p in (tree/'src/compiler').rglob('*.ts'))
assert len(roots)==79,len(roots)
r=subprocess.run([sys.executable,str(out/'run.py'),str(out/('types-'+label)),str(out/'adamic'),'types',*roots])
print('types supervisor child_exit='+str(r.returncode),flush=True);sys.exit(r.returncode)
