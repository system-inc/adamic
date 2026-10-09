"""Observe each output independently on a terminal or a regular file."""
import base64
import errno
import json
import os
import pty
import subprocess
import sys
sys.dont_write_bytecode = True
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../../../oracle')))
from child_progress import progress, watch_files, communicate
import tempfile
import threading

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

watching = watch_files(child, streams)


def drain(master, parts):
    while True:
        try:
            part = os.read(master, 4096)
        except OSError as error:
            if error.errno != errno.EIO:
                raise
            break
        if not part:
            break
        parts.append(part)
        progress(part)


# Each terminal is read while the child runs: macOS drops what a terminal holds once its
# last writer closes, where Linux keeps it readable until EIO.
readers = []
for stream, master in zip(streams, masters):
    parts = []
    reading = None
    if master is not None:
        stream.close()
        reading = threading.Thread(target=drain, args=(master, parts))
        reading.start()
    readers.append((reading, parts))
child.wait()
outputs = []
for stream, master, (reading, parts) in zip(streams, masters, readers):
    if master is None:
        stream.seek(0)
        output = stream.read()
    else:
        reading.join()
        if reading.is_alive():
            raise RuntimeError("terminal reader did not finish")
        output = b"".join(parts)
        os.close(master)
    stream.close()
    outputs.append(base64.b64encode(output).decode("ascii"))
print(json.dumps({"stdout": outputs[0], "stderr": outputs[1], "exitCode": child.returncode}))
