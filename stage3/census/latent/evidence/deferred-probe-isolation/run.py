import json,os,pathlib,resource,signal,subprocess,sys,time
out=pathlib.Path(sys.argv[1]);out.mkdir(parents=True,exist_ok=True)
argv=sys.argv[2:];start=time.monotonic()
deadline=float(os.environ.get("LDPI_DEADLINE_SECONDS","1700"))
if not 0 < deadline <= 1700: raise ValueError("deadline must be in (0,1700]")
with (out/'output.log').open('w') as log:
 child=subprocess.Popen(argv,stdout=log,stderr=log,start_new_session=True)
 expired=False
 try: code=child.wait(timeout=deadline)
 except subprocess.TimeoutExpired:
  expired=True;os.killpg(child.pid,signal.SIGTERM)
  try: code=child.wait(timeout=5)
  except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);code=child.wait()
row=dict(argv=argv,child_exit_code=code,deadline_expired=expired,wall_seconds=time.monotonic()-start,peak_rss_kib=resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss)
(out/'status.json').write_text(json.dumps(row,indent=2)+'\n');print(json.dumps(row),flush=True)
if code<0:
 if -code not in (signal.SIGKILL,signal.SIGSTOP):signal.signal(-code,signal.SIG_DFL)
 os.kill(os.getpid(),-code)
sys.exit(code)
