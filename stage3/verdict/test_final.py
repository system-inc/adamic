"""Exercise timing permission, count integrity and failures through real artifacts."""
import contextlib
import copy
import io
import json
from pathlib import Path
import statistics
import subprocess
import tempfile
import unittest
from unittest.mock import patch
import final


class FinalChecks(unittest.TestCase):
    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory()
        self.addCleanup(self.scratch.cleanup)
        self.root = Path(self.scratch.name)
        self.binary = self.root / 'compiler with spaces'
        self.binary.write_bytes(b'fixture executable\n')
        self.verdict = {'success': True, 'harness_errors': [], 'suites': {}}
        for name, count in [('acceptance', 301), ('tiny', 1), ('baselines', 2)]:
            self.verdict['suites'][name] = {'total': count, 'passed': count, 'failed': 0, 'deferred': 0, 'failures': []}
        self.calls = []
        self.verdict_exit = 0
        self.performance_exit = 0
        self.native_mismatches = []
        self.change_binary = False
        self.missing_summary = False
        self.missing_timing = False
        self.slow_native = False
        final.write_json(self.root / 'selection.json', {'configurations': 2})

    def child(self, argv, **kwargs):
        self.calls.append(argv)
        out = Path(argv[-1]);out.mkdir()
        if Path(argv[0]).name == 'run.sh' and '--tsc' in argv:
            if not self.missing_summary:
                final.write_json(out / 'summary.json', self.verdict)
            if self.change_binary:
                self.binary.write_bytes(b'different executable\n')
            return subprocess.CompletedProcess(argv, self.verdict_exit)
        self.assertEqual(argv[1:3], ['--native', str(self.binary)])
        self.assertTrue((out.parent / 'correctness/summary.json').exists())
        final.write_json(out / 'preflight.json', {'accepted': ['typescript-compiler'], 'excluded': [], 'native_mismatches': self.native_mismatches})
        rows = []
        for mode in final.MODES:
            if self.missing_timing and mode == 'native':
                continue
            start = 8.0 if self.slow_native and mode == 'native' else 1.0
            samples = [start + i / 100 for i in range(10)]
            rows.append({'input': 'typescript-compiler', 'compiler': mode, 'seconds': samples,
                         'mean': statistics.mean(samples), 'stdev': statistics.stdev(samples)})
        final.write_json(out / 'results.json', rows)
        return subprocess.CompletedProcess(argv, self.performance_exit)

    def run_fixture(self, folder='out'):
        out = self.root / folder;out.mkdir()
        with patch.object(final, 'ROOT', self.root), patch.object(final, 'PERFORMANCE', self.root / 'performance/run.sh'), patch.object(final.subprocess, 'run', side_effect=self.child), contextlib.redirect_stdout(io.StringIO()), contextlib.redirect_stderr(io.StringIO()):
            code = final.execute(self.binary, out)
        return code, json.loads((out / 'summary.json').read_text()), (out / 'summary.md').read_text()

    def fail_suites(self, empty=False):
        self.verdict['success'], self.verdict_exit = False, 1
        for row in self.verdict['suites'].values():
            row['failed'] = row['total'] if empty else 1
            row['passed'] = row['total'] - row['failed']
            row['failures'] = [{'case': 'fixture', 'differences': {'stdout': {'byte_offset': 0}}}] * row['failed']

    def test_A_reaches_timing_and_one_table(self):
        code, summary, text = self.run_fixture()
        self.assertEqual(code, 0)
        self.assertTrue(summary['success'])
        self.assertEqual(len(self.calls), 2)
        self.assertNotIn('--baseline-limit', self.calls[0])
        self.assertEqual(text.count('| Phase / input |'), 1)
        self.assertLess(text.index('correctness: baselines'), text.index('timing: typescript-compiler'))
        self.assertIn('met', text)

    def test_B_diagnostic_failure_never_calls_performance(self):
        self.fail_suites()
        code, summary, _ = self.run_fixture()
        self.assertEqual(code, 1)
        self.assertEqual(summary['performance']['status'], 'skipped')
        self.assertEqual(len(self.calls), 1)
        self.assertFalse((self.root / 'out/performance-command.json').exists())

    def test_C_zero_passes_never_calls_performance(self):
        self.fail_suites(empty=True)
        code, summary, _ = self.run_fixture()
        self.assertEqual(code, 1)
        self.assertTrue(all(row['passed'] == 0 for row in summary['correctness']['suites'].values()))
        self.assertEqual(len(self.calls), 1)

    def test_timing_permission_mutant_is_caught(self):
        self.fail_suites()
        with patch.object(final, 'verdict_passes', return_value=True):
            self.run_fixture('mutant')
        # A mutant authorizes timing despite a failed real artifact.
        with self.assertRaises(AssertionError):
            self.assertEqual(len(self.calls), 1, 'failed correctness reached performance')

    def test_deferred_and_count_mutants_refuse_timing(self):
        for mutation in ('deferred', 'total', 'passed'):
            with self.subTest(mutation=mutation):
                original = copy.deepcopy(self.verdict)
                self.verdict['suites']['baselines'][mutation] += 1
                self.calls = []
                code, summary, _ = self.run_fixture(mutation)
                self.assertEqual(code, 2)
                self.assertIn('harness_error', summary)
                self.assertEqual(len(self.calls), 1)
                self.verdict = original

    def test_missing_summary_refuses_timing(self):
        self.missing_summary = True
        code, summary, _ = self.run_fixture()
        self.assertEqual(code, 2)
        self.assertEqual(len(self.calls), 1)
        self.assertIn('harness_error', summary)

    def test_exit_summary_mutant_refuses_timing(self):
        self.verdict_exit = 1
        code, _, _ = self.run_fixture()
        self.assertEqual(code, 2)
        self.assertEqual(len(self.calls), 1)

    def test_executable_byte_mutant_refuses_timing(self):
        self.change_binary = True
        code, summary, _ = self.run_fixture()
        self.assertEqual(code, 2)
        self.assertIn('compiler changed', summary['harness_error'])
        self.assertEqual(len(self.calls), 1)

    def test_performance_preflight_failure(self):
        self.native_mismatches, self.performance_exit = ['typescript-compiler'], 1
        code, summary, _ = self.run_fixture()
        self.assertEqual(code, 1)
        self.assertFalse(summary['success'])
        self.assertEqual(summary['performance']['status'], 'failed')

    def test_missing_native_samples_mutant(self):
        self.missing_timing = True
        code, summary, _ = self.run_fixture()
        self.assertEqual(code, 2)
        self.assertIn('missing native', summary['harness_error'])

    def test_slow_native_is_reported_without_correctness_failure(self):
        self.slow_native = True
        code, summary, text = self.run_fixture()
        self.assertEqual(code, 0)
        self.assertFalse(summary['performance']['compiler_bar_met'])
        self.assertIn('missed', text)


if __name__ == '__main__':
    unittest.main()
