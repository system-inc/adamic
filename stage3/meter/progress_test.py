"""Full-output proof, missing-mutant rejection and regression-only resets."""
import copy
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import tempfile
import unittest

SOURCE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('meter_progress_under_test',
    os.environ.get('METER_PROGRESS_WRITER_UNDER_TEST', SOURCE / 'progress.py'))
writer = importlib.util.module_from_spec(spec)
spec.loader.exec_module(writer)


class ProgressTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.repository = Path(self.temporary.name)
        self.run = self.repository / 'stage3/meter/runs/synthetic'
        self.run.mkdir(parents=True)
        self.progress = json.loads((SOURCE.parent / 'progress.json').read_text())
        self.progress['milestones'] = dict.fromkeys(self.progress['milestones'], False)
        self.progress['evidence'] = dict.fromkeys(self.progress['milestones'], None)
        self.save(self.repository / 'stage3/progress.json', self.progress)
        output = b'full scanner output\nEOF\n'
        for name in ('node.stdout', 'native.stdout'):
            (self.run / name).write_bytes(output)
        (self.run / 'native-mutant.stdout').write_bytes(bytes([output[0] ^ 1]) + output[1:])
        (self.run / 'comparison.log').write_text('full-output comparison exit=0\n')
        (self.run / 'mutant-comparison.log').write_text('full-output comparison exit=1: byte 0 differs\n')
        self.record = dict(adamic_sha='a' * 40, source_sha='b' * 40,
            run_directory='stage3/meter/runs/synthetic', node_sha256=hashlib.sha256(output).hexdigest(),
            native_sha256=hashlib.sha256(output).hexdigest(), node_output='node.stdout',
            native_output='native.stdout', comparison=dict(exit=0, log='comparison.log'),
            mutant=dict(output='native-mutant.stdout', comparison=dict(exit=1, log='mutant-comparison.log')))

    def save(self, path, data):
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(json.dumps(data))

    def update(self, record):
        self.save(self.run / 'milestones.json', {'scanner_native': record})
        result = writer.update(self.repository, self.run)
        self.assertEqual(result, json.loads((self.run / 'progress.json').read_text()))
        self.assertEqual(result, json.loads((self.repository / 'stage3/progress.json').read_text()))
        return result

    def test_complete_synthetic_run_flips_scanner(self):
        result = self.update(self.record)
        self.assertTrue(result['milestones']['scanner_native'])
        self.assertFalse(result['milestones']['parser_native'])
        self.assertEqual(result['evidence']['scanner_native']['adamic_sha'], 'a' * 40)

    def test_missing_mutant(self):
        del self.record['mutant']
        result = self.update(self.record)
        self.assertFalse(result['milestones']['scanner_native'])
        self.assertIn('missing mutant', result['reasons']['scanner_native'])

    def test_missing_sha_or_comparison_does_not_flip(self):
        for name in ('adamic_sha', 'source_sha', 'run_directory', 'comparison'):
            with self.subTest(field=name):
                record = copy.deepcopy(self.record)
                del record[name]
                self.assertFalse(self.update(record)['milestones']['scanner_native'])

    def test_hashes_must_cover_full_output(self):
        digest = hashlib.sha256(b'full scanner').hexdigest()
        self.record.update(node_sha256=digest, native_sha256=digest)
        result = self.update(self.record)
        self.assertFalse(result['milestones']['scanner_native'])
        self.assertIn('full output', result['reasons']['scanner_native'])

    def test_mutant_must_be_one_byte_and_caught(self):
        original = (self.run / 'native-mutant.stdout').read_bytes()
        (self.run / 'native-mutant.stdout').write_bytes(b'XX' + (self.run / 'native.stdout').read_bytes()[2:])
        self.assertFalse(self.update(self.record)['milestones']['scanner_native'])
        (self.run / 'native-mutant.stdout').write_bytes(original)
        self.record['mutant']['comparison']['exit'] = 0
        self.assertFalse(self.update(self.record)['milestones']['scanner_native'])

    def test_prior_true_survives_missing_evidence(self):
        self.assertTrue(self.update(self.record)['milestones']['scanner_native'])
        del self.record['mutant']
        result = self.update(self.record)
        self.assertTrue(result['milestones']['scanner_native'])
        self.assertIn('no verified regression', result['reasons']['scanner_native'])

    def test_real_output_regression_can_reset(self):
        self.assertTrue(self.update(self.record)['milestones']['scanner_native'])
        output = b'wrong scanner output\nEOF\n'
        (self.run / 'native.stdout').write_bytes(output)
        self.record.update(status='regression', native_sha256=hashlib.sha256(output).hexdigest(),
                           comparison=dict(exit=1, log='mutant-comparison.log'))
        result = self.update(self.record)
        self.assertFalse(result['milestones']['scanner_native'])
        self.assertIn('regression:', result['reasons']['scanner_native'])

    def test_named_run_and_blocked_first_stop(self):
        self.save(self.run / 'proof.json', {'scanner_native': self.record})
        reference = dict(run_directory=self.record['run_directory'], evidence_file='proof.json')
        self.assertTrue(self.update(reference)['milestones']['scanner_native'])
        self.save(self.repository / 'stage3/progress.json', self.progress)
        self.save(self.run / 'build-report.json', dict(native='blocked at compilation', build_exit=1))
        (self.run / 'build.stderr').write_text('adamic: /scratch/src/compiler/debug.ts:7:1: refuses a namespace\n')
        blocked = dict(run_directory=self.record['run_directory'], blocked_report='build-report.json', build_log='build.stderr')
        result = self.update(blocked)
        self.assertFalse(result['milestones']['scanner_native'])
        self.assertIn('debug.ts:7:1', result['reasons']['scanner_native'])
        self.assertIn(self.record['run_directory'], result['reasons']['scanner_native'])

    def test_unknown_requirements_fail_loudly(self):
        self.progress['evidence_required'].append('new_proof: a future requirement')
        self.save(self.repository / 'stage3/progress.json', self.progress)
        with self.assertRaisesRegex(ValueError, 'unsupported evidence_required'):
            self.update(self.record)

    def test_old_blocked_observation_preserves_true_but_explicit_build_regression_resets(self):
        self.assertTrue(self.update(self.record)['milestones']['scanner_native'])
        self.save(self.run / 'build-report.json', dict(native='blocked at compilation', build_exit=1))
        (self.run / 'build.stderr').write_text('adamic: /scratch/src/compiler/debug.ts:7:1: refuses a namespace\n')
        blocked = dict(run_directory=self.record['run_directory'], blocked_report='build-report.json', build_log='build.stderr')
        self.save(self.repository / 'stage3/meter/progress-inputs.json', {'scanner_native': blocked})
        self.save(self.run / 'milestones.json', {})
        self.assertTrue(writer.update(self.repository, self.run)['milestones']['scanner_native'])
        blocked.update(status='regression', adamic_sha='c' * 40, source_sha='d' * 40)
        result = self.update(blocked)
        self.assertFalse(result['milestones']['scanner_native'])
        self.assertIn('regression: native build stops', result['reasons']['scanner_native'])


if __name__ == '__main__':
    unittest.main()
