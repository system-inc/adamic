"""Linux output edges. Node runs first; every assertion is against an observation of Node.
Python supplies POSIX child setup and PTYs that Go's portable os/exec does not expose.
All children have a deadline and are reaped, including on assertion failures.
"""
import fcntl
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


def run(command, **kwargs):
    return subprocess.run(command, env=environment, capture_output=True, timeout=15, **kwargs)


def checked(node, other, label):
    assert (other.returncode, other.stdout, other.stderr) == (node.returncode, node.stdout, node.stderr), (
        label, 'Node', node.returncode, len(node.stdout), node.stderr[:200],
        'other', other.returncode, len(other.stdout), other.stderr[:200],
        'stdout prefixes', node.stdout[:100], other.stdout[:100])


def limit():
    resource.setrlimit(resource.RLIMIT_FSIZE, (102400, 102400))


def stopped(command, sig, ignored=False, cpu=False):
    def setup():
        resource.setrlimit(resource.RLIMIT_CORE, (0, 0))
        if ignored:
            signal.signal(sig, signal.SIG_IGN)
        if cpu:
            resource.setrlimit(resource.RLIMIT_CPU, (1, 5))
    child = subprocess.Popen(command, env=environment, stdout=subprocess.PIPE,
                             stderr=subprocess.PIPE, preexec_fn=setup)
    try:
        if not cpu:
            time.sleep(0.75)
            assert child.poll() is None, ('child ended before signal', child.returncode)
            child.send_signal(sig)
        try:
            output, errors = child.communicate(timeout=3)
        except subprocess.TimeoutExpired:
            child.kill()
            output, errors = child.communicate()
            return subprocess.CompletedProcess(command, 124, output, errors)
        return subprocess.CompletedProcess(command, child.returncode, output, errors)
    finally:
        if child.poll() is None:
            child.kill()
        child.wait()


if mode == 'fsize':
    reports = []
    for command in commands:
        with tempfile.TemporaryDirectory() as directory:
            report = run(command + [directory], preexec_fn=limit)
            size = os.stat(directory + '/big.txt').st_size
            print('fsize', command[0], report.returncode, report.stdout, report.stderr, size, flush=True)
            assert size == 102400, size
            reports.append(report)
    node = reports[0]
    assert (node.returncode, node.stdout, node.stderr) == (0, b'cannot write <dir>/big.txt: failed\n', b'after\n')
    for report in reports[1:]:
        checked(node, report, mode)
elif mode == 'fsize_out':
    reports = []
    for command in commands:
        with tempfile.TemporaryFile() as output:
            child = subprocess.run(command, env=environment, stdout=output, stderr=subprocess.PIPE,
                                   preexec_fn=limit, timeout=15)
            output.seek(0)
            report = subprocess.CompletedProcess(command, child.returncode, output.read(), child.stderr)
            print(mode, command[0], report.returncode, len(report.stdout), report.stderr, flush=True)
            reports.append(report)
    assert reports[0].returncode == 70 and len(reports[0].stdout) == 102400
    for report in reports[1:]:
        checked(reports[0], report, mode)
elif mode == 'nonblock':
    reports = []
    for command in commands:
        reader, writer = os.pipe()
        fcntl.fcntl(writer, fcntl.F_SETFL, fcntl.fcntl(writer, fcntl.F_GETFL) | os.O_NONBLOCK)
        with tempfile.TemporaryFile() as errors:
            child = subprocess.Popen(command, env=environment, stdout=writer, stderr=errors)
            os.close(writer)
            chunks = []
            try:
                time.sleep(0.5)
                deadline = time.monotonic() + 15
                while True:
                    assert time.monotonic() < deadline, 'nonblocking pipe timed out'
                    assert select.select([reader], [], [], 2)[0], 'pipe made no progress'
                    chunk = os.read(reader, 4096)
                    if not chunk:
                        break
                    chunks.append(chunk)
                    time.sleep(0.0005)
                child.wait(timeout=5)
            finally:
                os.close(reader)
                if child.poll() is None:
                    child.kill()
                child.wait()
            errors.seek(0)
            report = subprocess.CompletedProcess(command, child.returncode, b''.join(chunks), errors.read())
            print(mode, command[0], report.returncode, len(report.stdout), report.stderr, flush=True)
            reports.append(report)
    assert reports[0].returncode == 0 and len(reports[0].stdout) == 1660112
    for report in reports[1:]:
        checked(reports[0], report, mode)
elif mode == 'closed':
    for descriptors in [(0,), (1,), (2,), (0, 1, 2)]:
        def close():
            for descriptor in descriptors:
                os.close(descriptor)
        reports = [run(command, preexec_fn=close) for command in commands]
        node = reports[0]
        print(mode, descriptors, [(r.returncode, r.stdout, r.stderr) for r in reports], flush=True)
        assert node.returncode == 0
        assert node.stdout == (b'' if 1 in descriptors else b'to stdout\n')
        assert node.stderr == (b'' if 2 in descriptors else b'to stderr\n')
        for report in reports[1:]:
            checked(node, report, (mode, descriptors))
