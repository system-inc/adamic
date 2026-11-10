import json, os, pathlib, signal, subprocess, sys, time, resource
name=sys.argv[1]; command=sys.argv[2:]; out=pathlib.Path(__file__).parent
log=out/(name+'.log.txt'); status=out/(name+'.status.json')
start=time.monotonic(); last=start; size=-1
with log.open('w') as f:
 p=subprocess.Popen(command,stdout=f,stderr=subprocess.STDOUT,start_new_session=True)
 stopped=None
 try:
  while p.poll() is None:
   current=[(str(x),x.stat().st_size,x.stat().st_mtime_ns) for x in sorted(out.glob(name+'*')) if x.is_file()]
   if current!=size: size=current; last=time.monotonic()
   if time.monotonic()-last>1800:
    stopped='silent past 30 minutes'; os.killpg(p.pid,signal.SIGTERM)
    try:p.wait(timeout=10)
    except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL)
    break
   time.sleep(2)
  code=p.wait()
 finally:
  try:os.killpg(p.pid,signal.SIGTERM)
  except ProcessLookupError:pass
r=resource.getrusage(resource.RUSAGE_CHILDREN)
result=dict(command=command,child_exit=code,stopped=stopped,wall_seconds=time.monotonic()-start,peak_rss_kib=r.ru_maxrss,user_seconds=r.ru_utime,system_seconds=r.ru_stime)
status.write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
sys.exit(code if 0<=code<=255 else 1)
