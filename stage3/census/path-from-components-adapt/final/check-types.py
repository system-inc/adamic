from pathlib import Path
import subprocess,sys
out=Path('/workspace/scratch/path-adapt-final');tree=Path(sys.argv[1]);label=sys.argv[2]
roots=sorted(str(p) for p in (tree/'src/compiler').rglob('*.ts'));assert len(roots)==79,len(roots)
r=subprocess.run([sys.executable,str(out/'run.py'),str(out/label),str(out/'adamic'),'types',*roots])
print('types supervisor child_exit='+str(r.returncode),flush=True);sys.exit(r.returncode)
