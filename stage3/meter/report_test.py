import json
import os
import subprocess
from pathlib import Path
import tempfile
import unittest

from report import report, report_pair, latent_summary, MEASUREMENT, reason_owner, unowned_table


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
        for label in ['main', 'area']:
            latent = [{'measurement': MEASUREMENT, 'status': 'measurement', 'checker_rejected': True}]
            latent += [{'measurement': MEASUREMENT, 'file': str(path), 'findings': []}
                       for path in sorted((self.tree / 'src/compiler').glob('*.ts'))]
            if label == 'area':
                latent[1]['findings'] = [{'measurement': MEASUREMENT, 'kind': 'NotYet',
                                         'where': latent[1]['file'] + f':{line}:1',
                                         'reason': 'area-only blocker', 'text': 'area-only blocker'}
                                        for line in range(1, 11)]
            (self.run / label / 'latent.jsonl').write_text(''.join(json.dumps(row) + '\n' for row in latent))
        entry = self.tree / 'src/tsc/tsc.ts'
        entry.parent.mkdir(parents=True)
        entry.write_text('export {};\n')
        from entry_test import write_entry
        for label in ['main', 'area']:
            write_entry(self.tree, self.run / label / 'tsc',
                        diagnostics=[] if label == 'main' else ['error TS2307: entry-only missing module'])
        result = report_pair(self.tree, self.tree, self.run, 'stamp', 'main-sha', 'area-sha')
        self.assertEqual(result['files'], result['trees']['area']['files'])
        text = (self.run / 'report.md').read_text()
        self.assertIn(MEASUREMENT, text)
        self.assertIn('| area-only blocker | OWNER BLANK |', text)
        self.assertLess(text.index('### Unowned'), text.index('Latent lowering, main:'))
        self.assertEqual(result['trees']['main']['latent_lowering']['totals']['NotYet'], 0)
        self.assertEqual(result['trees']['area']['latent_lowering']['totals']['NotYet'], 10)
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


class LatentReportTests(unittest.TestCase):
    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory()
        self.addCleanup(self.scratch.cleanup)
        self.run = Path(self.scratch.name)
        self.tree = self.run / 'tree'
        compiler = self.tree / 'src/compiler'
        compiler.mkdir(parents=True)
        self.source = compiler / 'main.a'
        self.source.write_text('')
        self.records = [{'measurement': MEASUREMENT, 'status': 'measurement', 'checker_rejected': True},
                        {'measurement': MEASUREMENT, 'file': str(self.source), 'findings': [
                            self.finding('NotYet', 'existing', 1), self.finding('Refused', 'policy', 2)]}]

    def finding(self, kind, reason, line):
        return {'measurement': MEASUREMENT, 'kind': kind, 'reason': reason,
                'where': f'{self.source}:{line}:1', 'text': reason}

    def render(self):
        (self.run / 'latent.jsonl').write_text(''.join(json.dumps(row) + '\n' for row in self.records))
        return latent_summary(self.tree, self.run)

    def test_planted_notyet_moves_only_its_reason_count(self):
        before = self.render()
        planted = self.finding('NotYet', 'planted', 3)
        self.records[1]['findings'] += [planted, planted]
        after = self.render()
        changed = {key: after['per_reason'].get(key, 0) - before['per_reason'].get(key, 0)
                   for key in before['per_reason'].keys() | after['per_reason'].keys()
                   if after['per_reason'].get(key, 0) != before['per_reason'].get(key, 0)}
        self.assertEqual(changed, {'NotYet: planted': 1})
        self.assertEqual(after['totals']['NotYet'], before['totals']['NotYet'] + 1)
        self.assertEqual(after['totals']['Refused'], before['totals']['Refused'])

    def test_unmatched_reason_is_owner_blank_and_in_unowned_first(self):
        self.records[1]['findings'] += [self.finding('NotYet', 'existing', line) for line in range(3, 12)]
        result = self.render()
        row = next(row for row in result['reason_rows'] if row['reason'] == 'existing')
        self.assertEqual(row['owner'], 'OWNER BLANK')
        rows, lines = unowned_table({'main': {'latent_lowering': result}, 'area': {'latent_lowering': result}})
        self.assertIn('### Unowned', lines)
        text = '\n'.join(lines)
        self.assertIn('| existing | OWNER BLANK | 10 | 0 | 10 | 0 |', text)
        self.assertTrue(any(row['reason'] == 'existing' for row in rows))

    def test_unowned_threshold_uses_either_tree_and_summarizes_tail(self):
        def row(reason, notyet, refused):
            return {'reason': reason, 'owner': 'OWNER BLANK', 'NotYet': notyet, 'Refused': refused}
        main = [row('nine on both', 9, 0), row('main edge', 6, 4), row('large', 11, 0), row('tiny', 1, 0)]
        area = [row('nine on both', 9, 0), row('area edge', 10, 0)]
        all_rows, lines = unowned_table({'main': {'latent_lowering': {'unowned_reasons': main}},
                                        'area': {'latent_lowering': {'unowned_reasons': area}}})
        text = '\n'.join(lines)
        self.assertIn('| main edge | OWNER BLANK | 6 | 4 | 0 | 0 |', text)
        self.assertIn('| area edge | OWNER BLANK | 0 | 0 | 10 | 0 |', text)
        self.assertNotIn('| nine on both |', text)
        self.assertNotIn('| tiny |', text)
        self.assertLess(text.index('| large |'), text.index('| main edge |'))
        self.assertIn('2 more unowned reasons, 19 sites in all', text)
        self.assertEqual(len(all_rows), 5)

    def test_exact_prefix_and_variance_owners(self):
        owners = json.loads((Path(__file__).parent / 'owners.json').read_text())
        self.assertEqual(reason_owner('a type predicate', owners), '01a1143b-d691')
        self.assertEqual(reason_owner('reading SyntaxKind.SomeFlag', owners), 'compiler/stage3-front')
        self.assertEqual(reason_owner('a method call through a structural signature on X', owners), '01a1143c')
        self.assertEqual(reason_owner('a value of type X seen as Y', owners), 'adaptation 70, stage 3')
        for name in ('ModifierFlags', 'Extension', 'NodeFlags', 'Comparison', 'ModuleKind', 'EmitFlags'):
            self.assertEqual(reason_owner('reading ' + name, owners), 'compiler/stage3-front')
        self.assertEqual(reason_owner('reading Error', owners), 'OWNER BLANK')
        self.assertEqual(reason_owner('reading addOutput', owners), 'OWNER BLANK')
        self.assertEqual(reason_owner('a function returning T seen as U', owners), 'adaptation 70, stage 3')
        self.assertEqual(reason_owner('a value of type any seen as X', owners), 'adaptation 70, stage 3')
        for reason, owner in {'||=': '01a113e3-a058', 'a string as a condition': '01a113e3-a058',
                              'a boolean | undefined as a condition': '01a113e3-a058',
                              'a namespace object used as a value': 'codex/namespaces-tsc',
                              'an index signature': '01a113e7', 'a value of type any': 'adaptation 40, stage 3',
                              "a cast the runtime can't check": '01a11410', 'a value of type unknown': '01a114ab'}.items():
            self.assertEqual(reason_owner(reason, owners), owner)
        self.assertEqual(reason_owner('unmatched', owners), 'OWNER BLANK')
        self.assertEqual(reason_owner('reading SyntaxKind', {'reading': 'general', 'reading SyntaxKind': 'specific'}), 'specific')

    def test_table_groups_same_reason_across_kinds(self):
        self.records[1]['findings'].append(self.finding('Refused', 'existing', 3))
        result = self.render()
        group = next(row for row in result['reason_rows'] if row['reason'] == 'existing')
        self.assertEqual((group['NotYet'], group['Refused'], group['count']), (1, 1, 2))

    def test_missing_source_is_rejected(self):
        self.records.pop()
        with self.assertRaisesRegex(ValueError, 'coverage incomplete'):
            self.render()

    def test_measurement_label_is_required(self):
        self.records[1]['findings'][0]['measurement'] = 'compiled'
        with self.assertRaisesRegex(ValueError, 'invalid latent finding'):
            self.render()

    def test_top_ten_excludes_dependency_skips_and_orders_ties(self):
        self.records[1]['findings'] += [self.finding('NotYet', f'reason {n:02}', n+3) for n in range(12)]
        self.records[1]['findings'] += [self.finding('SkippedDependency', 'skip', 20)]
        result = self.render()
        self.assertEqual(len(result['top_reasons']), 10)
        self.assertNotIn('SkippedDependency: skip', result['per_reason'])
        self.assertEqual(result['totals']['SkippedDependency'], 1)
        self.assertEqual(result['top_reasons'][0]['reason'], 'existing')


