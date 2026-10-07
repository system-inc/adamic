import json
from pathlib import Path
import tempfile
import unittest

from report import report


class ReportTests(unittest.TestCase):
    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory()
        self.addCleanup(self.scratch.cleanup)
        self.tree = Path(self.scratch.name) / 'adapted'
        self.run = Path(self.scratch.name) / 'run'
        self.run.mkdir()
        compiler = self.tree / 'src/compiler'
        compiler.mkdir(parents=True)
        self.records = []
        for name, kind in [('good.ts', 'accepted'), ('refused.ts', 'Refused'),
                           ('later.ts', 'NotYet'), ('bad.ts', 'checker'),
                           ('tsconfig.json', 'error')]:
            path = compiler / name
            path.write_text('')
            self.records.append({'roots': [str(path)], 'kind': kind})
        self.records.append({'roots': sorted(str(path) for path in compiler.glob('*.ts')), 'kind': 'checker'})

    def render(self):
        (self.run / 'census.jsonl').write_text(''.join(json.dumps(row) + '\n' for row in self.records))
        report(self.tree, self.run, '20261007T000000Z')
        return json.loads((self.run / 'report.json').read_text())

    def test_checker_and_lowering_are_separate(self):
        result = self.render()
        self.assertEqual(result['totals'], {'files': 5, 'source_files': 4, 'checker': 3, 'lowering': 1})
        rows = {row['file']: row for row in result['files']}
        self.assertTrue(rows['src/compiler/refused.ts']['checker'])
        self.assertFalse(rows['src/compiler/refused.ts']['lowering'])
        self.assertFalse(rows['src/compiler/bad.ts']['lowering_attempted'])
        self.assertIn('| src/compiler/bad.ts | fail | blocked |', (self.run / 'report.md').read_text())
        self.assertIn('| src/compiler/refused.ts | pass | fail |', (self.run / 'report.md').read_text())

    def test_missing_file_is_rejected(self):
        del self.records[0]
        with self.assertRaisesRegex(ValueError, 'coverage incomplete'):
            self.render()

    def test_duplicate_file_is_rejected(self):
        self.records.insert(0, self.records[0])
        with self.assertRaisesRegex(ValueError, 'duplicate census file'):
            self.render()

    def test_missing_whole_program_is_rejected(self):
        self.records.pop()
        with self.assertRaisesRegex(ValueError, 'coverage incomplete'):
            self.render()

    def test_unknown_kind_is_rejected(self):
        self.records[0]['kind'] = 'maybe'
        with self.assertRaisesRegex(ValueError, 'unknown census kind'):
            self.render()

    def test_whole_program_roots_are_checked(self):
        self.records[-1]['roots'].pop()
        with self.assertRaisesRegex(ValueError, 'whole-program'):
            self.render()


if __name__ == '__main__':
    unittest.main()
