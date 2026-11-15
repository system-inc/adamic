"""Linux wall time and peak child RSS via wait4; no external time package."""
import json, os, pathlib, sys, time
stem=pathlib.Path(sys.argv[1]); stem.parent.mkdir(parents=True,exist_ok=True)
start=time.monotonic()
pid=os.fork()
if pid==0:
    with open(str(stem)+'.json','wb') as out, open(str(stem)+'.stderr','wb') as err:
        os.dup2(out.fileno(),1); os.dup2(err.fileno(),2)
        os.execvpe(sys.argv[2],sys.argv[2:],os.environ)
_,status,usage=os.wait4(pid,0)
code=os.waitstatus_to_exitcode(status)
pathlib.Path(str(stem)+'.time.json').write_text(json.dumps(dict(wall_seconds=time.monotonic()-start,peak_rss_kib=usage.ru_maxrss,exit_code=code),indent=2)+'\n')
sys.exit(code if code>=0 else 128-code)
