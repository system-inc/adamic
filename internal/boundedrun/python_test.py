"""Executable hang proofs for the shared Python and shell entry points."""
import os
from pathlib import Path
import signal
import shutil
import sys
from unittest.mock import patch
import subprocess
import tempfile
import time
import unittest
from python import run, DeadlineExceeded


class Deadlines(unittest.TestCase):
    def temporary(self):
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        return directory.name

    def tree(self, directory, ignore_leader=True):
        heartbeat, pidfile, child = [Path(directory)/name for name in ("heartbeat", "pid", "forever-child")]
        child.write_text(f"#!/bin/sh\necho $$ > '{pidfile}'\ntrap '' TERM\n(sh -c 'trap \"\" TERM; while :; do echo tick >> \"$1\"; sleep .01; done' child '{heartbeat}') &\nwhile [ ! -s '{heartbeat}' ]; do sleep .01; done\necho ready > '{heartbeat}.ready'\nwait\n")
        if not ignore_leader:
            child.write_text(child.read_text().replace("trap '' TERM\n", "", 1))
        child.chmod(0o755)
        def cleanup():
            if pidfile.exists():
                try:
                    os.killpg(int(pidfile.read_text()), signal.SIGKILL)
                except ProcessLookupError:
                    pass
        self.addCleanup(cleanup)
        return child, heartbeat

    def stopped(self, heartbeat):
        first = heartbeat.read_bytes()
        self.assertTrue(first)
        time.sleep(.1)
        self.assertEqual(first, heartbeat.read_bytes(), "grandchild survived group kill")

    def wait_ready(self, heartbeat):
        deadline = time.monotonic()+30
        ready = Path(str(heartbeat)+".ready")
        while not (ready.exists() and ready.stat().st_size and heartbeat.exists() and heartbeat.stat().st_size):
            if time.monotonic()>deadline:
                self.fail("child did not signal readiness within bounded startup")
            time.sleep(.005)

    def shell_ready_env(self, directory, heartbeat):
        # Exercise GNU timeout's real SIGALRM/kill-after path, but send its
        # execution alarm only after readiness. Its 30s startup watchdog remains.
        timeout = shutil.which("timeout")
        self.assertTrue(timeout, "GNU timeout is required for the shell proof")
        shimdir = Path(directory)/"ready-bin"
        shimdir.mkdir()
        shim = shimdir/"timeout"
        shim.write_text("#!"+sys.executable+"\n"+r'''import os, signal, subprocess, sys, time
from pathlib import Path
args=sys.argv[1:]
index=next(i for i,arg in enumerate(args) if not arg.startswith('-'))
seconds=float(args[index].removesuffix('s'))
args[index]='30s'
process=subprocess.Popen([os.environ['TEST_GNU_TIMEOUT'],*args],close_fds=False)
heartbeat=Path(os.environ['TEST_HEARTBEAT'])
Path(str(heartbeat)+'.timeout-pid').write_text(str(process.pid))
startup=time.monotonic()+30
while not Path(str(heartbeat)+'.ready').exists():
    if process.poll() is not None:
        sys.exit(process.returncode)
    if time.monotonic()>startup:
        raise RuntimeError('child did not signal readiness')
    time.sleep(.005)
Path(str(heartbeat)+'.armed').write_text(str(time.monotonic()))
time.sleep(seconds)
os.kill(process.pid,signal.SIGALRM)
code=process.wait()
sys.exit(code if code>=0 else 128-code)
''')
        shim.chmod(0o755)
        def cleanup():
            marker=Path(str(heartbeat)+".timeout-pid")
            if marker.exists():
                try: os.killpg(int(marker.read_text()),signal.SIGKILL)
                except ProcessLookupError: pass
        self.addCleanup(cleanup)
        return dict(os.environ, PATH=str(shimdir)+os.pathsep+os.environ['PATH'],
                    ADAMIC_CHILD_DEADLINE="0.2", TEST_GNU_TIMEOUT=timeout,
                    TEST_HEARTBEAT=str(heartbeat))

    def test_python_group_deadline(self):
        directory = self.temporary()
        child, heartbeat = self.tree(directory)
        original = subprocess.Popen.communicate
        armed = []
        def communicate(process, *args, **kwargs):
            if not armed:
                self.wait_ready(heartbeat)
                armed.append(time.monotonic())
            return original(process, *args, **kwargs)
        with patch.object(subprocess.Popen, "communicate", communicate):
            with self.assertRaisesRegex(DeadlineExceeded, "forever-child.*deadline"):
                run([str(child)], deadline=.2)
        self.assertLess(time.monotonic()-armed[0], 5)
        self.stopped(heartbeat)
        self.doCleanups()

    def test_shell_kill_after(self):
        directory = self.temporary()
        child, heartbeat = self.tree(directory)
        helper = Path(__file__).with_name('shell.sh')
        result = run(["bash", "-c", 'source "$1"; bounded 1 "$2"', 'proof', str(helper), str(child)],
                     env=self.shell_ready_env(directory, heartbeat), stdout=subprocess.PIPE,
                     stderr=subprocess.PIPE, text=True, deadline=40)
        self.assertIn(result.returncode, (124,137))
        self.assertIn("forever-child: deadline", result.stderr)
        self.assertLess(time.monotonic()-float(Path(str(heartbeat)+".armed").read_text()), 5)
        self.stopped(heartbeat)
        self.doCleanups()

    def test_shell_exited_leader_still_kills_grandchild(self):
        directory = self.temporary()
        child, heartbeat = self.tree(directory, ignore_leader=False)
        helper = Path(__file__).with_name('shell.sh')
        result = run(["bash", "-c", 'source "$1"; bounded 1 "$2"', 'proof', str(helper), str(child)],
                     env=self.shell_ready_env(directory, heartbeat), stdout=subprocess.PIPE,
                     stderr=subprocess.PIPE, text=True, deadline=40)
        self.assertIn(result.returncode, (124,137))
        self.assertIn("forever-child: deadline", result.stderr)
        self.stopped(heartbeat)
        self.doCleanups()

    def test_shell_signal_exit_is_not_deadline(self):
        helper = Path(__file__).with_name('shell.sh')
        result = run(["bash", "-c", 'source "$1"; shift; bounded 2 "$@"', 'proof', str(helper), "sh", "-c", "kill -KILL $$"],
                     stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, deadline=4)
        self.assertNotIn("deadline", result.stderr)
        self.assertNotEqual(result.returncode, 0)

    def test_shell_exported_wrapper_and_pipeline(self):
        helper = Path(__file__).with_name('shell.sh')
        script = chr(10).join(['source "$1"', 'cat() { bounded 2 cat "$@"; }',
                           'export -f cat bounded',
                           'printf payload | bounded 4 bash -c cat'])
        result = run(["bash", "-c", script, "proof", str(helper)],
                     stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, deadline=6)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout, "payload")

    def test_signal_exit_is_not_deadline(self):
        result = run(["sh", "-c", "kill -KILL $$"], deadline=2)
        self.assertEqual(result.returncode, -signal.SIGKILL)


if __name__ == '__main__':
    unittest.main()
