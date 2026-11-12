import json
import os
import subprocess
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

from report import (report, report_pair, latent_summary, MEASUREMENT, reason_owner, unowned_table, write_readme,
                    entry_table, latent_table)


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
        report(self.tree, self.run, '20261007T000000Z', '1' * 40)
        return json.loads((self.run / 'report.json').read_text())

    def test_missing_compiler_commit_is_an_error(self):
        with patch.dict(os.environ, {}, clear=True):
            with self.assertRaisesRegex(ValueError, 'compiler commit is required'):
                report(self.tree, self.run, 'stamp')

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

    def render_pair(self):
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
        if hasattr(self, 'relation_trailer'):
            from report import ENTRY_MEASUREMENT
            with (self.run / 'main/tsc/latent.jsonl').open('a') as output:
                output.write(json.dumps({'status': 'relation_report', 'measurement': ENTRY_MEASUREMENT,
                                         'relation_overflow': self.relation_trailer, 'findings': []}) + '\n')
        for label in ['main', 'area']:
            if hasattr(self, 'unit_times'):
                (self.run / label / 'latent-times.jsonl').write_text(
                    ''.join(json.dumps(row) + '\n' for row in self.unit_times))
        return report_pair(self.tree, self.tree, self.run, 'stamp', 'main-sha', 'area-sha', self.run, '1' * 40)

    def test_pair_preserves_area_fields_and_prints_four_numbers_first(self):
        result = self.render_pair()
        self.assertEqual(result['files'], result['trees']['area']['files'])
        text = (self.run / 'report.md').read_text()
        self.assertIn(MEASUREMENT, text)
        self.assertIn('Body lowering, main: 1/4; root-order artifacts: 0.', text)
        self.assertIn('| Root | Stop | Reason | Node result | Classification | Program result from entry census |', text)
        self.assertIn('| area-only blocker | OWNER BLANK |', text)
        self.assertLess(text.index('### Unowned'), text.index('Latent body lowering, main:'))
        self.assertEqual(result['trees']['main']['latent_lowering']['totals']['NotYet'], 0)
        self.assertEqual(result['trees']['area']['latent_lowering']['totals']['NotYet'], 10)
        self.assertEqual(result['trees']['main']['tree_ref'], 'origin/main')
        self.assertEqual((self.run / 'report.md').read_text().splitlines()[:2],
                         ['Whole program: main: 3/4; area: 0/4', 'Own file: main: 4/4; area: 3/4'])

    def test_relation_budget_reaches_report_json_and_heads_the_entry_section(self):
        from report import RELATION_CACHES
        self.relation_trailer = {'after_load': '', 'end': '',
                                 'cache_sizes': dict(dict.fromkeys(RELATION_CACHES, 0), assignable=1_000_001)}
        result = self.render_pair()
        disk = json.loads((self.run / 'report.json').read_text())
        self.assertEqual(disk['relation_budget'], result['relation_budget'])
        self.assertEqual(disk['relation_budget'], {'status': 'red', 'assignable': 1_000_001, 'threshold': 1_000_000,
                                                   'overflow_marked': False,
                                                   'why': 'assignable cache 1,000,001 pairs, past 1,000,000'})
        lines = (self.run / 'report.md').read_text().splitlines()
        heading = lines.index('tsc entry body lowering, main: measured on a checker-clean entry-root program.')
        self.assertEqual(lines[heading - 1],
                         'Relation budget: red; assignable 1,000,001 of 1,000,000 pairs; assignable cache 1,000,001 pairs, past 1,000,000.')

    def test_meter_unit_times_block_sorts_files_and_shortens_units(self):
        prefix = str(self.tree / 'src/compiler')
        self.unit_times = [
            {'file': prefix + '/fast.ts', 'where': prefix + '/fast.ts:3:1', 'wall_seconds': 1.25, 'cpu_seconds': 0.5},
            {'file': prefix + '/slow.ts', 'where': prefix + '/slow.ts:7:2', 'wall_seconds': 3.5, 'cpu_seconds': 2.0}]
        box = {'hostname': 'meter-box', 'nproc': 8, 'cpu_quota': None, 'total_memory': '1000 kB',
               'loadavg_start': '1 2 3 1/8 12', 'loadavg_end': '2 3 4 1/8 13'}
        (self.run / 'box.json').write_text(json.dumps(box))
        result = self.render_pair()
        self.assertEqual(result['box'], box)
        disk = json.loads((self.run / 'report.json').read_text())
        for label in ('main', 'area'):
            summary = disk['trees'][label]['latent_time']
            self.assertEqual((summary['wall_seconds'], summary['cpu_seconds']), (4.75, 2.5))
            self.assertEqual(summary['files'], [
                {'file': 'src/compiler/slow.ts', 'wall_seconds': 3.5, 'cpu_seconds': 2.0},
                {'file': 'src/compiler/fast.ts', 'wall_seconds': 1.25, 'cpu_seconds': 0.5}])
            self.assertEqual(summary['slowest_units'], [
                {'where': 'src/compiler/slow.ts:7:2', 'wall_seconds': 3.5, 'cpu_seconds': 2.0},
                {'where': 'src/compiler/fast.ts:3:1', 'wall_seconds': 1.25, 'cpu_seconds': 0.5}])
        text = (self.run / 'report.md').read_text()
        self.assertIn('Stage 3 meter stamp\nBox: meter-box;', text)
        self.assertIn("This box's times, not comparable across boxes: 4.750 wall seconds; 2.500 CPU seconds.", text)
        self.assertIn('| src/compiler/slow.ts | 3.500 | 2.000 |', text)

    def test_meter_old_run_has_no_unit_times(self):
        result = self.render_pair()
        for label in ('main', 'area'):
            self.assertIsNone(result['trees'][label]['latent_time'])
        self.assertIsNone(result['box'])
        self.assertEqual((self.run / 'report.md').read_text().count('no unit times recorded'), 2)

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
        self.assertEqual(reason_owner('a type predicate', owners), '#1w5bve2 (Optimizer)')
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
                              "a cast the runtime can't check": '01a11410', 'a value of type unknown': '01a114ab',
                              'an object refinement using an open numeric enum as a literal tag': '#drwm5pj (Lower)'}.items():
            self.assertEqual(reason_owner(reason, owners), owner)
        # Group 1's classes as the daily meter on Oct 27 words them (#bbssdcy), each with its task and owner.
        for reason, owner in {
            'a type predicate whose return is not proven (return expression is not a trusted check on node)': '#1w5bve2 (Optimizer)',
            'a type predicate whose return is not proven (there is no body proving this parameter)': '#1w5bve2 (Optimizer)',
            'an unproven predicate argument for parameter test (argument "isNodeArray")': '#1w5bve2 (Optimizer)',
            'checked predicate membership over an open tag domain': '#k1t2wp9 (Optimizer)',
            'an indirect call of a checked predicate': '#3hxh5jm (Workers)',
            'an indirect call of a checked predicate overload': '#3hxh5jm (Workers)',
            'checked predicate overload target ClassElement': '#c7w6j0y (Optimizer)',
            'a checked predicate without a reifiable target': '#1ge1bv8 (Diagnostics)',
            "a checked predicate result read by a CallExpression inside a condition, which its direction walk doesn't read": '#1761dah (Lower)',
            "a tag-proven predicate whose target has an optional field (name on Declaration) that the view can't check": '#1w5bve2 (Optimizer)',
            "node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started)": '#redxes2 (WebAssembly)',
            'node:module.require("crypto"): CommonJS loading with no ES-module answer yet (fix: import it with an import declaration)': '#redxes2 (WebAssembly)',
            'node:module.require("source-map-support"): CommonJS loading with no ES-module answer yet (fix: import it with an import declaration)': '#redxes2 (WebAssembly)',
            'node:module.require of a computed module path: CommonJS loading with no ES-module answer yet (fix: import each module it can load with an import declaration)': '#redxes2 (WebAssembly)',
            # The bare reason, the inspector require's (#cn4yapb), stays unowned: no CommonJS prefix matches it.
            'node:module.require': 'OWNER BLANK',
        }.items():
            self.assertEqual(reason_owner(reason, owners), owner)
        self.assertEqual(reason_owner('unmatched', owners), 'OWNER BLANK')
        self.assertEqual(reason_owner('reading SyntaxKind', {'reading': 'general', 'reading SyntaxKind': 'specific'}), 'specific')

    def test_table_groups_same_reason_across_kinds(self):
        self.records[1]['findings'].append(self.finding('Refused', 'existing', 3))
        result = self.render()
        group = next(row for row in result['reason_rows'] if row['reason'] == 'existing')
        self.assertEqual((group['NotYet'], group['Refused'], group['count']), (1, 1, 2))

    def test_panic_and_error_sites_are_named_from_src(self):
        panic = self.finding('panic', 'runtime error: invalid memory address or nil pointer dereference', 3)
        panic['text'] = 'lowering a call'
        error = dict(self.finding('error', 'unreadable', 4), where='generated:4:1')
        # A site repeated across attempts is still one site.
        self.records[1]['findings'] += [panic, panic, error]
        result = self.render()
        self.assertEqual(result['panic_sites'], [{'where': 'src/compiler/main.a:3:1', 'text': 'lowering a call',
                                                  'reason': 'runtime error: invalid memory address or nil pointer dereference'}])
        self.assertEqual(result['error_sites'], [{'where': 'generated:4:1', 'reason': 'unreadable', 'text': 'unreadable'}])
        lines = ['- panic `src/compiler/main.a:3:1`: runtime error: invalid memory address or nil pointer dereference, lowering a call',
                 '- error `generated:4:1`: unreadable, unreadable']
        for table in (latent_table('main', result), entry_table('main', {'lowering_census': result})):
            errors = next(index for index, line in enumerate(table) if line.startswith('Errors: 1; panics: 1'))
            self.assertEqual(table[errors + 1:errors + 3], lines)
            self.assertEqual(table[errors + 3], '')

    def test_no_panic_or_error_names_no_site(self):
        result = self.render()
        self.assertEqual((result['panic_sites'], result['error_sites']), ([], []))
        for table in (latent_table('main', result), entry_table('main', {'lowering_census': result})):
            errors = next(index for index, line in enumerate(table) if line.startswith('Errors: 0; panics: 0'))
            self.assertEqual(table[errors + 1], '')
            self.assertFalse(any(line.startswith('- ') for line in table))

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
                observations.append(report(tree, run, label, '1' * 40))
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

