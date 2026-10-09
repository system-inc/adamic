#!/usr/bin/env python3
"""rerun_merge.py: a rerun lands on a run's kept verdicts only when nothing it kept changed inputs or failed.

usage: python3 cloud/integration/rerun_merge_test.py
"""
import importlib.util
from pathlib import Path
import unittest

specification = importlib.util.spec_from_file_location('rerun_merge', Path(__file__).with_name('rerun_merge.py'))
rerunMerge = importlib.util.module_from_spec(specification)
specification.loader.exec_module(rerunMerge)


def unit(name, inputs, verdict):
    return {'id': name, 'input_sha256': inputs, 'verdict': verdict}


# The base run at sha a: the slow test was killed at 90 s, everything else passed.
base = {'sha': 'a' * 40, 'units': [unit('lower/TestA', 'h1', 'passed'), unit('oracle/TestSlow', 'h2', 'killed'),
                                   unit('oracle/TestB', 'h3', 'passed')]}


class RerunMergeTests(unittest.TestCase):
    def check(self, units, descends=True):
        return rerunMerge.problems(base, {'sha': 'b' * 40, 'units': units}, descends)

    def test_a_clean_one_unit_rerun_lands(self):
        self.assertEqual(self.check([unit('lower/TestA', 'h1', 'kept'), unit('oracle/TestSlow', 'h2b', 'passed'),
                                     unit('oracle/TestB', 'h3', 'kept')]), [])

    def test_a_killed_test_split_into_new_units_lands_when_the_new_units_ran(self):
        # The P0 fix splits TestSlow: its id is gone and its shards are new ids, which can only have been run.
        self.assertEqual(self.check([unit('lower/TestA', 'h1', 'kept'), unit('oracle/TestSlow1', 'h4', 'passed'),
                                     unit('oracle/TestSlow2', 'h5', 'passed'), unit('oracle/TestB', 'h3b', 'passed')]), [])

    def test_a_kept_unit_whose_inputs_changed_is_refused_by_name(self):
        found = self.check([unit('lower/TestA', 'h1-moved', 'kept'), unit('oracle/TestSlow', 'h2b', 'passed'),
                            unit('oracle/TestB', 'h3', 'kept')])
        self.assertEqual(len(found), 1)
        self.assertIn('lower/TestA is kept but its inputs changed', found[0])

    def test_rerunning_a_unit_whose_inputs_changed_is_accepted(self):
        self.assertEqual(self.check([unit('lower/TestA', 'h1-moved', 'passed'), unit('oracle/TestSlow', 'h2b', 'passed'),
                                     unit('oracle/TestB', 'h3', 'kept')]), [])

    def test_keeping_the_red_unit_is_refused(self):
        found = self.check([unit('lower/TestA', 'h1', 'kept'), unit('oracle/TestSlow', 'h2', 'kept'), unit('oracle/TestB', 'h3', 'kept')])
        self.assertEqual(found, ['unit oracle/TestSlow is kept but was killed in the base run'])

    def test_a_rerun_that_fails_or_is_killed_again_is_refused(self):
        found = self.check([unit('lower/TestA', 'h1', 'kept'), unit('oracle/TestSlow', 'h2b', 'killed'), unit('oracle/TestB', 'h3', 'kept')])
        self.assertEqual(found, ['rerun unit oracle/TestSlow is killed'])

    def test_a_kept_unit_the_base_never_ran_is_refused(self):
        found = self.check([unit('lower/TestA', 'h1', 'kept'), unit('oracle/TestSlow', 'h2b', 'passed'), unit('oracle/TestB', 'h3', 'kept'),
                            unit('native/TestNew', 'h9', 'kept')])
        self.assertEqual(found, ['unit native/TestNew is kept but the base run never ran it'])

    def test_a_sha_that_doesnt_descend_from_the_base_is_refused(self):
        found = self.check([unit('lower/TestA', 'h1', 'kept'), unit('oracle/TestSlow', 'h2b', 'passed'), unit('oracle/TestB', 'h3', 'kept')],
                           descends=False)
        self.assertEqual(len(found), 1)
        self.assertIn("doesn't descend", found[0])

    def test_records_without_units_are_refused(self):
        self.assertIn('needs every unit listed', rerunMerge.problems(base, {'sha': 'b' * 40}, True)[0])


if __name__ == '__main__':
    unittest.main()
