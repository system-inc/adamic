"""Pure stream-schema checks; no compiler or go subprocess."""
import unittest
from stream_records import census_records


class LatentRelationStreamRecordsTests(unittest.TestCase):
    def test_pass_ledger_survives_stream_and_eligibility_pb47(self):
        from unit_status import eligibility_units
        old = {'where': 'only.a:1:1', 'status': 'attempted'}
        new = dict(old, pass_ledger=[{'name': 'some-step', 'status': 'not_run_scoped'}])
        file = {'file': 'only.a', 'units': [new], 'findings': []}
        self.assertEqual(census_records([{'status': 'measurement'}, file])[1]['units'][0], new)
        self.assertEqual(eligibility_units([old]), eligibility_units([new]))

    def test_partial_header_has_no_invented_end_answer(self):
        header = {'status': 'measurement', 'diagnostics': ['retained']}
        self.assertEqual(census_records([header]), [header])
        self.assertNotIn('relation_overflow', header)

    def test_completed_trailer_is_not_counted_as_a_file(self):
        header = {'status': 'measurement'}
        file = {'file': 'only.a', 'units': [], 'findings': []}
        trailer = {'status': 'relation_report',
                   'relation_overflow': {'after_load': '', 'end': ''}, 'findings': []}
        rows = census_records([header, file, trailer])
        self.assertEqual(len(rows), 2)
        self.assertEqual(rows[1], file)
        self.assertEqual(rows[0]['relation_overflow'], trailer['relation_overflow'])
        self.assertNotIn('relation_overflow', header)

    def test_misplaced_or_unfinished_trailer_is_rejected(self):
        header = {'status': 'measurement'}
        trailer = {'status': 'relation_report',
                   'relation_overflow': {'after_load': '', 'end': ''}, 'findings': []}
        for rows in ([header, trailer, {'file': 'late.a'}], [header, trailer, trailer],
                     [header, dict(trailer, relation_overflow={'after_load': ''})]):
            with self.subTest(rows=rows), self.assertRaises(ValueError):
                census_records(rows)


if __name__ == '__main__':
    unittest.main()
