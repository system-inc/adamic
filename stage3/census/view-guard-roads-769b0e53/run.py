#!/usr/bin/env python3
import json, os, signal, subprocess, sys, time
from pathlib import Path
prefix=Path(sys.argv[1]); command=sys.argv[2:]; prefix.parent.mkdir(parents=True,exist_ok=True)
started=time.monotonic()
with prefix.with_suffix('.log').open('wb') as log:
    child=subprocess.Popen(command,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
    def stop(sig,frame):
        os.killpg(child.pid,signal.SIGTERM)
    signal.signal(signal.SIGTERM,stop); signal.signal(signal.SIGINT,stop)
    pid,status,usage=os.wait4(child.pid,0)
    code=os.waitstatus_to_exitcode(status); child.returncode=code
    try: os.killpg(child.pid,signal.SIGKILL)
    except ProcessLookupError: pass
result={'command':command,'cwd':os.getcwd(),'child_exit':code,'elapsed_seconds':time.monotonic()-started,'peak_rss_kib':usage.ru_maxrss,'user_seconds':usage.ru_utime,'system_seconds':usage.ru_stime,'environment':{k:v for k,v in os.environ.items() if k.startswith(('LATENT_','ROADS_GUARD_')) or k in ('ADAMIC_TYPESCRIPT_SOURCE','GOMAXPROCS','GOMEMLIMIT')}}
prefix.with_suffix('.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result),flush=True)
sys.exit(code if code>=0 else 128-code)
