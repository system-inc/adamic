"""Ordinary census, one file root per child, at most 20 roots per checkpoint."""
import hashlib,json,os,signal,subprocess,sys,time
from pathlib import Path
out=Path(__file__).resolve().parent
repo=Path('/workspace/scratch/adamic-f44b5c27')
scratch=Path('/workspace/scratch')
tree=scratch/'next-stops-f44b5c27-adapted'
def run(name,args,limit=1780):
    start=time.monotonic()
    with (out/(name+'.log')).open('wb') as log:
        child=subprocess.Popen(args,cwd=repo,stdout=log,stderr=log,start_new_session=True)
        stopped=None
        try:
            try: rc=child.wait(timeout=limit)
            except subprocess.TimeoutExpired:
                stopped='elapsed bound'
                os.killpg(child.pid,signal.SIGTERM)
                try: rc=child.wait(timeout=10)
                except subprocess.TimeoutExpired:
                    os.killpg(child.pid,signal.SIGKILL);rc=child.wait()
        finally:
            try: os.killpg(child.pid,signal.SIGKILL)
            except ProcessLookupError: pass
            child.wait()
    status=dict(command=args,exit_code=rc,seconds=time.monotonic()-start,stopping_reason=stopped,process_group_cleaned=True)
    (out/(name+'-status.json')).write_text(json.dumps(status,indent=2)+'\n')
    print(name+' child_exit='+str(rc),flush=True)
    if rc: raise SystemExit(rc if rc>0 else 128-rc)
    return status
if __name__=='__main__':
    run(sys.argv[1],sys.argv[2:])
