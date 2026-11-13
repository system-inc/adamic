import subprocess,sys,time,json
from pathlib import Path
name,*cmd=sys.argv[1:]
p=Path(__file__).parent
t=time.monotonic()
with (p/(name+'.stdout')).open('w') as out,(p/(name+'.stderr')).open('w') as err:
 r=subprocess.run(['timeout','1700',*cmd],stdout=out,stderr=err)
(p/(name+'.status.json')).write_text(json.dumps(dict(command=cmd,cwd=str(Path.cwd()),child_exit=r.returncode,seconds=time.monotonic()-t),indent=2)+'\n')
sys.exit(r.returncode)
