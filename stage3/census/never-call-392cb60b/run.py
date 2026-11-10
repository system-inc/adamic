import os,sys,time,json,subprocess,signal,pathlib,resource
name,cwd,out,*cmd=sys.argv[1:]
out=pathlib.Path(out);out.mkdir(parents=True,exist_ok=True)
start=time.monotonic(); started=time.strftime('%Y-%m-%dT%H:%M:%SZ',time.gmtime())
with (out/(name+'.log')).open('wb') as log:
 p=subprocess.Popen(cmd,cwd=cwd,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 try:
  last=time.monotonic(); previous=None; stopped=False
  while p.poll() is None:
   paths=sorted(out.glob(name+'.*'))
   stamp=[(q.stat().st_size,q.stat().st_mtime_ns) if q.exists() else None for q in paths]
   if stamp!=previous:last=time.monotonic();previous=stamp
   if time.monotonic()-last>1800:
    stopped=True;os.killpg(p.pid,signal.SIGTERM)
    try:p.wait(timeout=10)
    except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL)
    break
   time.sleep(1)
  rc=p.wait()
 finally:
  try:os.killpg(p.pid,signal.SIGTERM)
  except ProcessLookupError:pass
usage=resource.getrusage(resource.RUSAGE_CHILDREN)
row=dict(name=name,command=cmd,cwd=cwd,started_utc=started,seconds=time.monotonic()-start,peak_rss_kib=usage.ru_maxrss,child_exit_code=rc,silent_cutoff=stopped)
(out/(name+'.timing.json')).write_text(json.dumps(row,indent=2)+'\n')
print(json.dumps(row),flush=True)
sys.exit(rc if rc>=0 else 128-rc)
