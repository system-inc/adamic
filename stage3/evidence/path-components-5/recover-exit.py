import json,os,signal,sys,time
from pathlib import Path
pid=110579;start='4465972';out=Path(__file__).parent; observed=time.monotonic(); stopped=None
while True:
 try:s=Path(f'/proc/{pid}/stat').read_text().rsplit(')',1)[1].split()
 except FileNotFoundError:
  print('Go process reaped before its own exit status could be recovered',flush=True);sys.exit(125)
 if s[19]!=start:print('PID reused',flush=True);sys.exit(125)
 age=float(Path('/proc/uptime').read_text().split()[0])-int(start)/os.sysconf('SC_CLK_TCK')
 if s[0]=='Z':
  raw=int(s[49]);rc=os.waitstatus_to_exitcode(raw);break
 if age>=1780:
  stopped='1780 second process limit after restart';os.killpg(pid,signal.SIGTERM)
  if age>=1785:os.killpg(pid,signal.SIGKILL)
 time.sleep(1)
result=dict(command=['go','test','-json','-count=1','-parallel=3','-timeout=29m','./stage3/lane'],exit_code=rc,raw_wait_status=raw,stopping_reason=stopped,seconds=age,recovery='Original wrapper was lost during workspace restart; actual Go exit status recovered from zombie /proc/PID/stat field 52.',pid=pid,starttime_ticks=start)
(out/'lane-whole-final-status.json').write_text(json.dumps(result,indent=2)+'\n');print('lane-whole-final recovered_go_exit='+str(rc),flush=True);sys.exit(rc if rc>=0 else 128-rc)
