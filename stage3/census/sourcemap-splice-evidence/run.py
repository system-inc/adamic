#!/usr/bin/env python3
import subprocess,sys,json,time,os
from pathlib import Path
out=Path(__file__).resolve().parent
name=sys.argv[1];cmd=sys.argv[2:];start=time.time()
with (out/(name+'.log.txt')).open('wb') as log:
 p=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
row=dict(name=name,command=cmd,cwd=os.getcwd(),exit=p.returncode,seconds=time.time()-start,environment={k:v for k,v in os.environ.items() if k.startswith('LATENT_') or k in ('ADAMIC_TYPESCRIPT_SOURCE','NODE_PATH','NODE_OPTIONS','SM_COUNT_DIR')})
with (out/'commands.jsonl').open('a') as f:f.write(json.dumps(row)+'\n')
print(json.dumps(row),flush=True)
sys.exit(p.returncode)
