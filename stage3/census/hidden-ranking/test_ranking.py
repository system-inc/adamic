"""A synthetic census with independently enumerated byte ownership."""
import json
from pathlib import Path
import unittest
import ranking


class RankingTest(unittest.TestCase):
    def test_synthetic_census(self):
        fixture = json.loads((Path(__file__).parent / 'fixture.json').read_text())
        entries = ranking.boundaries(fixture['census'], fixture['stock'], Path('/fixture'))
        residual = ranking.hidden.subtract([(e['start'], e['end']) for e in entries],
            fixture['independently_examined'])
        self.assertEqual([list(span) for span in residual], fixture['files']['fixture.a']['hidden_ranges'])
        result = ranking.partition(entries, fixture['files'], {})
        actual = {r['reason']: r['bytes_revealed_if_fixed_alone'] for r in result['ranked_reasons']}
        self.assertEqual(actual, {'outer reason': 40, 'inner reason': 20})
        self.assertEqual(result['hidden_bytes'], 60)
        # Oracle enumerates the known bytes, without interval union or sweeps.
        expected = {i: 'outer reason' for i in list(range(10, 40)) + list(range(50, 60))}
        expected.update({i: 'inner reason' for i in range(70, 90)})
        observed = {}
        for segment in result['attributed_segments']:
            for byte in range(segment['start'], segment['end']):
                self.assertNotIn(byte, observed)
                observed[byte] = segment['reason']
        self.assertEqual(observed, expected)
        self.assertEqual(sum(e['attributed_hidden_bytes'] for e in entries), 60)
        inner = next(r for r in result['ranked_reasons'] if r['reason'] == 'inner reason')
        self.assertEqual(inner['boundary_count'], 2)
        self.assertEqual(inner['contributing_boundary_count'], 1)
        self.assertEqual(inner['examples'], ['fixture.a:8'])

    def test_nested_byte_mutant(self):
        fixture = json.loads((Path(__file__).parent / 'fixture.json').read_text())
        entries = ranking.boundaries(fixture['census'], fixture['stock'], Path('/fixture'))
        with self.assertRaisesRegex(AssertionError, 'each hidden byte must count exactly once'):
            ranking.partition(entries, fixture['files'], {}, mutant=True)
        print('CAUGHT: nested-byte mutant by exact-once accounting assertion')

    def test_conflicting_equal_and_crossing_spans(self):
        for first, second in [((0, 10), (0, 10)), ((0, 8), (2, 10))]:
            entries = [dict(file='fixture.a', start=a, end=b, kind='NotYet', reason=str(i),
                where=f'fixture.a:{i+1}:1') for i, (a,b) in enumerate((first, second))]
            result = ranking.partition(entries, {'fixture.a': dict(hidden_ranges=[[0,10]], hidden_bytes=10)}, {})
            conflict = sum(r['bytes_revealed_if_fixed_alone'] for r in result['other_causes'])
            self.assertEqual(conflict, 10 if first == second else 6)

    def test_checker_ancestor(self):
        entries = [dict(file='fixture.a',start=0,end=10,kind='checker',reason='checker-rejected body',where='fixture.a:1:1'),
            dict(file='fixture.a',start=2,end=8,kind='NotYet',reason='nested',where='fixture.a:2:1')]
        result = ranking.partition(entries, {'fixture.a': dict(hidden_ranges=[[0,10]],hidden_bytes=10)}, {})
        self.assertEqual(result['ranked_reasons'][0]['bytes_revealed_if_fixed_alone'], 0)
        self.assertEqual(result['other_causes'][0]['bytes_revealed_if_fixed_alone'], 10)


if __name__ == '__main__':
    unittest.main(verbosity=2)
