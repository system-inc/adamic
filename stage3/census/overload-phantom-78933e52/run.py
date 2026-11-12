"""Bound every child; report and propagate its own exit status."""
import sys,subprocess,time,resource,json,os,signal
log,stats,*cmd=sys.argv[1:]
t=time.monotonic(); timed_out=False
with open(log,'wb') as out:
 p=subprocess.Popen(cmd,stdout=out,stderr=subprocess.STDOUT,start_new_session=True)
 try: rc=p.wait(timeout=1740)
 except subprocess.TimeoutExpired:
  timed_out=True;os.killpg(p.pid,signal.SIGTERM)
  try: rc=p.wait(timeout=10)
  except subprocess.TimeoutExpired: os.killpg(p.pid,signal.SIGKILL);rc=p.wait()
r=resource.getrusage(resource.RUSAGE_CHILDREN)
d=dict(command=cmd,exit_code=rc,timed_out=timed_out,wall_seconds=time.monotonic()-t,user_seconds=r.ru_utime,system_seconds=r.ru_stime,peak_rss_kib=r.ru_maxrss)
open(stats,'w').write(json.dumps(d,indent=2)+'\n')
print('child_exit='+str(rc),flush=True)
sys.exit(rc if rc>=0 else 128-rc)
