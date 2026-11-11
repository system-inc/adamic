from pathlib import Path
import subprocess,sys
S=Path(__file__).resolve().parent;side=sys.argv[1];tree=Path('/workspace/scratch/dp24r5m-'+side);roots=sorted((tree/'src/compiler').rglob('*.ts'));assert len(roots)==79,len(roots)
p=subprocess.run(['python3',S/'run.py',side+'-types',S/'adamic','types',*roots]);print('types wrapper exit:',p.returncode,flush=True);sys.exit(p.returncode)
