#!/usr/bin/env python3
"""Run one bounded foreground measurement and retain logs, metrics and descendants."""
import argparse
import json
import os
from pathlib import Path
import signal
import subprocess
import time
p=argparse.ArgumentParser();p.add_argument('output',type=Path);p.add_argument('--seconds',type=int,default=1440);p.add_argument('command',nargs=argparse.REMAINDER);a=p.parse_args()
assert 1<=a.seconds<1500 and a.command
command=a.command[1:] if a.command[0]=='--' else a.command
out=a.output.resolve();out.mkdir(parents=True,exist_ok=False)
start=time.monotonic()
with (out/'process.log').open('w') as log:
    child=subprocess.Popen(command,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
    try:
        code=child.wait(timeout=a.seconds)
    except subprocess.TimeoutExpired:
        code=124
    finally:
        try:os.killpg(child.pid,signal.SIGTERM)
        except ProcessLookupError:pass
        try:child.wait(timeout=5)
        except subprocess.TimeoutExpired:
            os.killpg(child.pid,signal.SIGKILL);child.wait()
metrics=dict(command=command,exit=code,seconds=time.monotonic()-start,limit_seconds=a.seconds)
(out/'metrics.json').write_text(json.dumps(metrics,indent=2)+'\n')
print(json.dumps(metrics),flush=True)
raise SystemExit(code)
