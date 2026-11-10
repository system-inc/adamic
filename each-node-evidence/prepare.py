from pathlib import Path
import subprocess,os,time,json
s=Path('/workspace/scratch/each-node');repo=Path('/workspace/adamic');env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE=str(s/'upstream'))
def run(name,args):
 start=time.monotonic()
 with (s/(name+'.log')).open('w') as f:p=subprocess.run(args,cwd=repo,env=env,stdout=f,stderr=subprocess.STDOUT)
 result=dict(command=args,exit=p.returncode,seconds=time.monotonic()-start);(s/(name+'-result.json')).write_text(json.dumps(result,indent=2)+'\n');print(name,json.dumps(result),flush=True)
 if p.returncode:raise SystemExit(p.returncode)
run('prepare',['python3','stage3/lane/prepare_shards.py',str(s/'artifact-cache')])
