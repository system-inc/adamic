#!/usr/bin/env python3
"""Targeted summary-contract checks, including the real conflict-pass mutant."""
import copy
import importlib.util
import json
from pathlib import Path
import sys
import unittest

sys.dont_write_bytecode = True
HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('scratch_summary', HERE / 'scratch-summary.py')
checker = importlib.util.module_from_spec(spec)
spec.loader.exec_module(checker)


def passed_modes():
    report = dict(build_exit=0, node_exit=0, native_exit=0, native_diff_exit=0,
                  native_byte_mutant_diff_exit=1, comparison_control_exit=0, end_mutant_diff_exit=1)
    return [dict(split=i, exit=0, report=copy.deepcopy(report), full_tree_diff_exit=0) for i in [0, 1]]


class ScratchSummaryTests(unittest.TestCase):
    def test_success_requires_both_verified_modes(self):
        data = dict(status='pass', merges=[dict(exit=0, conflicts=[])], compiler=dict(exit=0), scanner=passed_modes(), scanner_native_evidence=dict(adamic_sha='a'*40, source_sha='b'*40, run_directory='run', node_sha256='c'*64, native_sha256='c'*64, comparison='comparison.stdout', mutant={'comparison_exit': 1}))
        self.assertEqual(checker.validate(data), [])
        data['scanner'][1]['report']['native_byte_mutant_diff_exit'] = 0
        self.assertIn('split 1: native_byte_mutant_diff_exit must be 1', checker.validate(data))

    def test_real_conflict_pass_mutant_is_caught_only_by_merge_rules(self):
        data = json.loads((HERE / 'evidence/scratch-run/conflict-summary.json').read_text())
        self.assertEqual(checker.validate(data), [])
        mutant = copy.deepcopy(data)
        mutant.update(status='pass', compiler={'exit': 0}, scanner=passed_modes(), scanner_native_evidence=dict(adamic_sha='a'*40, source_sha='b'*40, run_directory='run', node_sha256='c'*64, native_sha256='c'*64, comparison='comparison.stdout', mutant={'comparison_exit': 1}))
        problems = checker.validate(mutant)
        self.assertEqual(problems, ['merge conflict cannot be reported as pass', 'pass requires every merge to exit zero'])


    def test_pass_without_evidence_is_caught(self):
        data = dict(status='pass', merges=[], compiler=dict(exit=0), scanner=passed_modes())
        self.assertIn('pass requires scanner_native evidence: adamic_sha', checker.validate(data))

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
