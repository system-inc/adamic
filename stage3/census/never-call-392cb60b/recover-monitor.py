import os,sys,time,json,ctypes,signal,pathlib,shutil,subprocess,concurrent.futures,datetime
out=pathlib.Path(__file__).resolve().parent;live=pathlib.Path('/workspace/never-call-live')
libc=ctypes.CDLL(None,use_errno=True);libc.ptrace.restype=ctypes.c_long
SEIZE=0x4206;CONT=7;GETEVENTMSG=0x4201;TRACEEXIT=0x40;EVENT_EXIT=6;WALL=0x40000000
pids={'parent':14034,'head':14121};states={};results={}
def trace(request,pid,data=0):
 arg=ctypes.cast(data,ctypes.c_void_p) if not isinstance(data,int) else ctypes.c_void_p(data)
 value=libc.ptrace(ctypes.c_ulong(request),ctypes.c_ulong(pid),ctypes.c_void_p(),arg)
 if value==-1:raise OSError(ctypes.get_errno(),os.strerror(ctypes.get_errno()))
 return value
def hwm(pid):
 for line in pathlib.Path('/proc/'+str(pid)+'/status').read_text().splitlines():
  if line.startswith('VmHWM:'):return int(line.split()[1])
 return None
def boot():return time.clock_gettime(time.CLOCK_BOOTTIME)
def watch(side,pid):
 name='tsc-entry-ordered-first-'+side
 expected=pathlib.Path('/workspace/never-call-tools')/side/'census-entry-ordered'
 assert pathlib.Path('/proc/'+str(pid)+'/exe').resolve()==expected
 stat=pathlib.Path('/proc/'+str(pid)+'/stat').read_text().rsplit(')',1)[1].split()
 start=float(stat[19])/os.sysconf('SC_CLK_TCK');started_epoch=time.time()-(boot()-start)
 peak=hwm(pid) or 0;attached=False;exit_peak=False;event_status=None;ended=None
 try:
  trace(SEIZE,pid,TRACEEXIT);attached=True
  states[side]=dict(pid=pid,attached=True,start_boot_seconds=start,peak_rss_kib=peak)
  (live/(name+'.monitor-state.json')).write_text(json.dumps(states[side],indent=2)+'\n')
  print(side,'attached exit observer to existing compiler',pid,flush=True)
  while True:
   _,status=os.waitpid(pid,WALL)
   if os.WIFEXITED(status) or os.WIFSIGNALED(status):
    attached=False;rc=os.waitstatus_to_exitcode(status);break
   if os.WIFSTOPPED(status):
    event=status>>16
    if event==EVENT_EXIT:
     message=ctypes.c_ulong();trace(GETEVENTMSG,pid,ctypes.byref(message));event_status=message.value
     sample=hwm(pid)
     if sample is not None:peak=max(peak,sample);exit_peak=True
     ended=boot();trace(CONT,pid,0)
    else:
     trace(CONT,pid,0 if event else os.WSTOPSIG(status))
  row=dict(name=name,command=[str(expected),'/workspace/never-call-adapted/src/tsc/tsc.ts',str(live/(name+'.jsonl'))],cwd='/workspace/never-call-report' if side=='parent' else '/workspace/never-call-head',started_utc=datetime.datetime.fromtimestamp(started_epoch,datetime.timezone.utc).isoformat(),seconds=(ended or boot())-start,peak_rss_kib=peak,child_exit_code=rc,silent_cutoff=states.get(side,{}).get('silent_cutoff',False),measurement_recovered=True,peak_rss_method='/proc/PID/status VmHWM at PTRACE_EVENT_EXIT' if exit_peak else 'observed VmHWM lower bound',peak_rss_exact_exit_observation=exit_peak,raw_wait_status=status,ptrace_exit_event_wait_status=event_status,pid=pid)
  (live/(name+'.timing.json')).write_text(json.dumps(row,indent=2)+'\n')
  print(json.dumps(row),flush=True)
  return side,row
 except Exception as error:
  print(side,'recovery observer error:',repr(error),flush=True)
  (live/(name+'.monitor-error.json')).write_text(json.dumps(dict(pid=pid,error=repr(error),attached=attached),indent=2)+'\n')
  raise
 finally:
  # No EXITKILL option is set: interruption of this observer cannot kill a compiler.
  pass
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 pending={pool.submit(watch,side,pid):side for side,pid in pids.items()}
 stamps={};last={side:time.monotonic() for side in pids};terminated={}
 while pending:
  for future in list(pending):
   if not future.done():continue
   side,row=future.result();name=row['name'];results[side]=row['child_exit_code'];del pending[future]
   for p in live.glob(name+'.*'):shutil.copyfile(p,out/p.name)
   (out/'ordered-entry-execution.json').write_text(json.dumps(dict(concurrent=True,launcher_interrupted=True,recovered_via='non-stopping ptrace exit observation; compiler signals forwarded unchanged',completed_compiler_exit_codes=results),indent=2)+'\n')
   for cmd in [['python3','/tmp/never-call-archive.py'],['python3','/tmp/never-call-checkpoint.py',name+'-recovered-complete']]:
    r=subprocess.run(cmd);print('post-run child exit code:',r.returncode,flush=True)
    if r.returncode:raise RuntimeError('checkpoint failed '+str(r.returncode))
  for future,side in pending.items():
   name='tsc-entry-ordered-first-'+side;paths=[live/(name+suffix) for suffix in ['.log','.jsonl','.units.jsonl','.prepasses.log']]
   current=[(p.stat().st_size,p.stat().st_mtime_ns) if p.exists() else None for p in paths]
   if current!=stamps.get(side):last[side]=time.monotonic();stamps[side]=current
   if time.monotonic()-last[side]>1800 and side not in terminated:
    states.setdefault(side,{})['silent_cutoff']=True;os.killpg(pids[side],signal.SIGTERM);terminated[side]=time.monotonic()
   if side in terminated and time.monotonic()-terminated[side]>10:
    try:os.killpg(pids[side],signal.SIGKILL)
    except ProcessLookupError:pass
  if pending:time.sleep(0.1)
(out/'recovery-monitor-complete.json').write_text(json.dumps(dict(completed_compiler_exit_codes=results,monitor_exit_code=0 if not any(results.values()) else 1),indent=2)+'\n')
sys.exit(0 if not any(results.values()) else 1)
