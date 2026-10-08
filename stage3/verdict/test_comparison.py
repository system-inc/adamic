"""Independent comparison checks with byte, ordering and identity mutants."""
import json
from pathlib import Path
import tempfile
import unittest
from comparison import classify, compare_results

A = b'a.ts(1,1): error TS2322: bad\n'
B = b'b.ts(2,3): error TS2345: other\n'


def sample(stdout=A, stderr=b'', code=b'2\n'):
    return {'stdout': stdout, 'stderr': stderr, 'exit': code}


class ComparisonChecks(unittest.TestCase):
    def test_cause_mutants(self):
        original = sample(A + B)
        for mutant, cause in [(sample(B + A), 'diagnostic ordering'),
                              (sample((A + B).replace(b'bad', b'Bad')), 'message wording/presentation'),
                              (sample(A), 'checker diagnostic set'),
                              (sample(A + B, code=b'1\n'), 'exit status'),
                              (sample(A + B, stderr=b'x'), 'stderr'),
                              (sample(b"error TS5023: Unknown compiler option 'fake'.\n"), 'unsupported CLI option/value'),
                              (sample(A + B, code=b'124\n'), 'timeout')]:
            self.assertEqual(classify(original, mutant), cause)
        self.assertEqual(classify(sample(A), sample(A.replace(b'(1,1)', b'(1,2)'))), 'diagnostic locations')
        # A code mutation must not disappear into wording or ordering.
        self.assertEqual(classify(sample(A), sample(A.replace(b'2322', b'2323'))), 'checker diagnostic set')

    def fixture(self, directory, values, failed=False):
        directory.mkdir()
        suites = {}
        for suite in ('acceptance', 'tiny', 'baselines'):
            identifier = 'a.ts' if suite == 'baselines' else 'tiny'
            folder = directory / suite / ('00001_a' if suite == 'baselines' else 'tiny')
            folder.mkdir(parents=True)
            for stream, value in values.items():
                (folder / ('actual.' + stream)).write_bytes(value)
            (folder / 'actual.diagnostics').write_bytes(values['stdout'])
            if suite == 'baselines':
                (directory / suite / 'selection.json').write_text(json.dumps([{'source': identifier,
                    'configuration': '', 'source_sha256': 'source', 'expected_sha256': 'expected', 'options': {}}]))
            suites[suite] = {'total': 1, 'passed': int(not failed), 'failed': int(failed), 'excluded': 0,
                             'failures': [{'case': identifier, 'configuration': ''}] if failed else []}
        (directory / 'summary.json').write_text(json.dumps({'tsc': str(directory), 'upstream_commit': 'pin',
                                                           'harness_errors': [], 'suites': suites}))

    def test_exact_agreement_and_stream_mutants(self):
        for changed in (None, 'stdout', 'stderr', 'exit'):
            with tempfile.TemporaryDirectory() as scratch:
                root = Path(scratch)
                self.fixture(root / 'left', sample())
                values = sample()
                if changed:
                    values[changed] = {'stdout': A.replace(b'bad', b'Bad'), 'stderr': b'x', 'exit': b'1\n'}[changed]
                self.fixture(root / 'right', values, bool(changed))
                result = compare_results(root / 'left', root / 'right', root / 'out')
                self.assertEqual(result['success'], changed is None)
                for row in result['disagreements']:
                    self.assertEqual(row['streams'], [changed])
                self.assertEqual(len(result['disagreements']), 3 if changed else 0)

    def test_identical_wrong_outputs_do_not_succeed(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            for label in ('left', 'right'):
                self.fixture(root / label, sample(A.replace(b'bad', b'Bad')), True)
            result = compare_results(root / 'left', root / 'right', root / 'out')
            self.assertFalse(result['success'])
            self.assertEqual(result['suites']['baselines']['agreed_but_both_oracle_failed'], 1)

    def test_same_timeout_does_not_agree(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            for label in ('left', 'right'):
                self.fixture(root / label, sample(code=b'124\n'), True)
            result = compare_results(root / 'left', root / 'right', root / 'out')
            self.assertEqual({k: v for k, v in result['cause_counts'].items() if v}, {'timeout': 3})
            self.assertFalse(result['success'])

    def test_input_and_harness_error_mutants(self):
        for mutant in ('hash', 'identity', 'harness'):
            with tempfile.TemporaryDirectory() as scratch:
                root = Path(scratch)
                for label in ('left', 'right'):
                    self.fixture(root / label, sample())
                if mutant == 'harness':
                    p = root / 'right/summary.json';data = json.loads(p.read_text());data['harness_errors'] = [{'error': 'bad'}]
                else:
                    p = root / 'right/baselines/selection.json';data = json.loads(p.read_text());data[0]['source_sha256' if mutant == 'hash' else 'source'] = 'mutated'
                p.write_text(json.dumps(data))
                with self.assertRaisesRegex(RuntimeError, 'harness errors|populations or input hashes'):
                    compare_results(root / 'left', root / 'right', root / 'out')


if __name__ == '__main__':
    unittest.main()
