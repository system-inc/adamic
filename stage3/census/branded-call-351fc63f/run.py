"""Bounded subprocess runner. Logs and wait status retain the child's exit status."""
import hashlib, json, os, re, signal, resource, subprocess, sys, time
from pathlib import Path
out = Path(__file__).resolve().parent
name, cwd, *cmd = sys.argv[1:]
# A completed full-core stream may be reused by the long-running entry runner.
# Preserve the original child's log, resources and exit status; no new child runs.
core_names={f'{a}-{b}-core' for a in ['A','B'] for b in ['parent','head']}
focus_name=bool(re.fullmatch(r'[AB]-(parent|head)-(focus-[1-4]|serializer)',name))
status=out/(name+'.status.json')
binary=cmd[-3] if name in core_names else cmd[cmd.index('-project')-1] if focus_name else None
binary_hash=hashlib.sha256(Path(binary).read_bytes()).hexdigest() if binary else None
if binary and status.exists():
    old=json.loads(status.read_text())
    stream=Path(cmd[-1])
    if old['exit']==0 and old['command']==cmd and old['cwd']==cwd and old.get('binary_sha256')==binary_hash and stream.exists():
        records=[json.loads(line) for line in stream.read_text().splitlines()]
        assert records[0]['status']=='measurement' and len(records)==2
        if name in core_names:
            assert records[0]['latent_mode']=='full' and records[1]['file']==cmd[-2]+'/core.ts'
        else:
            assert records[0]['scope']=='focused-full-unit' and records[0]['selector']==cmd[cmd.index('-where')+1] and len(records[1]['units'])==1
        print(name,'reused recorded_child_exit=0; no new child',flush=True)
        sys.exit(0)
log = out / (name + '.log')
resource_file = out / (name + '.time.json')
started = time.monotonic()
with log.open('w') as stream:
    p = subprocess.Popen(cmd, cwd=cwd,
                         stdout=stream,stderr=subprocess.STDOUT,start_new_session=True)
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
usage=resource.getrusage(resource.RUSAGE_CHILDREN)
resource_file.write_text(json.dumps(dict(peak_rss_kib=usage.ru_maxrss,user_seconds=usage.ru_utime,system_seconds=usage.ru_stime),indent=2)+'\n')
row=dict(command=cmd,cwd=cwd,binary_sha256=binary_hash,exit=code,wall_seconds=round(time.monotonic()-started,3),stopped=stopped)
(out/(name+'.status.json')).write_text(json.dumps(row,indent=2)+'\n')
print(name, 'child_exit='+str(code), 'wall_seconds='+str(row['wall_seconds']),flush=True)
sys.exit(code if code>=0 else 128-code)
