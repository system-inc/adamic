from pathlib import Path
import subprocess,sys,json
root=Path(sys.argv[1]);binary=sys.argv[2]
paths=sorted(str(p) for p in (root/'src/compiler').rglob('*.ts') if not p.name.endswith('.d.ts'))
p=subprocess.run([binary,'types',*paths],text=True,capture_output=True)
print(json.dumps({'command':[binary,'types',*paths],'exit':p.returncode,'roots':len(paths)},indent=2));print(p.stdout);print(p.stderr)
assert p.returncode==0
assert not p.stderr.strip(),p.stderr
