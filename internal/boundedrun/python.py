"""Bound Python shell-outs, including the catalog and setup dependency installers."""
import os
import math
from pathlib import Path
import runpy
import subprocess
import sys
import signal

class DeadlineExceeded(subprocess.TimeoutExpired):
    def __str__(self):
        return f"child {self.cmd[0]}: deadline exceeded after {self.timeout}s; process group killed"


def run(command, **kwargs):
    # Observed cold setup 114s and cold directory 49s. Ten minutes allows
    # toolchain downloads/builds; catalog go tests retain a 70-minute ceiling.
    deadline = kwargs.pop("deadline", 4200 if command[:2] == ["go", "test"] else 600)
    cap = float(os.environ.get("ADAMIC_CHILD_DEADLINE", deadline))
    if not math.isfinite(cap) or cap <= 0:
        raise ValueError("ADAMIC_CHILD_DEADLINE must be positive seconds")
    deadline = min(deadline, cap)
    requested = kwargs.pop("timeout", None)
    if requested is not None:
        deadline = min(deadline, requested)
    check = kwargs.pop("check", False)
    data = kwargs.pop("input", None)
    if data is not None:
        if kwargs.get("stdin") is not None:
            raise ValueError("stdin and input cannot both be supplied")
        kwargs["stdin"] = subprocess.PIPE
    if kwargs.pop("capture_output", False):
        if kwargs.get("stdout") is not None or kwargs.get("stderr") is not None:
            raise ValueError("capture_output and output streams cannot both be supplied")
        kwargs.update(stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    kwargs["start_new_session"] = True
    process = subprocess.Popen(command, **kwargs)
    def kill_group():
        try:
            os.killpg(process.pid, signal.SIGKILL)
        except ProcessLookupError:
            pass
    try:
        stdout, stderr = process.communicate(data, timeout=deadline)
    except subprocess.TimeoutExpired:
        kill_group()
        try:
            stdout, stderr = process.communicate(timeout=1)
        except subprocess.TimeoutExpired:
            for stream in (process.stdin, process.stdout, process.stderr):
                if stream is not None:
                    stream.close()
            stdout = stderr = None
        raise DeadlineExceeded(command, deadline, output=stdout, stderr=stderr) from None
    finally:
        kill_group()
    result = subprocess.CompletedProcess(command, process.returncode, stdout, stderr)
    if check:
        result.check_returncode()
    return result


if __name__ == "__main__":
    # Only this interpreter's children are adapted; the repository scripts need
    # no edits outside the unit's territory.
    subprocess.run = run
    script = sys.argv[1]
    sys.argv = sys.argv[1:]
    sys.path.insert(0, str(Path(script).resolve().parent))
    runpy.run_path(script, run_name="__main__")
