from pathlib import Path
import subprocess,os,json,time,sys
root=Path('/workspace/adamic');raw=Path('/workspace/scratch/tracing-length-evidence')
def run(name,args,cwd=root,env=None,required=True):
 start=time.monotonic()
 with (raw/(name+'.log')).open('w') as log:
  p=subprocess.run([str(x) for x in args],cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT)
 result=dict(command=[str(x) for x in args],cwd=str(cwd),exit=p.returncode,seconds=time.monotonic()-start)
 (raw/(name+'.json')).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
 if required and p.returncode:sys.exit(p.returncode)
 return p.returncode
if __name__=='__main__':
 run(sys.argv[1],sys.argv[2:])