class PreviousRunTests(unittest.TestCase):
    def test_oct19_beside_committed_oct8(self):
        root = Path(__file__).resolve().parent
        current = json.loads((root / 'runs/20261019T233755Z.9xwivl15/report.json').read_text())
        with tempfile.TemporaryDirectory() as temporary:
            run = Path(temporary) / '20261019T233755Z.9xwivl15'
            run.mkdir()
            write_readme(run, current, root / 'runs')
            text = (run / 'README.md').read_text()
            self.assertIn('| Current | 295df3bc6aa3d47933ccf675da1c33e745856313 | 2/79 | 56/79 | 2/79 | 235 |', text)
            self.assertIn('| 20261008T171021Z.iBLy18 | 45487a809f89885a3fc651cd590e7dabf31362dc | 2/79 | 56/79 | 0/79 | 320 |', text)


class RootOrderTests(unittest.TestCase):
    setUp = ReportTests.setUp
    render = ReportTests.render

    def readiness(self, *, entry_reason=None, attempted=True, node_throws=True):
        from entry_test import write_entry
        from report import READINESS_REASONS, ENTRY_MEASUREMENT
        self.reason = sorted(READINESS_REASONS)[0]
        self.root = self.tree / 'src/compiler/later.ts'
        self.stop = self.tree / 'src/compiler/good.ts'
        self.records[2].update(where=f'{self.stop}:7:9', reason=self.reason)
        entry = self.tree / 'src/tsc/tsc.ts'
        entry.parent.mkdir(parents=True)
        entry.write_text('export {};\n')
        directory = self.run / 'tsc'
        write_entry(self.tree, directory, sources=[entry, self.stop, self.root])
        path = directory / 'latent.jsonl'
        rows = [json.loads(line) for line in path.read_text().splitlines()]
        rows[3]['units'] = [{'status': 'attempted'}]
        rows[2]['units'] = [{'status': 'attempted' if attempted else 'skipped_checker_body'}]
        if entry_reason:
            # Attribute by finding location even when a different file's unit found it.
            rows[1]['findings'] = [{'measurement': ENTRY_MEASUREMENT, 'kind': 'NotYet',
                                    'where': f'{self.stop}:99:1', 'reason': entry_reason, 'text': 'stop'}]
        path.write_text(''.join(json.dumps(row) + '\n' for row in rows))
        self.node = {'program_files': [str(self.root), str(self.stop)], 'program_files_complete': True,
                     'node_version': 'v24.19.0', 'build_exit': 0, 'exit': 1 if node_throws else 0,
                     'timeout': False, 'throws': node_throws,
                     'first_error': "ReferenceError: Cannot access 'early' before initialization" if node_throws else ''}
        self.probe = patch('report.run_roots', return_value={str(self.root): self.node})
        self.mock = self.probe.start()
        self.addCleanup(self.probe.stop)

    def root_row(self, result):
        return next(row for row in result['files'] if row['file'] == 'src/compiler/later.ts')

    def test_entry_order_readiness_still_counts_against_root_file(self):
        from report import READINESS_REASONS
        self.readiness(entry_reason=sorted(READINESS_REASONS)[1])
        result = self.render()
        self.assertFalse(self.root_row(result)['lowering'])
        self.assertEqual(result['totals']['lowering'], 1)
        self.assertEqual(result['root_order_artifacts'], [])
        self.mock.assert_called_once()
        self.assertEqual(self.mock.call_args.args[1], [])

    def test_root_only_readiness_with_node_throw_is_visible_artifact(self):
        self.readiness()
        result = self.render()
        self.assertTrue(self.root_row(result)['lowering'])
        self.assertEqual(result['totals']['lowering'], 2)
        artifact = {'root': 'src/compiler/later.ts', 'where': 'src/compiler/good.ts:7:9',
                    'reason': self.reason, 'node': self.node,
                    'entry_program_attempted': True, 'entry_unattempted_sources': [], 'entry_next_stop': None,
                    'entry_whole_program_passes_measured': False}
        self.assertEqual(result['root_order_artifacts'], [artifact])
        self.assertEqual(result['readiness_node_checks'], [artifact])
        text = (self.run / 'report.md').read_text()
        self.assertIn('root-order artifacts: 1', text)
        self.assertIn('| src/compiler/later.ts | src/compiler/good.ts:7:9 | ' + self.reason, text)
        self.assertIn('Node throws here as well; exit 1; ReferenceError:', text)
        self.assertIn('| src/compiler/later.ts | pass | pass | pass |', text)
        self.assertEqual(self.mock.call_args.args[1], [str(self.root)])

    def test_root_only_readiness_with_clean_node_still_counts(self):
        from report import READINESS_REASONS
        self.readiness(node_throws=False)
        self.records[2]['reason'] = sorted(READINESS_REASONS)[1]
        result = self.render()
        self.assertFalse(self.root_row(result)['lowering'])
        self.assertEqual(result['totals']['lowering'], 1)
        self.assertEqual(result['root_order_artifacts'], [])
        self.assertEqual(result['readiness_node_checks'][0]['node']['exit'], 0)
        self.assertIn('Node runs clean; exit 0;', (self.run / 'report.md').read_text())

    def test_unattempted_file_is_not_evidence(self):
        self.readiness(attempted=False)
        result = self.render()
        self.assertFalse(self.root_row(result)['lowering'])
        self.assertEqual(result['readiness_node_checks'], [])
        self.assertEqual(self.mock.call_args.args[1], [])

    def test_other_reason_never_gets_an_artifact(self):
        self.readiness()
        self.records[2]['reason'] = 'reading unrelated'
        result = self.render()
        self.assertFalse(self.root_row(result)['lowering'])
        self.assertEqual(result['readiness_node_checks'], [])
        self.assertEqual(self.mock.call_args.args[1], [])

    def test_inconclusive_node_still_counts(self):
        self.readiness()
        self.node.update(throws=False, exit=None, timeout=True, first_error='')
        result = self.render()
        self.assertFalse(self.root_row(result)['lowering'])
        self.assertEqual(result['root_order_artifacts'], [])
        self.assertIn('Node probe inconclusive', (self.run / 'report.md').read_text())

    def test_daily_table_keeps_artifacts_beside_counts(self):
        from shutil import copytree
        self.readiness()
        self.render()
        for label in ('main', 'area'):
            directory = self.run / label
            directory.mkdir()
            (directory / 'census.jsonl').write_bytes((self.run / 'census.jsonl').read_bytes())
            copytree(self.run / 'tsc', directory / 'tsc')
            latent = [{'measurement': MEASUREMENT, 'status': 'measurement', 'checker_rejected': True}]
            latent += [{'measurement': MEASUREMENT, 'file': str(path), 'findings': []}
                       for path in sorted((self.tree / 'src/compiler').glob('*.ts'))]
            (directory / 'latent.jsonl').write_text(''.join(json.dumps(row) + '\n' for row in latent))
        result = report_pair(self.tree, self.tree, self.run, 'stamp', 'main-sha', 'main-sha', self.run, '1' * 40)
        text = (self.run / 'report.md').read_text()
        for label in ('main', 'area'):
            self.assertEqual(len(result['trees'][label]['root_order_artifacts']), 1)
            self.assertIn(f'Body lowering, {label}: 2/4; root-order artifacts: 1.', text)
        self.assertEqual(text.count('Node throws here as well; exit 1; ReferenceError:'), 2)
        self.assertIn('Node throws here as well; exit 1; ReferenceError:', (self.run / 'README.md').read_text())
        disk = json.loads((self.run / 'report.json').read_text())
        self.assertEqual(disk['trees']['main']['root_order_artifacts'], result['root_order_artifacts'])

    def test_enum_readiness_with_node_throw_is_also_an_artifact(self):
        from report import READINESS_REASONS
        self.readiness()
        self.records[2]['reason'] = sorted(READINESS_REASONS)[1]
        result = self.render()
        self.assertTrue(self.root_row(result)['lowering'])
        self.assertEqual(result['root_order_artifacts'][0]['reason'], sorted(READINESS_REASONS)[1])

    def test_artifact_tail_stop_preserves_body_count_loom03dg(self):
        from report import ENTRY_MEASUREMENT
        self.readiness()
        path = self.run / 'tsc/latent.jsonl'
        rows = [json.loads(line) for line in path.read_text().splitlines()]
        rows[1]['findings'] = [dict(kind='NotYet', where=f'{self.root}:167:31',
                                    unit=f'{self.root}:1:1', phase='tail',
                                    **{'pass': 'checkRecordGuardedReads'},
                                    reason='guarded record', text='guarded record',
                                    measurement=ENTRY_MEASUREMENT)]
        path.write_text(''.join(json.dumps(row) + '\n' for row in rows))
        result = self.render()
        self.assertTrue(self.root_row(result)['lowering'])
        self.assertEqual(result['totals']['lowering'], 2)
        artifact = result['root_order_artifacts'][0]
        self.assertIsNone(artifact['entry_next_stop'])
        self.assertEqual(artifact['entry_tail_stops'][0]['pass_name'], 'checkRecordGuardedReads')
        self.assertFalse(artifact['entry_whole_program_passes_measured'])
        self.assertIn('scoped tail stop checkRecordGuardedReads: guarded record',
                      (self.run / 'report.md').read_text())

    def test_artifact_with_entry_refusal_does_not_lower_whole(self):
        from report import ENTRY_MEASUREMENT
        self.readiness()
        path = self.run / 'tsc/latent.jsonl'
        rows = [json.loads(line) for line in path.read_text().splitlines()]
        # A finding in this file counts even when another file's unit encountered it.
        rows[1]['findings'] = [{'kind': 'Refused', 'where': f'{self.root}:167:31',
                                'reason': 'a cast the runtime cannot check',
                                'text': 'refused', 'measurement': ENTRY_MEASUREMENT}]
        path.write_text(''.join(json.dumps(row) + '\n' for row in rows))
        result = self.render()
        self.assertFalse(self.root_row(result)['lowering'])
        self.assertEqual(result['totals']['lowering'], 1)
        self.assertEqual(len(result['root_order_artifacts']), 1)
        stop = {'kind': 'Refused', 'where': 'src/compiler/later.ts:167:31',
                'reason': 'a cast the runtime cannot check', 'source': 'entry census', 'order': 'source order'}
        self.assertEqual(self.root_row(result)['lowering_stop'], stop)
        self.assertEqual(self.root_row(result)['lowering_evidence'], 'entry census')
        self.assertEqual(result['root_order_artifacts'][0]['entry_next_stop'], stop)
        text = (self.run / 'report.md').read_text()
        self.assertIn('Refused src/compiler/later.ts:167:31: a cast the runtime cannot check (read from the entry census; source order)', text)
        self.assertIn('root-order artifact', text)

    def test_program_relation_stop_prevents_artifact_lowering_credit(self):
        from report import ENTRY_MEASUREMENT
        self.readiness()
        path = self.run / 'tsc/latent.jsonl'
        rows = [json.loads(line) for line in path.read_text().splitlines()]
        rows.append(dict(status='relation_report', measurement=ENTRY_MEASUREMENT,
                         relation_overflow=dict(after_load='', end='overflow'),
                         findings=[dict(measurement=ENTRY_MEASUREMENT, kind='NotYet', where='program',
                                        reason='relation limit', text='relation limit')]))
        path.write_text(''.join(json.dumps(row) + '\n' for row in rows))
        result = self.render()
        self.assertFalse(self.root_row(result)['lowering'])
        self.assertEqual(self.root_row(result)['lowering_stop']['where'], 'program')

    def test_artifact_without_attempted_root_units_does_not_lower_whole(self):
        self.readiness()
        path = self.run / 'tsc/latent.jsonl'
        rows = [json.loads(line) for line in path.read_text().splitlines()]
        rows[3]['units'] = []
        path.write_text(''.join(json.dumps(row) + '\n' for row in rows))
        result = self.render()
        self.assertFalse(self.root_row(result)['lowering'])
        self.assertEqual(result['totals']['lowering'], 1)
        self.assertEqual(len(result['root_order_artifacts']), 1)
        self.assertIn('program reach unavailable or units not attempted; lowering unproven', (self.run / 'report.md').read_text())

    def test_artifact_with_clean_own_file_and_refused_dependency_does_not_lower_whole(self):
        from report import ENTRY_MEASUREMENT
        self.readiness()
        path = self.run / 'tsc/latent.jsonl'
        rows = [json.loads(line) for line in path.read_text().splitlines()]
        rows[2]['findings'] = [{'kind': 'Refused', 'where': f'{self.stop}:356:13',
                                'reason': 'refused dependency', 'text': 'refused',
                                'measurement': ENTRY_MEASUREMENT}]
        path.write_text(''.join(json.dumps(row) + '\n' for row in rows))
        result = self.render()
        self.assertFalse(self.root_row(result)['lowering'])
        self.assertEqual(result['totals']['lowering'], 1)
        self.assertEqual(len(result['root_order_artifacts']), 1)
        stop = self.root_row(result)['lowering_stop']
        self.assertEqual(stop, {'kind': 'Refused', 'where': 'src/compiler/good.ts:356:13',
                                'reason': 'refused dependency', 'source': 'entry census', 'order': 'source order'})
        self.assertIn('refused dependency (read from the entry census; source order)',
                      (self.run / 'report.md').read_text())

    def test_artifact_needs_attempted_units_in_every_program_source(self):
        from report import ENTRY_MEASUREMENT
        self.readiness()
        dependency = self.tree / 'src/compiler/refused.ts'
        self.node['program_files'].append(str(dependency))
        path = self.run / 'tsc/latent.jsonl'
        rows = [json.loads(line) for line in path.read_text().splitlines()]
        rows[0]['sources'].append(str(dependency))
        rows.append({'file': str(dependency), 'measurement': ENTRY_MEASUREMENT, 'units': [], 'findings': []})
        path.write_text(''.join(json.dumps(row) + '\n' for row in rows))
        result = self.render()
        self.assertFalse(self.root_row(result)['lowering'])
        self.assertEqual(result['totals']['lowering'], 1)
        self.assertEqual(result['root_order_artifacts'][0]['entry_unattempted_sources'],
                         ['src/compiler/refused.ts'])

