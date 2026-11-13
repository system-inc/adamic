#!/usr/bin/env python3
"""Run one bounded child, retain its log, and explicitly report its exit code."""
import subprocess,sys,json,time,os,signal
from pathlib import Path
out=Path(sys.argv[1]);out.parent.mkdir(parents=True,exist_ok=True)
start=time.monotonic()
with out.with_suffix('.log').open('w') as log:
    child=subprocess.Popen(sys.argv[2:],stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
    try: code=child.wait(timeout=1700)
    except subprocess.TimeoutExpired:
        os.killpg(child.pid,signal.SIGTERM)
        try: child.wait(timeout=5)
        except subprocess.TimeoutExpired: os.killpg(child.pid,signal.SIGKILL);child.wait()
        code=124
result={'command':sys.argv[2:],'child_exit':child.returncode,'wrapper_exit':code,'seconds':round(time.monotonic()-start,3)}
out.with_suffix('.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
sys.exit(code if code>=0 else 128-code)
