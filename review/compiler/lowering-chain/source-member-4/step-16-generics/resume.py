"""Recover a file-independent census only with matching overlap and complete coverage."""
import json
from pathlib import Path
import sys
import unittest

MUTANT = '--mutant' in sys.argv

def merge(parts, expected_files):
    metadata = parts[0][0]
    records = {}
    overlaps = []
    for rows in parts:
        assert rows[0] == metadata, 'measurement metadata changed'
        for record in rows[1:]:
            filename = record['file']
            if filename in records:
                if not MUTANT:
                    assert record == records[filename], 'overlap record changed'
                overlaps.append(filename)
            records[filename] = record
    assert overlaps, 'no repeated checkpoint file'
    assert set(records) == set(expected_files), 'file coverage changed'
    return [metadata] + [records[name] for name in sorted(records)], overlaps

class ResumeTests(unittest.TestCase):
    def test_changed_overlap_is_rejected(self):
        metadata = {'measurement': 'no output'}
        first = [metadata, {'file':'a', 'findings':[]}, {'file':'b', 'findings':[]}]
        second = [metadata, {'file':'b', 'findings':['planted']}, {'file':'c', 'findings':[]}]
        with self.assertRaisesRegex(AssertionError, 'overlap record changed'):
            merge([first,second], ['a','b','c'])
    def test_complete_unchanged_overlap(self):
        metadata = {'measurement': 'no output'}
        records = [{'file':name,'findings':[]} for name in ['a','b','c']]
        rows, overlap = merge([[metadata]+records[:2], [metadata]+records[1:]], ['a','b','c'])
        self.assertEqual(rows, [metadata]+records)
        self.assertEqual(overlap, ['b'])

if __name__ == '__main__':
    if '--test' in sys.argv or MUTANT:
        unittest.main(argv=[sys.argv[0]])
    else:
        before, *parts, output = [Path(p) for p in sys.argv[1:]]
        def read(path):
            return [json.loads(line) for line in path.read_text().splitlines()]
        expected = [r['file'] for r in read(before)[1:]]
        rows, overlap = merge([read(part) for part in parts], expected)
        output.write_text(''.join(json.dumps(row,separators=(',',':'))+'\n' for row in rows))
        print(f'PASS: {len(rows)-1} files; identical checkpoint overlaps: {overlap}')
