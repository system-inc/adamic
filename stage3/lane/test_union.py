"""The union compares each identity's outcome and its multiplicity."""
import unittest
from merge_shards import outcome_difference


class UnionTests(unittest.TestCase):
    def test_missing_duplicate_and_swapped_outcomes(self):
        first = dict(task=dict(runner='compiler', file='one.a'), name=['suite', 'first'], status='pass')
        second = dict(task=dict(runner='compiler', file='one.a'), name=['suite', 'second'], status='fail')
        whole = [first, second]
        self.assertEqual(outcome_difference(whole, whole), ([], []))
        missing, extra = outcome_difference([first], whole)
        self.assertEqual(len(missing), 1)
        self.assertEqual(missing[0]['name'], 'suite second')
        self.assertEqual(extra, [])
        missing, extra = outcome_difference([first, first, second], whole)
        self.assertEqual(missing, [])
        self.assertEqual(len(extra), 1)
        self.assertEqual(extra[0]['count'], 1)
        swapped = [dict(first, status='fail'), dict(second, status='pass')]
        missing, extra = outcome_difference(swapped, whole)
        self.assertEqual(len(missing), 2)
        self.assertEqual(len(extra), 2)
