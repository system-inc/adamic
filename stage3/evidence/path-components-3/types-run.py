import json,os,sys
from pathlib import Path
out=Path(__file__).parent;label=sys.argv[1]
tree=Path('/workspace/scratch/path-components-upstream' if label=='before' else '/workspace/scratch/path-components-after')
roots=sorted((tree/'src/compiler').rglob('*.ts'));assert len(roots)==79,len(roots)
(out/('roots-'+label+'.json')).write_text(json.dumps([str(p) for p in roots],indent=2)+'\n')
os.execv(str(out/'adamic'),[str(out/'adamic'),'types',*[str(p) for p in roots]])
