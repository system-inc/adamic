"""Hold file partition identity and the oversized-file refusal to observations."""
import copy
from collections import Counter, defaultdict
import gzip
import json
from pathlib import Path
import unittest
from measure_plan import make_plan


def observation(runner, file, milliseconds):
    return dict(task=dict(runner=runner, file=file), duration=milliseconds,
                passes=[dict(name=[runner, file, 'baseline'])], errors=[], passing=1)


class PlanTests(unittest.TestCase):
    def test_file_over_budget_is_invalid_even_alone(self):
        plan = make_plan([observation('compiler', 'case.ts', 31000)], [])
        self.assertEqual(plan['status'], 'invalid')
        self.assertEqual(plan['oversized_files'][0]['file'], 'case.ts')

    def test_unit_suites_from_one_file_stay_together(self):
        rows = [observation('unittest', name, 1000) for name in ['first', 'second']]
        sources = [dict(title=name, file='src/testRunner/unittests/one.ts') for name in ['first', 'second']]
        plan = make_plan(rows, sources, .1)
        self.assertEqual(len({row['shard'] for row in plan['tests']}), 1)

    def test_order_does_not_change_assignment_or_inventory(self):
        rows = [observation('compiler', name, 1000) for name in ['a.ts', 'b.ts', 'c.ts']]
        self.assertEqual(make_plan(rows, [], .5), make_plan(rows[::-1], [], .5))

    def test_duplicate_task_is_rejected(self):
        row = observation('compiler', 'case.ts', 1)
        # Empty results isolate the task guard from the separate test-identity guard.
        row.update(passes=[], passing=0)
        with self.assertRaisesRegex(ValueError, 'duplicate upstream task'):
            make_plan([row, copy.deepcopy(row)], [])

    def test_recorded_whole_inventory_has_exact_union(self):
        evidence = Path(__file__).with_name('evidence') / 'shards'
        listing = evidence / 'tests-and-shards.jsonl.gz'
        # This is a required real fixture, not an optional evidence-dependent test.
        with gzip.open(listing, 'rt') as stream:
            rows = [json.loads(line) for line in stream]
        whole = json.loads((evidence / 'whole-report.json').read_text())
        self.assertEqual(Counter(row['status'] for row in rows),
                         Counter({'pass': whole['counts']['passing'], 'fail': whole['counts']['failing']}))
        self.assertEqual(len({row['id'] for row in rows}), len(rows))
        files = defaultdict(set)
        for row in rows:
            files[row['file']].add(row['shard'])
        self.assertTrue(all(len(shards) == 1 for shards in files.values()))


if __name__ == '__main__':
    unittest.main()
