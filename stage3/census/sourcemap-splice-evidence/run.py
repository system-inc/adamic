#!/usr/bin/env python3
import subprocess,sys,json,time,os
from pathlib import Path
out=Path(__file__).resolve().parent
name=sys.argv[1];cmd=sys.argv[2:];start=time.time()
with (out/(name+'.log')).open('wb') as log:
 p=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
row=dict(name=name,command=cmd,cwd=os.getcwd(),exit=p.returncode,seconds=time.time()-start)
with (out/'commands.jsonl').open('a') as f:f.write(json.dumps(row)+'\n')
print(json.dumps(row),flush=True)
sys.exit(p.returncode)