class RelationTrailerReportTests(unittest.TestCase):
    def summary(self, trailer=True, findings=None):
        from report import lowering_summary
        rows = [{'status': 'measurement', 'measurement': MEASUREMENT, 'checker_rejected': True},
                {'file': '/only.a', 'measurement': MEASUREMENT, 'findings': []}]
        if trailer:
            rows.append({'status': 'relation_report', 'measurement': MEASUREMENT,
                         'relation_overflow': {'after_load': '', 'end': 'planted overflow'},
                         'findings': findings or []})
        return lowering_summary(rows, {'/only.a'}, MEASUREMENT)

    def test_relation_trailer_is_not_a_source_file(self):
        summary = self.summary()
        self.assertEqual(summary['source_files'], 1)
        self.assertTrue(summary['relation_report_complete'])
        self.assertEqual(summary['relation_overflow']['end'], 'planted overflow')
        self.assertNotIn('run did not complete', '\n'.join(latent_table('main', summary)))

    def test_relation_trailer_findings_count_as_notyet_sites(self):
        reason = 'a type relation the checker gave up on as too complex, whose answer lowering read (planted overflow)'
        finding = dict(measurement=MEASUREMENT, kind='NotYet', where='program', reason=reason, text=reason)
        summary = self.summary(findings=[finding, finding])
        self.assertEqual(summary['totals']['NotYet'], 1)
        self.assertEqual(summary['recorded_events'], 2)
        self.assertEqual(summary['per_reason'], {'NotYet: ' + reason: 1})
        self.assertIn(reason, '\n'.join(latent_table('main', summary)))

    def test_missing_relation_trailer_reports_incomplete_run(self):
        summary = self.summary(trailer=False)
        self.assertFalse(summary['relation_report_complete'])
        self.assertIsNone(summary['relation_overflow'])
        self.assertIn('run did not complete; end answer unavailable', '\n'.join(latent_table('main', summary)))

    def test_misplaced_duplicate_or_unfinished_relation_trailer_is_rejected(self):
        from report import lowering_summary
        header = dict(status='measurement', measurement=MEASUREMENT, checker_rejected=True)
        file = dict(file='/only.a', measurement=MEASUREMENT, findings=[])
        trailer = dict(status='relation_report', measurement=MEASUREMENT,
                       relation_overflow=dict(after_load='', end=''), findings=[])
        for rows in ([header, trailer, file], [header, file, trailer, trailer],
                     [header, file, dict(trailer, relation_overflow=dict(after_load=''))]):
            with self.subTest(rows=rows), self.assertRaises(ValueError):
                lowering_summary(rows, {'/only.a'}, MEASUREMENT)


