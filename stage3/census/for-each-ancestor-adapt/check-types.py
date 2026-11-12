from pathlib import Path
import json,subprocess,sys
out=Path(__file__).resolve().parent;root=Path(sys.argv[1]);label=sys.argv[2]
roots=sorted(str(p)for p in (root/'src/compiler').rglob('*.ts'));assert len(roots)==79
(out/(label+'-roots.json')).write_text(json.dumps(roots,indent=2)+'\n')
for suffix,files in [('',roots),('-extended',roots+[str(root/'src/typingsInstallerCore/typingsInstaller.ts')])]:
 code=subprocess.run([sys.executable,str(out/'run.py'),str(out/(label+suffix)),'/workspace/scratch/ancestor-production/adamic','types',*files]).returncode
 print('types wrapper child exit',label+suffix,code,flush=True)
 if code:sys.exit(code)
