"""Bounded subprocess runner. Logs and wait status retain the child's exit status."""
import json, os, signal, resource, subprocess, sys, time
from pathlib import Path
out = Path(__file__).resolve().parent
name, cwd, *cmd = sys.argv[1:]
log = out / (name + '.log')
resource_file = out / (name + '.time.json')
started = time.monotonic()
with log.open('w') as stream:
    p = subprocess.Popen(cmd, cwd=cwd,
                         stdout=stream,stderr=subprocess.STDOUT,start_new_session=True)
    def stop(signum, frame):
        try: os.killpg(p.pid, signal.SIGTERM)
        except ProcessLookupError: pass
        try: p.wait(timeout=10)
        except subprocess.TimeoutExpired:
            os.killpg(p.pid, signal.SIGKILL); p.wait()
        print(name, 'child_exit='+str(p.returncode), flush=True)
        raise SystemExit(128+signum)
    signal.signal(signal.SIGTERM, stop)
    signal.signal(signal.SIGINT, stop)
    (out/(name+".pid.json")).write_text(json.dumps(dict(pid=p.pid,started_epoch=time.time(),command=cmd))+"\n")
    last = time.monotonic(); stamp = None; stopped = None
    while p.poll() is None:
        current = [(f.name,f.stat().st_size,f.stat().st_mtime_ns) for f in out.glob(name+'*') if f.is_file()]
        if current != stamp:
            last=time.monotonic(); stamp=current
        if time.monotonic()-last>1800:
            stopped='no output for 1800 seconds'
            os.killpg(p.pid,signal.SIGTERM)
            try:p.wait(timeout=10)
            except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL)
            break
        time.sleep(1)
    code=p.wait()
    try: os.killpg(p.pid,signal.SIGKILL)
    except ProcessLookupError: pass
usage=resource.getrusage(resource.RUSAGE_CHILDREN)
resource_file.write_text(json.dumps(dict(peak_rss_kib=usage.ru_maxrss,user_seconds=usage.ru_utime,system_seconds=usage.ru_stime),indent=2)+'\n')
row=dict(command=cmd,cwd=cwd,exit=code,wall_seconds=round(time.monotonic()-started,3),stopped=stopped)
(out/(name+'.status.json')).write_text(json.dumps(row,indent=2)+'\n')
print(name, 'child_exit='+str(code), 'wall_seconds='+str(row['wall_seconds']),flush=True)
sys.exit(code if code>=0 else 128-code)