elif mode in ('signals', 'ignored', 'realtime'):
    signals = [signal.SIGINT, signal.SIGHUP, signal.SIGTERM] if mode == 'ignored' else [
        signal.SIGTERM, signal.SIGINT, signal.SIGHUP, signal.SIGUSR2,
        signal.SIGALRM, signal.SIGXCPU, signal.SIGQUIT,
        signal.SIGVTALRM, signal.SIGPROF, signal.SIGIO, signal.SIGPWR]
    if mode == 'realtime':
        signals = [int(sys.argv[3])]
    for sig in signals:
        name = sig.name if isinstance(sig, signal.Signals) else f'SIGRT({sig})'
        reports = [stopped(command, sig, ignored=(mode == 'ignored')) for command in commands]
        node = reports[0]
        print(mode, name, [(r.returncode, len(r.stdout), r.stderr[:80]) for r in reports], flush=True)
        assert node.returncode == -sig and node.stdout == b'started, about to work for a long time\n' and node.stderr == b''
        for report in reports[1:]:
            checked(node, report, (mode, name))
    if mode == 'signals':
        reports = [stopped(command, signal.SIGXCPU, cpu=True) for command in commands]
        print('cpu limit', [(r.returncode, len(r.stdout)) for r in reports], flush=True)
        assert reports[0].returncode == -signal.SIGXCPU and reports[0].stdout
        for report in reports[1:]:
            checked(reports[0], report, 'cpu limit')
elif mode == 'usr1':
    # Node starts its inspector on SIGUSR1 and goes on, saying so on stderr; natively it's ignored and
    # the program goes on. Exit status and stdout must match; Node's stderr is its inspector's notice.
    reports = []
    for command in commands:
        child = subprocess.Popen(command, env=environment, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        try:
            time.sleep(0.3)
            assert child.poll() is None, ('child ended before signal', child.returncode)
            child.send_signal(signal.SIGUSR1)
            output, errors = child.communicate(timeout=30)
            reports.append(subprocess.CompletedProcess(command, child.returncode, output, errors))
        finally:
            if child.poll() is None:
                child.kill()
            child.wait()
    print(mode, [(r.returncode, r.stdout, r.stderr[:80]) for r in reports], flush=True)
    node = reports[0]
    assert node.returncode == 0 and node.stdout.startswith(b'started') and b'finished' in node.stdout
    for report in reports[1:]:
        assert (report.returncode, report.stdout) == (node.returncode, node.stdout), (mode, report.returncode, report.stdout, report.stderr[:200])
elif mode == 'panic':
    reports = [run(command) for command in commands]
    node = reports[0]
    print(mode, [(r.returncode, r.stdout, r.stderr) for r in reports], flush=True)
    assert (node.returncode, node.stdout, node.stderr) == (70, 'printed � first\n'.encode(), 'adamic: panic: lone � here\n'.encode())
    for report in reports[1:]:
        checked(node, report, mode)
elif mode in ('terminal', 'file'):
    for command in commands:
        with tempfile.TemporaryFile() as errors, tempfile.TemporaryFile() as file:
            if mode == 'terminal':
                reader, writer = pty.openpty()
                output = writer
            else:
                output = file
            child = subprocess.Popen(command, env=environment, stdout=output, stderr=errors)
            if mode == 'terminal':
                os.close(writer)
            arrivals = []
            held = b''
            started = time.monotonic()
            try:
                while time.monotonic() - started < 5:
                    if mode == 'terminal':
                        if select.select([reader], [], [], 0.01)[0]:
                            try:
                                held += os.read(reader, 4096).replace(b'\r\n', b'\n')
                            except OSError:
                                break
                    else:
                        held = os.pread(file.fileno(), 10000, 0)
                        time.sleep(0.01)
                    while len(arrivals) < held.count(b'\n'):
                        arrivals.append(time.monotonic() - started)
                    if child.poll() is not None:
                        if mode == 'file':
                            held = os.pread(file.fileno(), 10000, 0)
                            while len(arrivals) < held.count(b'\n'):
                                arrivals.append(time.monotonic() - started)
                        break
                code = child.wait(timeout=5)
                errors.seek(0)
                stderr = errors.read()
                print(mode, command[0], code, 'arrivals', [round(t, 3) for t in arrivals], stderr, flush=True)
                assert code == 0 and stderr == b'' and held == b'line\n' * 10
                assert len(arrivals) == 10 and arrivals[0] < 1 and arrivals[-1] - arrivals[0] > 1, 'output was held until exit'
                assert all(0.08 < b - a < 0.5 for a, b in zip(arrivals, arrivals[1:])), 'each line must arrive separately as written'
            finally:
                if mode == 'terminal':
                    os.close(reader)
                if child.poll() is None:
                    child.kill()
                child.wait()
else:
    raise AssertionError(mode)
