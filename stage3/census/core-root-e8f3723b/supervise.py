import json, os, signal, subprocess, sys, time
from pathlib import Path
base = Path('/tmp/core-root-avqasqt')
name, *command = sys.argv[1:]
started = time.monotonic()
with (base / (name + '.log')).open('wb') as log:
    p = subprocess.Popen(command, stdout=log, stderr=log, start_new_session=True)
    try:
        try:
            rc = p.wait(timeout=7200)
        except subprocess.TimeoutExpired:
            os.killpg(p.pid, signal.SIGTERM)
            try:
                rc = p.wait(timeout=10)
            except subprocess.TimeoutExpired:
                os.killpg(p.pid, signal.SIGKILL)
                rc = p.wait()
    finally:
        try:
            os.killpg(p.pid, signal.SIGKILL)
        except ProcessLookupError:
            pass
status = dict(command=command, exit_code=rc, seconds=time.monotonic()-started, process_group_cleaned=True)
(base / (name + '-status.json')).write_text(json.dumps(status, indent=2)+'\n')
print(name + ' child_exit=' + str(rc), flush=True)
sys.exit(rc if rc >= 0 else 128-rc)