class RelationBudgetTests(unittest.TestCase):
    def verdict(self, trailer=True, **relation):
        from report import lowering_summary, relation_budget
        rows = [{'status': 'measurement', 'measurement': MEASUREMENT, 'checker_rejected': False},
                {'file': '/only.a', 'measurement': MEASUREMENT, 'findings': []}]
        if trailer:
            rows.append({'status': 'relation_report', 'measurement': MEASUREMENT, 'findings': [],
                         'relation_overflow': dict({'after_load': '', 'end': ''}, **relation)})
        return relation_budget({'lowering_census': lowering_summary(rows, {'/only.a'}, MEASUREMENT)})

    def sizes(self, assignable, **others):
        from report import RELATION_CACHES
        return dict(dict.fromkeys(RELATION_CACHES, 0), assignable=assignable, **others)

    def test_measured_head_is_ok(self):
        self.assertEqual(self.verdict(cache_sizes=self.sizes(299_624)),
                         {'status': 'ok', 'assignable': 299_624, 'threshold': 1_000_000, 'overflow_marked': False,
                          'why': 'assignable cache 299,624 pairs, within 1,000,000'})

    def test_threshold_is_ok_and_one_past_it_is_red(self):
        self.assertEqual(self.verdict(cache_sizes=self.sizes(1_000_000))['status'], 'ok')
        verdict = self.verdict(cache_sizes=self.sizes(1_000_001))
        self.assertEqual((verdict['status'], verdict['assignable']), ('red', 1_000_001))
        self.assertEqual(verdict['why'], 'assignable cache 1,000,001 pairs, past 1,000,000')

    def test_overflow_mark_at_the_end_is_red_under_the_threshold(self):
        verdict = self.verdict(end='a type relation the checker marked as overflowed, whose diagnostic it dropped',
                               cache_sizes=self.sizes(299_624))
        self.assertEqual((verdict['status'], verdict['overflow_marked'], verdict['assignable']), ('red', True, 299_624))
        self.assertIn('marked as overflowed', verdict['why'])

    def test_one_flagged_pair_or_ts2859_is_red_under_the_threshold(self):
        # The end answer is left empty here so only the counts can turn it red (#ferqs8v).
        for field, why in (('flaggedPairs', '1 pairs flagged as overflowed'), ('ts2859', '1 TS2859 diagnostics')):
            with self.subTest(field=field):
                verdict = self.verdict(cache_sizes=self.sizes(299_624, **{field: 1}))
                self.assertEqual((verdict['status'], verdict['why']), ('red', why))

    def test_unreadable_cache_is_red(self):
        for sizes in (self.sizes(299_624, enum=-1), self.sizes(299_624, flaggedPairs=-1), self.sizes(299_624, ts2859=-1),
                      self.sizes(-1), {'assignable': 299_624}):
            with self.subTest(sizes=sizes):
                verdict = self.verdict(cache_sizes=sizes)
                self.assertEqual(verdict['status'], 'red')
                self.assertIn('unreadable relation caches', verdict['why'])

    def test_missing_trailer_is_unknown(self):
        verdict = self.verdict(trailer=False)
        self.assertEqual((verdict['status'], verdict['assignable']), ('unknown', None))
        self.assertIn('did not complete', verdict['why'])

    def test_trailer_without_sizes_is_unknown(self):
        self.assertEqual(self.verdict()['status'], 'unknown')
        self.assertEqual(self.verdict(cache_sizes={})['status'], 'unknown')

    def test_blocked_entry_is_unknown(self):
        from report import relation_budget
        self.assertEqual(relation_budget({'lowering_census': None})['status'], 'unknown')


