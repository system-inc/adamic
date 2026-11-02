#!/usr/bin/env python3
"""Run one bounded census and record Linux child resource usage."""
import json,resource,subprocess,sys,time
from pathlib import Path
output=Path(sys.argv[1]);command=sys.argv[2:]
started=time.monotonic()
status=subprocess.call(command)
usage=resource.getrusage(resource.RUSAGE_CHILDREN)
output.write_text(json.dumps({'command':command,'wall_seconds':time.monotonic()-started,'peak_rss_kib':usage.ru_maxrss,'user_seconds':usage.ru_utime,'system_seconds':usage.ru_stime,'exit_status':status},indent=2)+'\n')
sys.exit(status if status>=0 else 128-status)
