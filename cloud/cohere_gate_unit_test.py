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


class BootstrapRatchet(unittest.TestCase):
    def git(self, *arguments):
        return subprocess.check_output(['git', '-C', str(self.root), *arguments]).decode().strip()

    def commit(self, message):
        self.git('add', '.')
        self.git('-c', 'user.name=Test', '-c', 'user.email=test@example.com',
                 'commit', '--allow-empty', '-qm', message)
        return self.git('rev-parse', 'HEAD')

    def run_gate(self, findings, *arguments):
        with mock.patch.object(gate, 'ROOT', self.root), \
             mock.patch.object(gate, 'BASELINE', self.baseline), \
             mock.patch.object(gate, 'make_mirror', return_value=len(findings)), \
             mock.patch.object(gate, 'audit', return_value=(findings, [])), \
             mock.patch.dict('os.environ', {'COHERE_BINARY': 'unused'}), \
             mock.patch('sys.argv', ['cohere_gate.py', *arguments]), \
             contextlib.redirect_stdout(io.StringIO()), contextlib.redirect_stderr(io.StringIO()):
            return gate.main()

    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory()
        self.addCleanup(self.scratch.cleanup)
        self.root = Path(self.scratch.name)
        self.git('init', '-q')
        self.baseline = self.root / 'cloud/cohere-baseline.json'
        self.baseline.parent.mkdir()
        self.first = {'kind': 'finding', 'path': str(self.root / 'internal/first.ts'),
                      'rule': 'format', 'severity': 'error'}
        self.second = dict(self.first, path=str(self.root / 'stage1/second.ts'))
        self.third = dict(self.first, path=str(self.root / 'stage3/third.ts'))
        with mock.patch.object(gate, 'ROOT', self.root):
            self.first_key = gate.identity(self.first)
        self.baseline.write_text(json.dumps({self.first_key: 1}))
        self.initial_commit = self.commit('initial baseline')
        self.git('update-ref', 'refs/remotes/origin/main', 'HEAD')

    def test_bootstrap_then_growth_fails_and_record_lowers_ceiling(self):
        reason = 'new scope brings debt into view'
        self.assertEqual(self.run_gate([self.first, self.second], '--bootstrap', reason), 0)
        document = json.loads(self.baseline.read_text())
        provenance = document['bootstrap']
        self.assertEqual(provenance['reason'], reason)
        self.assertEqual(provenance['commit'], self.initial_commit)
        self.assertEqual(provenance['counts'], {'total': 2,
                         'byDirectory': {'internal': 1, 'stage1': 1}, 'byRule': {'format': 2}})
        self.assertEqual(provenance['newFindings'], {'total': 1,
                         'byDirectory': {'stage1': 1}, 'byRule': {'format': 1}})
        # Until committed, metadata does not waive either reviewed ceiling.
        with self.assertRaisesRegex(ValueError, 'baseline grew'):
            self.run_gate([self.first, self.second], '--record')
        bootstrap_commit = self.commit('explicit bootstrap')
        self.assertEqual(self.run_gate([self.first, self.second]), 0)
        before = self.baseline.read_bytes()
        for arguments in ((), ('--record',)):
            self.assertEqual(self.run_gate([self.first, self.second, self.third], *arguments), 1)
            self.assertEqual(self.baseline.read_bytes(), before)
        self.assertEqual(self.run_gate([self.first], '--record'), 0)
        reduced = json.loads(self.baseline.read_text())
        self.assertEqual(reduced['findings'], {self.first_key: 1})
        self.assertEqual(reduced['bootstrap'], provenance)  # Historical debt stays visible.
        self.assertEqual(self.run_gate([self.first]), 0)
        self.assertEqual(self.run_gate([self.first, self.second], '--record'), 1)
        self.commit('debt removed')
        self.assertEqual(self.run_gate([self.first]), 0)
        self.assertEqual(self.run_gate([self.first, self.second]), 1)
        # Once integration owns the bootstrap, its own ceiling is enforced normally.
        self.git('update-ref', 'refs/remotes/origin/main', bootstrap_commit)
        self.assertEqual(self.run_gate([self.first]), 0)
        self.assertEqual(self.run_gate([self.first, self.second], '--record'), 1)
        # Editing local provenance and findings cannot raise committed HEAD's ceiling.
        self.baseline.write_text(json.dumps(document))
        with self.assertRaisesRegex(ValueError, 'committed HEAD'):
            self.run_gate([self.first, self.second], '--record')

    def test_committing_hand_edited_growth_does_not_reuse_bootstrap_permission(self):
        self.assertEqual(self.run_gate([self.first, self.second], '--bootstrap', 'scope'), 0)
        self.commit('explicit bootstrap')
        document = json.loads(self.baseline.read_text())
        with mock.patch.object(gate, 'ROOT', self.root):
            document['findings'][gate.identity(self.third)] = 1
        self.baseline.write_text(json.dumps(document))
        self.commit('hand edited growth without a bootstrap')
        for arguments in ((), ('--record',)):
            with self.assertRaisesRegex(ValueError, 'committed bootstrap'):
                self.run_gate([self.first, self.second, self.third], *arguments)

    def test_committed_reduction_cannot_be_raised_by_hand(self):
        self.assertEqual(self.run_gate([self.first, self.second], '--bootstrap', 'scope'), 0)
        self.commit('explicit bootstrap')
        original = self.baseline.read_bytes()
        self.assertEqual(self.run_gate([self.first], '--record'), 0)
        self.commit('lower ceiling')
        self.baseline.write_bytes(original)
        self.commit('resurrect resolved debt by hand')
        with self.assertRaisesRegex(ValueError, 'committed bootstrap'):
            self.run_gate([self.first, self.second], '--record')

    def test_bootstrap_requires_reason_and_cannot_be_combined_with_record(self):
        before = self.baseline.read_bytes()
        for arguments in (('--bootstrap',), ('--bootstrap', ''), ('--bootstrap', '  '),
                          ('--bootstrap', 'scope', '--record')):
            with self.assertRaises(SystemExit) as error:
                self.run_gate([self.first], *arguments)
            self.assertEqual(error.exception.code, 2)
            self.assertEqual(self.baseline.read_bytes(), before)

    def test_bootstrap_requires_complete_audit(self):
        before = self.baseline.read_bytes()
        with mock.patch.object(gate, 'audit', side_effect=ValueError('incomplete lint check')):
            with mock.patch.object(gate, 'ROOT', self.root), \
                 mock.patch.object(gate, 'BASELINE', self.baseline), \
                 mock.patch.object(gate, 'make_mirror', return_value=1), \
                 mock.patch.dict('os.environ', {'COHERE_BINARY': 'unused'}), \
                 mock.patch('sys.argv', ['cohere_gate.py', '--bootstrap', 'scope']):
                with self.assertRaisesRegex(ValueError, 'incomplete lint check'):
                    gate.main()
        self.assertEqual(self.baseline.read_bytes(), before)
