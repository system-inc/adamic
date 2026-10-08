#!/usr/bin/env python3
"""Durable retry and collection tests without contacting gate boxes."""
import contextlib
import importlib.util
import io
import json
import os
import signal
import subprocess
import time
from pathlib import Path
import sys
import tarfile
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('idle', Path(__file__).with_name('idle-jobs.py'))
idle = importlib.util.module_from_spec(spec)
spec.loader.exec_module(idle)
SHA = 'a' * 40


class IdleTests(unittest.TestCase):
    def test_retry_and_completion(self):
        with tempfile.TemporaryDirectory() as tmp:
            state = Path(tmp)
            (state / 'enabled').touch()
            initial = {SHA: {'next': 2001, 'completed': 0, 'pending': {'home': {'seed': 1, 'count': 2000}}}}
            (state / 'state.json').write_text(json.dumps(initial))
            launches = []
            result = 'unfinished'

            def ssh(box, script, *args):
                if args == ('probe',):
                    return 'idle'
                if args[0] == 'start':
                    launches.append(args)
                    return 'started'
                return result

            with patch.dict(os.environ, {'ADAMIC_IDLE_STATE': tmp, 'ADAMIC_IDLE_BOXES': 'home'}), \
                    patch.object(sys, 'argv', ['idle-jobs.py', '--once']), \
                    patch.object(idle, 'ssh', ssh), patch.object(idle, 'collect'), \
                    patch.object(idle.subprocess, 'check_output', return_value=SHA + '\trefs/heads/main\n'), \
                    contextlib.redirect_stdout(io.StringIO()):
                idle.main()
                self.assertEqual(launches[-1][2:4], ('1', '2000'))
                self.assertEqual(json.loads((state / 'state.json').read_text()), initial)
                result = 'done'
                (state / 'enabled').unlink()
                idle.main()
                saved = json.loads((state / 'state.json').read_text())[SHA]
                self.assertEqual(saved['pending'], {})
                self.assertEqual(saved['by_box']['home'], 2000)
                self.assertEqual(saved['next'], 2001)
                self.assertEqual(len(launches), 1)

    def test_collection_dedupes_bytes_and_signature(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            raw = root / 'raw'
            raw.mkdir()
            for key, box in [('b' * 64, 'home'), ('c' * 64, 'cloud')]:
                folder = raw / key
                folder.mkdir()
                for name, text in {'program.a': 'original', 'reduced.a': 'minimal', 'signature.txt': 'crash:a store\n', 'seed': '1', 'main': SHA, 'box': box}.items():
                    (folder / name).write_text(text)
            archive = root / 'all.tar'
            with tarfile.open(archive, 'w') as tar:
                tar.add(raw, arcname='.')

            def run(*args, **kwargs):
                kwargs['stdout'].write(archive.read_bytes())

            output = io.StringIO()
            with patch.object(idle.subprocess, 'run', run), contextlib.redirect_stdout(output):
                idle.collect('home', root / 'collected')
                idle.collect('cloud', root / 'collected')
            self.assertEqual(len(list((root / 'collected').iterdir())), 1)
            self.assertEqual(len(output.getvalue().splitlines()), 1)
            self.assertIn('crash:a store', output.getvalue())

    def test_disabled_coordinator_skips_fresh_and_pending_jobs(self):
        for pending in (False, True):
            with self.subTest(pending=pending), tempfile.TemporaryDirectory() as tmp:
                state = Path(tmp)
                (state / 'enabled').touch()
                data = {SHA: {'next': 2001, 'completed': 0, 'pending':
                             {'home': {'seed': 1, 'count': 2000}} if pending else {}}}
                (state / 'state.json').write_text(json.dumps(data))
                calls = []

                def ssh(box, script, *args):
                    calls.append(args)
                    return 'disabled' if args == ('probe',) else 'running'

                output = io.StringIO()
                with patch.dict(os.environ, {'ADAMIC_IDLE_STATE': tmp, 'ADAMIC_IDLE_BOXES': 'home'}), \
                        patch.object(sys, 'argv', ['idle-jobs.py', '--once']), \
                        patch.object(idle, 'ssh', ssh), patch.object(idle, 'collect'), \
                        patch.object(idle.subprocess, 'check_output', return_value=SHA + '\trefs/heads/main\n'), \
                        contextlib.redirect_stdout(output):
                    idle.main()
                self.assertEqual(calls, [('probe',)])
                self.assertEqual(json.loads((state / 'state.json').read_text()), data)
                self.assertEqual(output.getvalue().count('home: disabled;'), 1)

    def test_unmarked_session_survivor_disables_and_is_killed(self):
        with tempfile.TemporaryDirectory() as tmp:
            child_file = Path(tmp) / 'child.pid'
            parent = subprocess.Popen(
                ['bash', '-c', 'env -u ADAMIC_IDLE_JOB sleep 120 & echo "$!" > "$1"; wait',
                 'parent', str(child_file)],
                env={**os.environ, 'ADAMIC_IDLE_JOB': '1'}, start_new_session=True)
            unrelated = subprocess.Popen(['sleep', '120'], start_new_session=True)
            child = None
            try:
                deadline = time.monotonic() + 3
                while not child_file.exists() or not child_file.read_text().strip():
                    self.assertLess(time.monotonic(), deadline, 'child startup timed out')
                    time.sleep(.01)
                child = int(child_file.read_text())
                # Wait for env to exec sleep and drop the inherited marker.
                while True:
                    with open(f'/proc/{child}/environ', 'rb') as handle:
                        if b'ADAMIC_IDLE_JOB=1' not in handle.read().split(b'\0'):
                            break
                    self.assertLess(time.monotonic(), deadline)
                    time.sleep(.01)
                result = subprocess.run(['bash', str(Path(__file__).with_name('idle-preempt.sh'))],
                                        env={**os.environ, 'HOME': tmp}, capture_output=True,
                                        text=True, timeout=12, check=True)
                self.assertRegex(result.stdout, r'^idle preemption: 1 marked, 1 survivors in their sessions, \d+ ms\n$')
                parent.wait(timeout=1)
                self.assertEqual(parent.returncode, -signal.SIGKILL)
                if Path(f'/proc/{child}/stat').exists():
                    fields = Path(f'/proc/{child}/stat').read_text().rsplit(')', 1)[1].split()
                    self.assertIn(fields[0], ('Z', 'X'))
                self.assertIsNone(unrelated.poll())
                disabled = Path(tmp) / 'idle/disabled'
                report = disabled.read_text()
                self.assertIn(f'pid={child} session={parent.pid}', report)
                self.assertIn('sleep 120', report)
                self.assertRegex(report, r'\d{4}-\d{2}-\d{2}T')
                # A later clean preemption must preserve the report.
                again = subprocess.run(['bash', str(Path(__file__).with_name('idle-preempt.sh'))],
                                       env={**os.environ, 'HOME': tmp}, capture_output=True,
                                       text=True, timeout=12, check=True)
                self.assertRegex(again.stdout, r'^idle preemption: 0 marked, 0 survivors in their sessions, \d+ ms\n$')
                self.assertEqual(disabled.read_text(), report)
                print(result.stdout.strip())
            finally:
                try:
                    os.killpg(parent.pid, signal.SIGKILL)
                except ProcessLookupError:
                    pass
                parent.wait()
                unrelated.kill()
                unrelated.wait()

    def test_full_gate_embeds_same_preemption(self):
        full = Path(__file__).with_name('full-gate-main.sh').read_text()
        start = full.index('# BEGIN idle preemption')
        end = full.index('# END idle preemption') + len('# END idle preemption')
        block = full[start:end].replace('\\\\', '\\').replace('\\$', '$').replace('\\`', '`')
        expected = Path(__file__).with_name('idle-preempt.sh').read_text()
        self.assertEqual(block, expected[expected.index('# BEGIN idle preemption'):].strip())


if __name__ == '__main__':
    unittest.main()
