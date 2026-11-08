"""Detach the resource runner from the transport; only its status file means completion."""
import json
from pathlib import Path
import subprocess
import sys

prefix = Path(sys.argv[1]).resolve()
arguments = sys.argv[2:]
prefix.parent.mkdir(parents=True,exist_ok=True)
with prefix.with_suffix('.supervisor.log').open('w') as log:
    process = subprocess.Popen([sys.executable,str(Path(__file__).with_name('run.py')),*arguments],
                               stdin=subprocess.DEVNULL,stdout=log,stderr=subprocess.STDOUT,
                               start_new_session=True)
value = dict(runner_pid=process.pid,state='launched; child exit is pending in the resource runner status file',arguments=arguments)
prefix.with_suffix('.launch.json').write_text(json.dumps(value,indent=2)+'\n')
print(json.dumps(value),flush=True)