class PassLedgerReportTestsPb47(unittest.TestCase):
    def ledger(self, status='not_run_scoped'):
        from pass_ledger import required_steps
        return [{'name': name, 'status': status} for name in required_steps()]

    def observation(self):
        from report import MEASUREMENT, lowering_summary
        units = [dict(where='input.ts:1:1', status='attempted', pass_ledger=self.ledger()),
                 dict(where='input.ts:2:1', status='attempted', pass_ledger=self.ledger('not_reached_lowering_stop')),
                 dict(where='input.ts:3:1', status='attempted'),
                 dict(where='input.ts:4:1', status='deferred_generic_instantiation', instantiated=True),
                 dict(where='input.ts:5:1', status='skipped_checker_body')]
        rows = [dict(status='measurement', measurement=MEASUREMENT, checker_rejected=False),
                dict(file='input.ts', measurement=MEASUREMENT, units=units, findings=[])]
        return lowering_summary(rows, {str(Path('input.ts').resolve())}, MEASUREMENT)

    def test_body_only_pass_ledger_report_line_pb47(self):
        from report import latent_table, entry_table
        summary = self.observation()
        line = '2 units lower without a body stop; 0 have their whole-program passes measured'
        self.assertIn(line, latent_table('main', summary))
        self.assertIn(line, entry_table('main', {'lowering_census':summary}))
        self.assertEqual(summary['totals']['NotYet'], 0)

    def test_omission_as_execution_report_mutant_is_red_pb47(self):
        from report import latent_table
        # A reporter that treats unmeasured passes as execution changes the new
        # line even though the old finding totals remain zero.
        with patch('pass_ledger.whole_passes_measured', return_value=True):
            mutant = self.observation()
        self.assertNotIn('2 units lower without a body stop; 0 have their whole-program passes measured',
                         latent_table('main', mutant))
        self.assertEqual(mutant['whole_program_passes_measured_units'], 2)

    def test_whole_coverage_requires_complete_executed_vector_pb47(self):
        from pass_ledger import whole_passes_measured
        unit = dict(status='attempted', pass_ledger=self.ledger('executed'))
        self.assertTrue(whole_passes_measured(unit))
        for ledger in (None, [], unit['pass_ledger'][:-1], self.ledger(),
                       [dict(unit['pass_ledger'][0])]*len(unit['pass_ledger'])):
            self.assertFalse(whole_passes_measured(dict(status='attempted', pass_ledger=ledger)))
        self.assertFalse(whole_passes_measured(dict(status='deferred_generic_instantiation',
                                                  pass_ledger=self.ledger('executed'))))


