import hashlib,json,subprocess,sys,time
from pathlib import Path
out=Path(__file__).resolve().parent; label,tree=sys.argv[1:];tree=Path(tree)
roots=sorted(tree.joinpath('src/compiler').rglob('*.ts'));assert len(roots)==79
cmd=[str(out/'adamic'),'types',*map(str,roots)]; start=time.monotonic()
with (out/(label+'-types.log')).open('w') as log:
 r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
receipt=dict(command=cmd,exit=r.returncode,seconds=time.monotonic()-start,root_count=len(roots),sys_sha256=hashlib.sha256((tree/'src/compiler/sys.ts').read_bytes()).hexdigest())
(out/(label+'-types-receipt.json')).write_text(json.dumps(receipt,indent=2)+'\n');print(json.dumps({k:receipt[k] for k in ['exit','seconds','root_count']}));sys.exit(r.returncode)
