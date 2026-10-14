#!/usr/bin/env python3
"""Targeted summary-contract checks, including the real conflict-pass mutant."""
import copy
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

sys.dont_write_bytecode = True
HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('scratch_summary', HERE / 'scratch-summary.py')
checker = importlib.util.module_from_spec(spec)
spec.loader.exec_module(checker)
runner_spec = importlib.util.spec_from_file_location('scratch_runner', HERE / 'scratch-run.py')
runner = importlib.util.module_from_spec(runner_spec)
runner_spec.loader.exec_module(runner)


def passed_modes():
    report = dict(build_exit=0, node_exit=0, native_exit=0, native_diff_exit=0,
                  native_byte_mutant_diff_exit=1, comparison_control_exit=0, end_mutant_diff_exit=1)
    return [dict(split=i, exit=0, report=copy.deepcopy(report), full_tree_diff_exit=0) for i in [0, 1]]


class ScratchPipelineTests(unittest.TestCase):
    def test_copied_apply_imports_cache_helper_without_pythonpath(self):
        with tempfile.TemporaryDirectory() as directory:
            pipeline = runner.copy_pipeline(Path(directory))
            env = dict(os.environ)
            env.pop('PYTHONPATH', None)
            env['PYTHONDONTWRITEBYTECODE'] = '1'
            result = subprocess.run([sys.executable, str(pipeline / 'apply.py'), '--key'],
                                    cwd=directory, env=env, capture_output=True, text=True, timeout=10)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertRegex(result.stdout.strip(), r'^[0-9a-f]{64}$')

    def test_copied_cache_helper_respects_declared_build_cache(self):
        with tempfile.TemporaryDirectory() as directory:
            pipeline = runner.copy_pipeline(Path(directory))
            declared = Path(directory) / 'go-products'
            env = dict(os.environ, ADAMIC_SANDBOX='1', ADAMIC_BUILD_CACHE_DIR=str(declared))
            env.pop('PYTHONPATH', None)
            env.pop('STAGE3_CACHE', None)
            result = subprocess.run([sys.executable, '-B', '-c',
                                     'from caches import stage3_cache; print(stage3_cache())'],
                                    cwd=pipeline, env=env, capture_output=True, text=True, timeout=10)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(result.stdout.strip(), str(declared / 'adamic-stage3'))

    def test_missing_configured_source_names_path_and_exact_pin(self):
        with tempfile.TemporaryDirectory() as directory:
            out = Path(directory)
            pipeline = runner.copy_pipeline(out)
            missing = out / 'missing-source'
            pin = json.loads((pipeline / 'source.json').read_text())
            def unexpected(*args, **kwargs):
                self.fail('a missing configured source must fail before installing anything')
            with self.assertRaises(RuntimeError) as caught:
                runner.prepare_gate_inputs(out, pipeline, out / 'cache',
                                           {'ADAMIC_TYPESCRIPT_SOURCE': str(missing)},
                                           unexpected, unexpected)
            self.assertIn(str(missing), str(caught.exception))
            self.assertIn(pin['commit'], str(caught.exception))
            self.assertIn('ADAMIC_TYPESCRIPT_SOURCE', str(caught.exception))

    def test_wrong_source_commit_is_rejected_before_api_install(self):
        with tempfile.TemporaryDirectory() as directory:
            out = Path(directory)
            pipeline = runner.copy_pipeline(out)
            source = out / 'wrong-source'
            source.mkdir()
            pin = json.loads((pipeline / 'source.json').read_text())
            def unexpected(*args, **kwargs):
                self.fail('an incorrect source pin must fail before API installation')
            with self.assertRaises(RuntimeError) as caught:
                runner.prepare_gate_inputs(out, pipeline, out / 'cache',
                                           {'ADAMIC_TYPESCRIPT_SOURCE': str(source)},
                                           unexpected, lambda *args: '0' * 40)
            self.assertIn(pin['commit'], str(caught.exception))
            self.assertIn('checkout is at ' + '0' * 40, str(caught.exception))


class ScratchSummaryTests(unittest.TestCase):
    def test_success_requires_both_verified_modes(self):
        data = dict(status='pass', merges=[dict(exit=0, conflicts=[])], compiler=dict(exit=0), scanner=passed_modes())
        self.assertEqual(checker.validate(data), [])
        data['scanner'][1]['report']['native_byte_mutant_diff_exit'] = 0
        self.assertIn('split 1: native_byte_mutant_diff_exit must be 1', checker.validate(data))

    def test_real_conflict_pass_mutant_is_caught_only_by_merge_rules(self):
        data = json.loads((HERE / 'evidence/scratch-run/conflict-summary.json').read_text())
        self.assertEqual(checker.validate(data), [])
        mutant = copy.deepcopy(data)
        mutant.update(status='pass', compiler={'exit': 0}, scanner=passed_modes())
        problems = checker.validate(mutant)
        self.assertEqual(problems, ['merge conflict cannot be reported as pass', 'pass requires every merge to exit zero'])

    def test_no_compilation_after_conflict(self):
        data = json.loads((HERE / 'evidence/scratch-run/conflict-summary.json').read_text())
        data['compiler'] = {'exit': 0}
        self.assertIn('compiler/scanner must not run after a merge conflict', checker.validate(data))

    def test_missing_mode_cannot_pass(self):
        data = dict(status='pass', merges=[], compiler={'exit': 0}, scanner=passed_modes()[:1])
        self.assertIn('pass requires split 0 and 1', checker.validate(data))

    def test_original_ref_has_the_prior_44_conflicts(self):
        old = json.loads((HERE / 'evidence/scratch-run/original-conflict-summary.json').read_text())
        current = json.loads((HERE / 'evidence/scratch-run/conflict-summary.json').read_text())
        old_paths = set(old['merges'][-1]['conflicts'])
        current_paths = set(current['merges'][-1]['conflicts'])
        self.assertEqual(len(old_paths), 44)
        self.assertEqual(len(current_paths), 43)
        self.assertEqual(old_paths-current_paths, {'internal/native/runtime/node_process.c'})
        self.assertEqual(current_paths-old_paths, set())


if __name__ == '__main__':
    unittest.main()
