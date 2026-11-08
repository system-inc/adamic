"""Measure the Node child's complete lifetime with Linux wait4, with a hard deadline."""
import json
import os
from pathlib import Path
import signal
import sys
import time

output, deadline_text, command, *arguments = sys.argv[1:]
deadline = float(deadline_text)
started = time.time()
child = os.fork()
if child == 0:
    os.execv(command, [command, *arguments])


def stop(signum, frame):
    try:
        os.kill(child, signal.SIGKILL)
    except ProcessLookupError:
        pass


signal.signal(signal.SIGTERM, stop)
signal.signal(signal.SIGINT, stop)
timed_out = False
while True:
    pid, status, usage = os.wait4(child, os.WNOHANG)
    if pid:
        break
    if time.time() >= deadline:
        timed_out = True
        stop(None, None)
    time.sleep(0.05)

code = os.waitstatus_to_exitcode(status)
result = dict(exitCode=code, signal=-code if code < 0 else None,
              timedOut=timed_out, peakRssKiB=usage.ru_maxrss,
              seconds=time.time() - started,
              node=os.environ.get("ARRAY_COPIES_NODE_VERSION"),
              command=command, arguments=arguments)
Path(output).write_text(json.dumps(result, indent=2) + "\n")
print(f"Node child exit code: {code}", flush=True)
sys.exit(code if code >= 0 else 128 - code)
