"""Prove first introduction and subsequent baseline growth are distinct."""
import collections
import contextlib
import io
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest import mock

import cohere_gate as gate


class FirstLanding(unittest.TestCase):
    def test_missing_base_introduces_baseline_but_local_growth_still_fails(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            subprocess.run(['git', 'init', '-q', scratch], check=True)
            subprocess.run(['git', '-C', scratch, '-c', 'user.name=Test',
                            '-c', 'user.email=test@example.com', 'commit',
                            '--allow-empty', '-qm', 'base without gate'], check=True)
            subprocess.run(['git', '-C', scratch, 'update-ref',
                            'refs/remotes/origin/main', 'HEAD'], check=True)
            baseline = root / 'cloud/cohere-baseline.json'
            baseline.parent.mkdir()
            finding = {'kind': 'finding', 'path': str(root / 'program.ts'),
                       'rule': 'format', 'severity': 'error'}
            with mock.patch.object(gate, 'ROOT', root), mock.patch.object(gate, 'BASELINE', baseline):
                self.assertIsNone(gate.baseline_at('origin/main'))
                key = gate.identity(finding)
                baseline.write_text(json.dumps({key: 1}))
                output = io.StringIO()
                with mock.patch.object(gate, 'make_mirror', return_value=1), \
                     mock.patch.object(gate, 'audit', return_value=([finding], [])), \
                     mock.patch.dict('os.environ', {'COHERE_BINARY': 'unused'}), \
                     mock.patch('sys.argv', ['cohere_gate.py']), contextlib.redirect_stdout(output):
                    self.assertEqual(gate.main(), 0)
                self.assertIn('origin/main has no baseline; this commit introduces it', output.getvalue())
                # Missing integration baseline never waives the local reviewed ceiling.
                extra = dict(finding, path=str(root / 'new.ts'))
                with mock.patch.object(gate, 'make_mirror', return_value=2), \
                     mock.patch.object(gate, 'audit', return_value=([finding, extra], [])), \
                     mock.patch.dict('os.environ', {'COHERE_BINARY': 'unused'}), \
                     mock.patch('sys.argv', ['cohere_gate.py', '--record']), \
                     contextlib.redirect_stdout(io.StringIO()), contextlib.redirect_stderr(io.StringIO()):
                    self.assertEqual(gate.main(), 1)
                self.assertEqual(collections.Counter(json.loads(baseline.read_text())), {key: 1})
                # Once the base owns a baseline, growth is rejected against it.
                subprocess.run(['git', '-C', scratch, 'add', '.'], check=True)
                subprocess.run(['git', '-C', scratch, '-c', 'user.name=Test',
                                '-c', 'user.email=test@example.com', 'commit', '-qm', 'baseline'], check=True)
                subprocess.run(['git', '-C', scratch, 'update-ref', 'refs/remotes/origin/main', 'HEAD'], check=True)
                baseline.write_text(json.dumps({key: 1, gate.identity(extra): 1}))
                with mock.patch.object(gate, 'make_mirror', return_value=2), \
                     mock.patch.object(gate, 'audit', return_value=([finding, extra], [])), \
                     mock.patch.dict('os.environ', {'COHERE_BINARY': 'unused'}), \
                     mock.patch('sys.argv', ['cohere_gate.py']), contextlib.redirect_stdout(io.StringIO()):
                    with self.assertRaisesRegex(ValueError, 'baseline grew relative to the integration ref'):
                        gate.main()
