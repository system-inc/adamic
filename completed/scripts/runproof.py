import subprocess as s,sys,os,json,time
from pathlib import Path
r=Path(__file__).resolve().parent
name,expected,*cmd=sys.argv[1:];expected=int(expected)
env=dict(os.environ,NODE_PATH='/workspace/scratch/watch-file-1532-merge/stage3/api/node_modules',TMPDIR='/workspace/scratch/watch-file-proof/tmp',GOTMPDIR='/workspace/scratch/watch-file-proof/tmp',GOCACHE='/workspace/scratch/watch-file-proof/go-cache',GOMODCACHE='/workspace/scratch/watch-file-proof/go-mod',STAGE3_CACHE=str(r/'cache'),ADAMIC_TYPESCRIPT_SOURCE='/workspace/scratch/watch-file-typescript')
env['PATH']='/workspace/adamic-tools/bin:/workspace/adamic-tools/go/bin:'+env['PATH']
start=time.time()
with (r/'logs'/(name+'.log')).open('w') as f:
 try: code=s.run(cmd,env=env,stdout=f,stderr=s.STDOUT,timeout=1800).returncode
 except s.TimeoutExpired:code=124
row=dict(name=name,command=cmd,cwd=os.getcwd(),exit=code,expected=expected,seconds=round(time.time()-start,3),log=str(r/'logs'/(name+'.log')))
with (r/'commands.jsonl').open('a') as f:f.write(json.dumps(row)+'\n')
print(json.dumps(row),flush=True)
sys.exit(0 if code==expected else 1)
