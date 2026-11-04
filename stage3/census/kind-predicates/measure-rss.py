import os,sys,time,json
from pathlib import Path
output=Path(sys.argv[1]); started=time.monotonic()
pid=os.fork()
if pid==0:
 os.execv(sys.argv[2],sys.argv[2:])
_,status,usage=os.wait4(pid,0)
code=os.waitstatus_to_exitcode(status)
output.write_text(json.dumps(dict(peak_rss_kib=usage.ru_maxrss,wall_seconds=time.monotonic()-started,user_seconds=usage.ru_utime,system_seconds=usage.ru_stime,exit_code=code,signal=-code if code<0 else None),indent=2)+'\n')
sys.exit(128-code if code<0 else code)
