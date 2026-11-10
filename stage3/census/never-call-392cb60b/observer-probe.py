import ctypes,os,subprocess,sys,json,resource
from pathlib import Path
out=Path(__file__).resolve().parent;libc=ctypes.CDLL(None,use_errno=True);libc.ptrace.restype=ctypes.c_long
with (out/'observer-probe-child.log').open('wb') as log:
 p=subprocess.Popen([sys.executable,'-c','import time,sys; value=bytearray(64*1024*1024); time.sleep(0.2); sys.exit(7)'],stdout=log,stderr=subprocess.STDOUT)
 def trace(request,data=0):
  arg=ctypes.cast(data,ctypes.c_void_p) if not isinstance(data,int) else ctypes.c_void_p(data)
  if libc.ptrace(ctypes.c_ulong(request),ctypes.c_ulong(p.pid),ctypes.c_void_p(),arg)==-1:raise OSError(ctypes.get_errno(),os.strerror(ctypes.get_errno()))
 try:
  trace(0x4206,0x40);peak=None;event_status=None
  while True:
   _,status=os.waitpid(p.pid,0x40000000)
   if os.WIFEXITED(status) or os.WIFSIGNALED(status):break
   event=status>>16
   if event==6:
    value=ctypes.c_ulong();trace(0x4201,ctypes.byref(value));event_status=value.value
    for line in Path('/proc/'+str(p.pid)+'/status').read_text().splitlines():
     if line.startswith('VmHWM:'):peak=int(line.split()[1])
    trace(7,0)
   else:trace(7,0 if event else os.WSTOPSIG(status))
  rc=os.waitstatus_to_exitcode(status);p.returncode=rc
  assert rc==7 and event_status==status==7<<8
  assert peak is not None and peak>=64*1024
  row=dict(expected_child_exit_code=7,actual_child_exit_code=rc,raw_exit_status=status,exit_event_status=event_status,peak_rss_kib_at_exit=peak,rusage_children_peak_rss_kib=resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss,passed=True)
  (out/'observer-probe.json').write_text(json.dumps(row,indent=2)+'\n');print(json.dumps(row))
 finally:
  if p.returncode is None:
   try:p.kill()
   except ProcessLookupError:pass
   try:p.wait()
   except ChildProcessError:pass
