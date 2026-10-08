"""Known-answer same-input deltas and mutants for their reconciliation checks."""
import copy
import unittest
from step05 import delta


class DeltaTest(unittest.TestCase):
    def test_delta_and_mutants(self):
        before = dict(total_bytes=100, hidden_bytes=60, hidden_share=.6, blocked_union_bytes=80,
                      independently_examined_bytes=20, files={'fixture.a': dict(bytes=100, sha256='known', hidden_bytes=60)})
        after = copy.deepcopy(before)
        after.update(hidden_bytes=40, hidden_share=.4, independently_examined_bytes=40)
        after['files']['fixture.a']['hidden_bytes'] = 40
        result = delta(before, after)
        self.assertEqual(result['hidden_bytes']['delta'], -20)
        self.assertAlmostEqual(result['hidden_share_percentage_point_delta'], -20)
        wrong = copy.deepcopy(after); wrong['hidden_bytes'] += 1
        with self.assertRaisesRegex(AssertionError, 'file deltas reconcile'):
            delta(before, wrong)
        wrong = copy.deepcopy(after); wrong['files']['fixture.a']['sha256'] = 'changed'
        with self.assertRaisesRegex(AssertionError, 'same source bytes'):
            delta(before, wrong)
        wrong = copy.deepcopy(after); wrong['total_bytes'] += 1
        with self.assertRaisesRegex(AssertionError, 'same denominator'):
            delta(before, wrong)
        print('CAUGHT: headline delta, changed-source and denominator mutants')


if __name__ == '__main__':
    unittest.main(verbosity=2)
