import sys,subprocess,time,json,os
from pathlib import Path
out=Path(__file__).resolve().parent
name=sys.argv[1];cmd=sys.argv[2:];start=time.time()
with (out/(name+'.log')).open('w') as f:
 p=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
r={'name':name,'cwd':os.getcwd(),'command':cmd,'exit':p.returncode,'seconds':time.time()-start,'env':{k:v for k,v in os.environ.items() if k.startswith(('LATENT_','CENSUS_','ADAMIC_','STAGE3_')) or k in ['NODE_PATH','GOCACHE','GOPATH','TMPDIR']}}
with (out/'commands.jsonl').open('a') as f:f.write(json.dumps(r)+'\n')
print(json.dumps(r),flush=True)
sys.exit(p.returncode)
