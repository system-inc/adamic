import json, os, pathlib, signal, subprocess, sys, time

out = pathlib.Path(__file__).resolve().parent
label, *command = sys.argv[1:]
started = time.monotonic()
with (out / (label + '.log')).open('wb') as log:
    child = subprocess.Popen(command, cwd='/workspace/adamic', stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
    timed_out = False
    try:
        child.wait(timeout=2400)
    except subprocess.TimeoutExpired:
        timed_out = True
        os.killpg(child.pid, signal.SIGTERM)
        try:
            child.wait(timeout=10)
        except subprocess.TimeoutExpired:
            os.killpg(child.pid, signal.SIGKILL)
            child.wait()
record = dict(command=command, child_exit=child.returncode, timed_out=timed_out, seconds=time.monotonic()-started)
(out / (label + '.status.json')).write_text(json.dumps(record, indent=2) + '\n')
print(json.dumps(record), flush=True)
sys.exit(child.returncode if child.returncode >= 0 else 128-child.returncode)
