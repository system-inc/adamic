"""Run one command and record wall time / wait4 peak RSS without GNU time."""
import json, os, sys, time
from pathlib import Path
output, *command = sys.argv[1:]
start=time.monotonic()
pid=os.fork()
if pid==0:
 os.execvp(command[0],command)
_,status,usage=os.wait4(pid,0)
code=os.waitstatus_to_exitcode(status)
Path(output).write_text(json.dumps({'command':command,'wall_seconds':time.monotonic()-start,'peak_rss_kb':usage.ru_maxrss,'user_seconds':usage.ru_utime,'system_seconds':usage.ru_stime,'exit_code':code},indent=2)+'\n')
sys.exit(code if code>=0 else 128-code)
