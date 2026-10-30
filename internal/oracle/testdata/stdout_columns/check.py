#!/usr/bin/env python3
"""Bounded pty oracle protocol; no polling of a program's output for correctness.
The ioctl log observes the loop checkpoint before the host completion is released.
"""
import argparse
import fcntl
import json
import os
from pathlib import Path
import pty
import select
import signal
import struct
import subprocess
import tempfile
import termios
import time


def observe(args, directory, native):
    label = 'native' if native else 'node'
    count = directory / (label + '.ioctls')
    allow = directory / (label + '.allow')
    initial_failure = args.fixture.startswith('InitialFailure')
    failure = args.fixture in ('RefreshFailure', 'InitialFailureRefreshFailure')
    synchronous = args.fixture == 'SynchronousResize'
    no_signal = args.fixture == 'InitialFailureNoSignal'
    master, slave = pty.openpty()
    fcntl.ioctl(master, termios.TIOCSWINSZ, struct.pack('HHHH', 24, 80, 0, 0))
    mode = 'sync' if synchronous else 'async'
    command = [args.native, mode] if native else [
        args.node, '--disable-warning=ExperimentalWarning',
        str(Path(args.root) / 'oracle/node.mjs'),
        str(Path(__file__).with_name('loop.mjs')), mode]
    environment = {
        'PATH': '/usr/bin:/bin', 'TERM': 'dumb',
        'LD_PRELOAD': args.preload, 'COLUMNS_IOCTL_LOG': str(count),
        'ASAN_OPTIONS': 'detect_leaks=1:halt_on_error=1',
        'UBSAN_OPTIONS': 'halt_on_error=1',
    }
    if initial_failure:
        environment['COLUMNS_FAIL_UNTIL'] = str(allow)
    child = subprocess.Popen(command, stdin=subprocess.PIPE, stdout=slave,
                             stderr=subprocess.PIPE, env=environment,
                             start_new_session=True)
    os.close(slave)
    try:
        if not select.select([child.stderr], [], [], 5)[0]:
            raise RuntimeError(label + ': no initial width')
        first = child.stderr.readline()
        expected = b'initial undefined\n' if initial_failure else b'initial 80\n'
        if first != expected:
            raise RuntimeError(label + ': invalid control startup: ' + repr(first))
        # Recovery is a real transient OS failure: never patch a JS property or
        # Node method. The injector also covers Node's duplicated TTY fd.
        if not failure:
            allow.touch()
        if not no_signal:
            if failure:
                os.close(master)
                master = None
            else:
                fcntl.ioctl(master, termios.TIOCSWINSZ,
                            struct.pack('HHHH', 24, 120, 0, 0))
            os.kill(child.pid, signal.SIGWINCH)
        if failure:
            _, rest = child.communicate(timeout=5)
        else:
            if not synchronous and not no_signal:
                # Wait only for the checkpoint, not for a fixed quiet delay.
                # A never-refresh mutant cannot reach it; release its host
                # request after the bound so the value comparison catches it.
                deadline = time.monotonic() + 0.5
                while time.monotonic() < deadline:
                    if count.exists() and count.stat().st_size >= 2:
                        break
                    time.sleep(0.001)
            child.stdin.write(b'x')
            child.stdin.flush()
            _, rest = child.communicate(timeout=5)
        stderr = first + rest
        if b'Sanitizer' in stderr or b'runtime error:' in stderr:
            raise RuntimeError(label + ': sanitizer failure: ' + repr(stderr))
        return {'exit': child.returncode, 'stderr': stderr.decode(),
                'ioctls': count.stat().st_size if count.exists() else 0}
    finally:
        if child.poll() is None:
            os.killpg(child.pid, signal.SIGKILL)
            child.wait()
        if master is not None:
            os.close(master)


def main():
    parser = argparse.ArgumentParser()
    for name in ('node', 'native', 'preload', 'fixture', 'root'):
        parser.add_argument('--' + name, required=True)
    args = parser.parse_args()
    with tempfile.TemporaryDirectory() as tmp:
        directory = Path(tmp)
        node = observe(args, directory, False)
        native = observe(args, directory, True)
    # Pin the Node controls separately from the comparison, so a broken setup
    # cannot accidentally turn a mutant green.
    initial = 'undefined' if args.fixture.startswith('InitialFailure') else '80'
    failure = args.fixture in ('RefreshFailure', 'InitialFailureRefreshFailure')
    later = '80' if args.fixture == 'SynchronousResize' else (
        'undefined' if args.fixture == 'InitialFailureNoSignal' else '120')
    expected = 'initial ' + initial + '\n' + ('' if failure else 'later ' + later + '\n')
    if node['exit'] != (70 if failure else 0) or node['stderr'] != expected:
        raise RuntimeError('Node control changed: ' + repr(node))
    print(json.dumps({'Node': {'Stderr': node['stderr'], 'ExitCode': node['exit']},
                      'Native': {'Stderr': native['stderr'], 'ExitCode': native['exit']},
                      'NodeCount': node['ioctls'], 'NativeCount': native['ioctls']}))


if __name__ == '__main__':
    main()
