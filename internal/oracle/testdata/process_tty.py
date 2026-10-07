"""Observe each output independently on a terminal or a regular file."""
import base64
import errno
import json
import os
import pty
import subprocess
import sys
import tempfile

streams = []
masters = []
for terminal in sys.argv[1:3]:
    if terminal == "true":
        master, slave = pty.openpty()
        masters.append(master)
        streams.append(os.fdopen(slave, "wb"))
    else:
        masters.append(None)
        streams.append(tempfile.TemporaryFile())
environment = os.environ.copy()
for key in ["NO_COLOR", "FORCE_COLOR", "ADAMIC_PROCESS_TEST"]:
    environment.pop(key, None)
environment.update(json.loads(sys.argv[3]))
environment["ASAN_OPTIONS"] = "detect_leaks=0"
child = subprocess.Popen(sys.argv[4:], stdout=streams[0], stderr=streams[1], env=environment)
child.wait(timeout=15)
outputs = []
for stream, master in zip(streams, masters):
    if master is None:
        stream.seek(0)
        output = stream.read()
    else:
        stream.close()
        output = b""
        while True:
            try:
                part = os.read(master, 4096)
            except OSError as error:
                if error.errno != errno.EIO:
                    raise
                break
            if not part:
                break
            output += part
        os.close(master)
    stream.close()
    outputs.append(base64.b64encode(output).decode("ascii"))
print(json.dumps({"stdout": outputs[0], "stderr": outputs[1], "exitCode": child.returncode}))
