"""Linux output edges. Node runs first; every assertion is against an observation of Node.
Python supplies POSIX child setup and PTYs that Go's portable os/exec does not expose.
All children have a deadline and are reaped, including on assertion failures.
"""
import json
import os
import pty
import resource
import select
import signal
import subprocess
import sys
import tempfile
import time

mode, commands = sys.argv[1], json.loads(sys.argv[2])
environment = dict(os.environ, ASAN_OPTIONS='detect_leaks=0')


def checked(node, other, label):
    assert (other.returncode, other.stdout, other.stderr) == (node.returncode, node.stdout, node.stderr), (
        label, 'Node', node.returncode, len(node.stdout), node.stderr[:200],
        'other', other.returncode, len(other.stdout), other.stderr[:200],
        'stdout prefixes', node.stdout[:100], other.stdout[:100])


def stopped(command, sig, ignored=False):
    def setup():
        resource.setrlimit(resource.RLIMIT_CORE, (0, 0))
        if ignored:
            signal.signal(sig, signal.SIG_IGN)
    child = subprocess.Popen(command, env=environment, stdout=subprocess.PIPE,
                             stderr=subprocess.PIPE, preexec_fn=setup)
    try:
        prefix = b''
        if os.path.basename(command[0]) == 'node':
            # Observe Node's startup line rather than assuming module loading completes in
            # 750 ms while the full gate is compiling other fixtures. Keep consumed bytes.
            deadline = time.monotonic() + 15
            while not prefix.endswith(b'\n'):
                remaining = deadline - time.monotonic()
                assert remaining > 0 and select.select([child.stdout], [], [], remaining)[0], 'Node startup timed out'
                part = os.read(child.stdout.fileno(), 4096)
                assert part, ('Node ended before startup line', child.poll())
                prefix += part
        else:
            # Native intentionally buffers this line until stopped; keep its existing delay.
            time.sleep(0.75)
        assert child.poll() is None, ('child ended before signal', child.returncode)
        child.send_signal(sig)
        try:
            output, errors = child.communicate(timeout=10)
        except subprocess.TimeoutExpired:
            child.kill()
            output, errors = child.communicate()
            return subprocess.CompletedProcess(command, 124, prefix + output, errors)
        return subprocess.CompletedProcess(command, child.returncode, prefix + output, errors)
    finally:
        if child.poll() is None:
            child.kill()
        child.wait()


if mode == 'ignored':
    for sig in [signal.SIGINT, signal.SIGHUP, signal.SIGTERM]:
        reports = [stopped(command, sig, ignored=True) for command in commands]
        node = reports[0]
        print(mode, sig.name, [(r.returncode, len(r.stdout), r.stderr[:80]) for r in reports], flush=True)
        assert node.returncode == -sig and node.stdout == b'started, about to work for a long time\n' and node.stderr == b''
        for report in reports[1:]:
            checked(node, report, (mode, sig.name))
elif mode == 'terminal':
    for command in commands:
        with tempfile.TemporaryFile() as errors:
            reader, writer = pty.openpty()
            output = writer
            child = subprocess.Popen(command, env=environment, stdout=output, stderr=errors)
            os.close(writer)
            arrivals = []
            held = b''
            started = time.monotonic()
            try:
                while time.monotonic() - started < 5:
                    if select.select([reader], [], [], 0.01)[0]:
                        try:
                            held += os.read(reader, 4096).replace(b'\r\n', b'\n')
                        except OSError:
                            break
                    while len(arrivals) < held.count(b'\n'):
                        arrivals.append(time.monotonic() - started)
                    if child.poll() is not None:
                        break
                code = child.wait(timeout=5)
                errors.seek(0)
                stderr = errors.read()
                print(mode, command[0], code, 'arrivals', [round(t, 3) for t in arrivals], stderr, flush=True)
                assert code == 0 and stderr == b'' and held == b'line\n' * 10
                assert len(arrivals) == 10 and arrivals[0] < 1 and arrivals[-1] - arrivals[0] > 1, 'output was held until exit'
                # A line held with the next one shows as a gap near zero. A starved reader can take two lines in one read
                # now and then (Loom's boxes run at loads above 100), so two such gaps are tolerated; a writer that holds
                # even every second line fails about half of them, and one that holds until exit failed above.
                separate = sum(1 for a, b in zip(arrivals, arrivals[1:]) if 0.08 < b - a < 0.5)
                assert separate >= 7, 'each line must arrive separately as written: %d of 9 gaps' % separate
            finally:
                os.close(reader)
                if child.poll() is None:
                    child.kill()
                child.wait()
else:
    raise AssertionError(mode)
