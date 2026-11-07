"""Record actual child exits, bound silence, and reap each process group."""
import sys
sys.dont_write_bytecode = True
import json, os, pathlib, signal, subprocess, time
SCRATCH=pathlib.Path('/workspace/scratch/getallkeys-proof')
SCRATCH.mkdir(parents=True,exist_ok=True)
def run(name,command,cwd='/workspace/adamic',env=None,watch=(),limit=3600):
    started=time.monotonic(); last=started; previous=None; stopped=None
    paths=[SCRATCH/(name+'.log'),*[pathlib.Path(p) for p in watch]]
    with paths[0].open('wb') as log:
        child=subprocess.Popen(command,cwd=cwd,env=env,stdout=log,stderr=log,start_new_session=True)
        try:
            while child.poll() is None:
                signature=[(p.stat().st_size,p.stat().st_mtime_ns) if p.exists() else None for p in paths]
                now=time.monotonic()
                if signature!=previous:previous=signature;last=now
                if now-last>=1800 or now-started>=limit:
                    stopped='30 minutes without output' if now-last>=1800 else 'elapsed-time bound'
                    os.killpg(child.pid,signal.SIGTERM)
                    try:child.wait(timeout=10)
                    except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);child.wait()
                    break
                time.sleep(1)
            code=child.wait()
        finally:
            try:os.killpg(child.pid,signal.SIGKILL)
            except ProcessLookupError:pass
            child.wait()
    record={'command':command,'cwd':cwd,'exit_code':code,'stopping_reason':stopped,'seconds':time.monotonic()-started,'process_group_cleaned':True}
    (SCRATCH/(name+'-status.json')).write_text(json.dumps(record,indent=2)+'\n')
    print(name+' child_exit='+str(code),flush=True)
    return record
if __name__=='__main__':
    import sys
    record=run(sys.argv[1],sys.argv[2:])
    raise SystemExit(0 if record['exit_code']==0 and not record['stopping_reason'] else 1)
