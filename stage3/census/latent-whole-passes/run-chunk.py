#!/usr/bin/env python3
"""One bounded child per wrapper. Records and propagates that child's status."""
import json, os, pathlib, resource, signal, subprocess, sys, time
binary, entry, dest, start, end, mode = sys.argv[1:]
dest=pathlib.Path(dest);dest.mkdir(parents=True,exist_ok=True)
env=dict(os.environ,LATENT_FULL='0',LATENT_ASSERT_NO_OUTPUT='1',LWP_START=start,LWP_END=end,
         LATENT_TIMES=str(dest/'units.jsonl'),LWP_TIMES=str(dest/'passes.jsonl'),
         LWP_PASSES='1' if mode=='prototype' else '')
deadline=float(os.environ.get("LWP_DEADLINE_SECONDS","1700"))
if not 0 < deadline <= 1700: raise ValueError("deadline must be in (0,1700]")
t=time.monotonic()
with (dest/'stderr.log').open('w') as log:
 child=subprocess.Popen([binary,entry,str(dest/'census.jsonl')],env=env,stdout=log,stderr=log,start_new_session=True)
 expired=False
 try: code=child.wait(timeout=deadline)
 except subprocess.TimeoutExpired:
  expired=True;os.killpg(child.pid,signal.SIGTERM)
  try: code=child.wait(timeout=5)
  except subprocess.TimeoutExpired: os.killpg(child.pid,signal.SIGKILL);code=child.wait()
row={'argv':[binary,entry],'start':int(start),'end':int(end),'mode':mode,'skip_refusal_scan':env.get('LWP_SKIP_SCAN')=='1','child_exit_code':code,'deadline_expired':expired,'wall_seconds':time.monotonic()-t,'peak_rss_kib':resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss}
(dest/'status.json').write_text(json.dumps(row,indent=2)+'\n')
print(json.dumps(row),flush=True)
if code<0:
 if -code not in (signal.SIGKILL,signal.SIGSTOP): signal.signal(-code,signal.SIG_DFL)
 os.kill(os.getpid(),-code)
sys.exit(code)
