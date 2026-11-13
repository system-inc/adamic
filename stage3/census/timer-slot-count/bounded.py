#!/usr/bin/env python3
"""Bound a child; print its actual return code (including signals), never hide it."""
import json,os,signal,subprocess,sys
limit=float(sys.argv[1])
child=subprocess.Popen(sys.argv[2:],start_new_session=True)
timed_out=False
try:
    code=child.wait(timeout=limit)
except subprocess.TimeoutExpired:
    timed_out=True
    os.killpg(child.pid,signal.SIGTERM)
    try:code=child.wait(timeout=10)
    except subprocess.TimeoutExpired:
        os.killpg(child.pid,signal.SIGKILL)
        code=child.wait()
print(json.dumps({'command':sys.argv[2:],'childReturncode':code,'timedOut':timed_out}),file=sys.stderr,flush=True)
sys.exit(code if code>=0 else 128-code)
