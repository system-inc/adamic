import subprocess,resource,time,json,sys,os,signal,fcntl
from pathlib import Path
name=sys.argv[1]; argv=sys.argv[2:]; d=Path('/workspace/scratch/recordreplay')
lock=(d/'measurement.lock').open('a'); fcntl.flock(lock,fcntl.LOCK_EX)
start=time.monotonic()
with (d/(name+'.log')).open('wb') as log:
 p=subprocess.Popen(argv,stdout=log,stderr=log,start_new_session=True)
 try: rc=p.wait(timeout=int(os.environ.get('RUN_LIMIT','3600')))
 except subprocess.TimeoutExpired:
  os.killpg(p.pid,signal.SIGTERM)
  try: rc=p.wait(timeout=10)
  except subprocess.TimeoutExpired: os.killpg(p.pid,signal.SIGKILL);rc=p.wait()
r=dict(argv=argv,cwd=os.getcwd(),exit=rc,seconds=time.monotonic()-start,peak_rss_kib=resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss)
(d/(name+'.json')).write_text(json.dumps(r,indent=2)+'\n');print(json.dumps(r));sys.exit(rc if rc>=0 else 128-rc)