class ScopedTailReportTestsLoom03dg(unittest.TestCase):
    def test_tail_stop_does_not_change_body_totals_loom03dg(self):
        from report import MEASUREMENT, lowering_summary, latent_table, entry_table
        rows = [dict(status='measurement', measurement=MEASUREMENT, checker_rejected=False),
                dict(file='tail.ts', measurement=MEASUREMENT,
                     units=[dict(where='tail.ts:1:1', status='attempted')],
                     findings=[dict(measurement=MEASUREMENT, unit='tail.ts:1:1',
                                    phase='tail', **{'pass':'checkRecordGuardedReads'},
                                    kind='NotYet', where='tail.ts:1:8', text='guarded record', reason='guarded record')])]
        summary = lowering_summary(rows, {str(Path('tail.ts').resolve())}, MEASUREMENT)
        self.assertEqual(summary['totals']['NotYet'], 0)
        self.assertEqual(summary['body_lowered_units'], 1)
        self.assertEqual(summary['whole_program_passes_measured_units'], 0)
        line = 'Scoped tail: 1 units stopped (checkRecordGuardedReads: 1).'
        self.assertIn(line, latent_table('main', summary))
        self.assertIn(line, entry_table('main', {'lowering_census':summary}))

    def test_unrun_tail_is_not_reported_as_measured_zero_loom03dg(self):
        from pass_ledger import tail_line
        self.assertEqual(tail_line({'scoped_tail_stops': {}}), 'Scoped tail: not run.')
        self.assertEqual(tail_line({'scoped_tail_stops': {}, 'scoped_tail_pass_executions': 13}),
                         'Scoped tail: 0 units stopped.')