@unittest.skipUnless(os.environ.get('LATENT_CENSUS_BINARY'), 'set LATENT_CENSUS_BINARY for the real latent probe')
class LatentCensusTests(unittest.TestCase):
    def test_overlay_planted_notyet_moves_only_its_reason_count(self):
        with tempfile.TemporaryDirectory() as scratch:
            tree = Path(scratch) / 'tree'
            compiler = tree / 'src/compiler'
            compiler.mkdir(parents=True)
            (compiler / 'main.a').write_text('function target(): number { return 1; }\n'
                                             'function existing(value: any): number { return value; }\n'
                                             'function asserted(value: number | undefined): number { debugger; return value!; }\n'
                                             'function bad(): number { return "wrong"; }\n')
            observations = []
            for label in ['before', 'planted']:
                run = Path(scratch) / label
                run.mkdir()
                env = dict(os.environ, LATENT_ASSERT_NO_OUTPUT='1')
                env.pop('LATENT_MUTANT_FUNCTION', None)
                env.pop('LATENT_MUTANT_WHERE', None)
                if label == 'planted':
                    env['LATENT_MUTANT_FUNCTION'] = 'target'
                with (run / 'latent.log').open('w') as log:
                    subprocess.run([os.environ['LATENT_CENSUS_BINARY'], str(compiler), str(run / 'latent.jsonl')],
                                   check=True, env=env, stdout=log, stderr=log)
                observations.append(latent_summary(tree, run))
            before, after = observations
            self.assertTrue(before['checker_rejected'])
            self.assertGreater(before['totals']['NotYet'], 0)
            self.assertGreater(before['totals']['Refused'], 0)
            changed = {key: after['per_reason'].get(key, 0) - before['per_reason'].get(key, 0)
                       for key in before['per_reason'].keys() | after['per_reason'].keys()
                       if after['per_reason'].get(key, 0) != before['per_reason'].get(key, 0)}
            self.assertEqual(changed, {'NotYet: latent planted extra NotYet': 1})
            self.assertEqual(after['totals']['NotYet'], before['totals']['NotYet'] + 1)
            self.assertEqual(after['totals']['Refused'], before['totals']['Refused'])


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
