"""Known-answer accounting and duplicate-placement proof, without compiler changes."""
import copy
import json
from pathlib import Path
import unittest

from attribute import ROOT, aggregate, placement


def read(name):
    return json.loads((ROOT / name).read_text())


def nonzero(groups):
    return [{k: g[k] for k in ('step', 'hidden_bytes', 'reason_count')} |
            {'top_three': [r['reason'] for r in g['top_three']]}
            for g in groups if g['hidden_bytes']]


class AttributionTests(unittest.TestCase):
    def test_three_reason_known_answer(self):
        fixture = read('fixture.json')
        self.assertEqual(nonzero(aggregate(fixture['reasons'], fixture['mapping'])),
                         fixture['expected'])
        self.assertEqual([placement(r)[0] for r in fixture['reasons']], [22, 22, 15])

    def test_two_step_mutant(self):
        fixture = read('fixture.json')
        mutant = fixture['mapping'] + [fixture['mutant']]
        with self.assertRaisesRegex(ValueError, fixture['expected_mutant_error']):
            aggregate(fixture['reasons'], mutant)
        print('CAUGHT two-step mutant: Object.entries assigned to both 22 and 30')

    def test_input_and_coverage_mutants(self):
        fixture = read('fixture.json')
        rows, mapping = fixture['reasons'], fixture['mapping']
        cases = [(rows + [rows[0]], mapping, 'duplicate input reason'),
                 (rows, mapping[:-1], 'mapping must cover exactly'),
                 (rows[:-1], mapping, 'mapping must cover exactly')]
        for bad in (-1, True, '60'):
            changed = copy.deepcopy(rows)
            changed[0]['bytes_revealed_if_fixed_alone'] = bad
            cases.append((changed, mapping, 'invalid credited bytes'))
        changed = copy.deepcopy(mapping)
        changed[0]['step'] = 999
        cases.append((rows, changed, 'unknown roadmap step'))
        for changed_rows, changed_mapping, expected in cases:
            with self.subTest(error=expected):
                with self.assertRaisesRegex(ValueError, expected):
                    aggregate(changed_rows, changed_mapping)
        print('CAUGHT seven input/coverage mutants')

    def test_real_ledger(self):
        source, mapping, result = read('input.json'), read('mapping.json'), read('RESULT.json')
        groups = aggregate(source['reasons'], mapping)
        self.assertEqual(groups, result['ranked_groups'])
        self.assertEqual(sum(g['hidden_bytes'] for g in groups), source['hidden_bytes'])
        self.assertEqual(source['hidden_bytes'], result['hidden_bytes'])
        self.assertEqual(len(source['reasons']), 1727)
        for row, entry in zip(source['reasons'], mapping):
            self.assertEqual((entry['step'], entry['rule']), placement(row))
        # Independent direct sums for every step, including unplaced, and top-three ordering.
        for group in groups:
            keys = {(m['kind'], m['reason']) for m in mapping if m['step'] == group['step']}
            values = [r for r in source['reasons'] if (r['kind'], r['reason']) in keys]
            self.assertEqual(group['hidden_bytes'], sum(r['bytes_revealed_if_fixed_alone'] for r in values))
            self.assertEqual(group['reason_count'], len(values))
            self.assertEqual(group['top_three'], sorted(values, key=lambda r:
                             (-r['bytes_revealed_if_fixed_alone'], r['kind'], r['reason']))[:3])
        self.assertEqual(sum(r['bytes_revealed_if_fixed_alone'] for r in source['reasons']
                             if r['kind'] in ('Refused', 'NotYet')), 1949534)
        print('PASS pinned ledger: 1,727 reasons; 3,654,880 bytes conserved')

    def test_known_answer_detects_wrong_bytes(self):
        fixture = read('fixture.json')
        rows = copy.deepcopy(fixture['reasons'])
        rows[0]['bytes_revealed_if_fixed_alone'] += 1
        with self.assertRaises(AssertionError):
            self.assertEqual(nonzero(aggregate(rows, fixture['mapping'])), fixture['expected'])
        print('CAUGHT +1 credited-byte mutant by independent known answer')

    def test_operation_precedence(self):
        # Names in the measured program must not become roadmap feature evidence.
        for reason in ('reading excludeRegex', 'reading currentNamespace', 'reading exception',
                       'a value of type CapturedThis'):
            self.assertIsNone(placement({'kind': 'NotYet', 'reason': reason})[0])
        for reason, step in [('a Map of T | undefined', 19),
                             ('a function returning T | undefined', 16),
                             ('a value of type Path', 11),
                             ('for...in over an array', 20),
                             ('a value of type string | undefined', 17),
                             ('a computed field name', 22),
                             ('RegExp with a nonconstant pattern', 23)]:
            self.assertEqual(placement({'kind': 'NotYet', 'reason': reason})[0], step)


if __name__ == '__main__':
    unittest.main(verbosity=2)
