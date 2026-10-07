"""Observe Linux file/pipe output, including an intentionally backpressured pipe."""
import base64
import fcntl
import json
import os
import subprocess
import sys
import tempfile
import threading

destination, shared, backpressure = sys.argv[1:4]
with tempfile.TemporaryFile() as output, tempfile.TemporaryFile() as errors:
    reader = writer = None
    capacity = 0
    if destination in ("pipe", "stderr-pipe"):
        reader, writer = os.pipe()
        if backpressure == "true":
            fcntl.fcntl(writer, fcntl.F_SETPIPE_SZ, 4096)
        capacity = fcntl.fcntl(writer, fcntl.F_GETPIPE_SZ)
    environment = os.environ.copy()
    environment["ASAN_OPTIONS"] = "detect_leaks=1"
    environment.pop("FORCE_COLOR", None)
    child = subprocess.Popen(sys.argv[4:], stdout=writer if destination == "pipe" else output,
                             stderr=subprocess.STDOUT if shared == "true" else
                             writer if destination == "stderr-pipe" else errors,
                             env=environment)
    pieces = []
    if writer is not None:
        os.close(writer)
        if backpressure == "true":
            # Node can exit with queued bytes dropped. A synchronous native writer blocks;
            # after one second the same reader starts draining it, so neither run deadlocks.
            try:
                child.wait(timeout=1)
            except subprocess.TimeoutExpired:
                pass

        def drain():
            while True:
                part = os.read(reader, 65536)
                if not part:
                    break
                pieces.append(part)
            os.close(reader)

        thread = threading.Thread(target=drain)
        thread.start()
    try:
        child.wait(timeout=15)
    finally:
        if child.poll() is None:
            child.kill()
            child.wait()
        if reader is not None:
            thread.join(timeout=15)
    output.seek(0)
    errors.seek(0)
    landed = b"".join(pieces) if destination == "pipe" else output.read()
    landed_errors = b"".join(pieces) if destination == "stderr-pipe" else errors.read()
    print(json.dumps({"stdout": base64.b64encode(landed).decode("ascii"),
                      "stderr": base64.b64encode(landed_errors).decode("ascii"),
                      "exitCode": child.returncode, "pipeCapacity": capacity}))
