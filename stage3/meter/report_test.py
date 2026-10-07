import json
import os
import subprocess
from pathlib import Path
import tempfile
import unittest

from report import report, report_pair


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
        self.assertEqual(result['totals'], {'files': 5, 'source_files': 4, 'checker': 3, 'lowering': 1,
                                           'checker_whole_program': 3, 'checker_own_file': 4})
        rows = {row['file']: row for row in result['files']}
        self.assertTrue(rows['src/compiler/refused.ts']['checker'])
        self.assertFalse(rows['src/compiler/refused.ts']['lowering'])
        self.assertFalse(rows['src/compiler/bad.ts']['lowering_attempted'])
        self.assertIn('| src/compiler/bad.ts | fail | pass | blocked |', (self.run / 'report.md').read_text())
        self.assertIn('| src/compiler/refused.ts | pass | pass | fail |', (self.run / 'report.md').read_text())

    def test_planted_diagnostic_changes_only_its_own_file(self):
        before = self.render()
        location = self.tree / 'src/compiler/good.ts'
        diagnostic = f'{location}:1:7: error TS2322: planted type mismatch\n  elaboration'
        # The same imported diagnostic appears in multiple loaded programs.
        for record in self.records:
            if record['kind'] != 'error':
                record['kind'] = 'checker'
                record['diagnostics'] = [diagnostic]
        after = self.render()
        changed = [row['file'] for old, row in zip(before['files'], after['files'])
                   if old['checker_own_file'] != row['checker_own_file']]
        self.assertEqual(changed, ['src/compiler/good.ts'])
        self.assertEqual(after['totals']['checker_whole_program'], 0)
        self.assertEqual(after['totals']['checker_own_file'], 3)
        self.assertEqual(after['files'][1]['own_file_diagnostics'], 1)

    def test_global_and_external_diagnostics_are_not_attributed_to_root(self):
        self.records[0]['kind'] = 'checker'
        self.records[0]['diagnostics'] = ['error TS2318: missing global',
                                        '/external/lib.d.ts:1:1: error TS2322: external']
        result = self.render()
        self.assertEqual(result['totals']['checker_own_file'], 4)
        self.assertEqual(result['unlocated_diagnostics'], 1)
        self.assertEqual(result['external_diagnostics'], 1)

    def test_pair_preserves_area_fields_and_prints_four_numbers_first(self):
        self.render()
        for label in ['main', 'area']:
            directory = self.run / label
            directory.mkdir()
            (directory / 'census.jsonl').write_text((self.run / 'census.jsonl').read_text())
        for record in self.records:
            if record['kind'] != 'error':
                record['kind'] = 'checker'
                record['diagnostics'] = [f'{self.tree}/src/compiler/good.ts:1:1: error TS2322: area only']
        (self.run / 'area/census.jsonl').write_text(''.join(json.dumps(row) + '\n' for row in self.records))
        result = report_pair(self.tree, self.tree, self.run, 'stamp', 'main-sha', 'area-sha')
        self.assertEqual(result['files'], result['trees']['area']['files'])
        self.assertEqual(result['trees']['main']['tree_ref'], 'origin/main')
        self.assertEqual((self.run / 'report.md').read_text().splitlines()[:2],
                         ['Whole program: main: 3/4; area: 0/4', 'Own file: main: 4/4; area: 3/4'])

    def test_malformed_diagnostic_is_rejected(self):
        self.records[0]['diagnostics'] = ['unexpected diagnostic encoding']
        with self.assertRaisesRegex(ValueError, 'unrecognized diagnostic format'):
            self.render()

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


@unittest.skipUnless(os.environ.get('CENSUS_BINARY'), 'set CENSUS_BINARY for the real checker attribution probe')
class CensusAttributionTests(unittest.TestCase):
    def test_type_error_in_dependency_changes_only_dependency_own_file(self):
        with tempfile.TemporaryDirectory() as scratch:
            tree = Path(scratch) / 'adapted'
            compiler = tree / 'src/compiler'
            compiler.mkdir(parents=True)
            dependency = compiler / 'dependency.a'
            dependency.write_text('export const value: number = 1;\n')
            (compiler / 'main.a').write_text('import { value } from "./dependency.a"; console.log(String(value));\n')
            observations = []
            for label in ['before', 'planted']:
                run = Path(scratch) / label
                run.mkdir()
                if label == 'planted':
                    dependency.write_text('export const value: number = "wrong";\n')
                with (run / 'census.log').open('w') as log:
                    subprocess.run([os.environ['CENSUS_BINARY'], str(compiler), str(run / 'census.jsonl')],
                                   check=True, stdout=log, stderr=log)
                observations.append(report(tree, run, label))
            before, after = observations
            self.assertEqual(before['totals']['checker_whole_program'], 2)
            self.assertEqual(before['totals']['checker_own_file'], 2)
            self.assertEqual(after['totals']['checker_whole_program'], 0)
            self.assertEqual(after['totals']['checker_own_file'], 1)
            changed = [row['file'] for old, row in zip(before['files'], after['files'])
                       if old['checker_own_file'] != row['checker_own_file']]
            self.assertEqual(changed, ['src/compiler/dependency.a'])


if __name__ == '__main__':
    unittest.main()
