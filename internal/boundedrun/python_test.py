"""Executable hang proofs for the shared Python and shell entry points."""
import os
from pathlib import Path
import signal
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

    def tree(self, directory):
        heartbeat, pidfile, child = [Path(directory)/name for name in ("heartbeat", "pid", "forever-child")]
        child.write_text(f"#!/bin/sh\necho $$ > '{pidfile}'\ntrap '' TERM\n(sh -c 'trap \"\" TERM; while :; do echo tick >> \"$1\"; sleep .01; done' child '{heartbeat}') &\nwait\n")
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

    def test_python_group_deadline(self):
        directory = self.temporary()
        child, heartbeat = self.tree(directory)
        start = time.monotonic()
        with self.assertRaisesRegex(DeadlineExceeded, "forever-child.*deadline"):
            run([str(child)], deadline=.2)
        self.assertLess(time.monotonic()-start, 3)
        self.stopped(heartbeat)
        self.doCleanups()

    def test_shell_kill_after(self):
        directory = self.temporary()
        child, heartbeat = self.tree(directory)
        helper = Path(__file__).with_name('shell.sh')
        start = time.monotonic()
        result = run(["bash", "-c", 'source "$1"; bounded 1 "$2"', 'proof', str(helper), str(child)],
                     env=dict(os.environ, ADAMIC_CHILD_DEADLINE="0.2"), stdout=subprocess.PIPE,
                     stderr=subprocess.PIPE, text=True, deadline=4)
        self.assertIn(result.returncode, (124,137))
        self.assertIn("forever-child: deadline", result.stderr)
        self.assertLess(time.monotonic()-start, 3)
        self.stopped(heartbeat)
        self.doCleanups()

    def test_signal_exit_is_not_deadline(self):
        result = run(["sh", "-c", "kill -KILL $$"], deadline=2)
        self.assertEqual(result.returncode, -signal.SIGKILL)


if __name__ == '__main__':
    unittest.main()
