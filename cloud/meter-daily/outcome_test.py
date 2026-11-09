#!/usr/bin/env python3
"""outcome.py names a relation budget day apart from a broken run, and submit.sh commits exactly the three it should."""
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

OUTCOME = Path(os.environ.get('METER_OUTCOME_UNDER_TEST', Path(__file__).resolve().parent / 'outcome.py'))


def outcome(*arguments):
    # -I, as run.sh and submit.sh run it.
    return subprocess.run([sys.executable, '-I', str(OUTCOME), *arguments], capture_output=True, text=True, timeout=30)


class MeterOutcome(unittest.TestCase):
    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory()
        self.addCleanup(self.scratch.cleanup)
        self.report = Path(self.scratch.name) / 'report.json'

    def meter(self, code):
        result = outcome('meter', str(code), str(self.report))
        self.assertEqual(result.returncode, 0, result.stderr)
        return result.stdout.splitlines()

    def write(self, **fields):
        self.report.write_text(json.dumps(dict({'trees': {}}, **fields)) + '\n')

    def test_zero_is_succeeded(self):
        self.write(relation_budget={'status': 'ok', 'why': 'assignable cache 299,624 pairs, within 1,000,000'})
        self.assertEqual(self.meter(0), ['Succeeded'])

    def test_three_red_is_over_budget_with_its_why(self):
        self.write(relation_budget={'status': 'red', 'assignable': 1_000_001,
                                    'why': 'assignable cache 1,000,001 pairs, past 1,000,000'})
        self.assertEqual(self.meter(3), ['OverBudget', 'assignable cache 1,000,001 pairs, past 1,000,000'])

    def test_three_unknown_is_budget_unknown_with_its_why(self):
        self.write(relation_budget={'status': 'unknown', 'why': 'relation report missing: run limited or did not complete'})
        self.assertEqual(self.meter(3), ['BudgetUnknown', 'relation report missing: run limited or did not complete'])

    def test_three_without_a_verdict_is_budget_unknown(self):
        self.write()
        self.assertEqual(self.meter(3)[0], 'BudgetUnknown')

    def test_three_without_a_report_is_budget_unknown(self):
        self.assertEqual(self.meter(3)[0], 'BudgetUnknown')

    def test_three_with_an_unreadable_report_is_budget_unknown(self):
        self.report.write_text('{"relation_budget": ')
        self.assertEqual(self.meter(3)[0], 'BudgetUnknown')

    def test_three_never_reads_ok_as_green(self):
        self.write(relation_budget={'status': 'ok', 'why': 'assignable cache 299,624 pairs, within 1,000,000'})
        self.assertEqual(self.meter(3)[0], 'BudgetUnknown')

    def test_any_other_exit_is_failed(self):
        self.write(relation_budget={'status': 'red', 'why': 'assignable cache 1,000,001 pairs, past 1,000,000'})
        for code in (1, 124, 2):
            with self.subTest(code=code):
                self.assertEqual(self.meter(code), ['Failed'])


class Submittable(unittest.TestCase):
    def test_exactly_the_three_are_submitted(self):
        for name, expected in (('Succeeded', 0), ('OverBudget', 0), ('BudgetUnknown', 0),
                               ('Failed', 1), ('Running', 1), ('None', 1), ('', 1)):
            with self.subTest(outcome=name):
                self.assertEqual(outcome('submittable', name).returncode, expected)


if __name__ == '__main__':
    unittest.main()
