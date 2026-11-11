from pathlib import Path
import subprocess,sys,json,time,os
S=Path(__file__).resolve().parent
name=sys.argv[1]; args=sys.argv[2:]; start=time.monotonic()
with (S/(name+'.log')).open('w') as log:
 p=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT)
r=dict(command=args,cwd=os.getcwd(),exit=p.returncode,seconds=time.monotonic()-start)
(S/(name+'.json')).write_text(json.dumps(r,indent=2)+'\n');print(json.dumps(r),flush=True);sys.exit(p.returncode)
