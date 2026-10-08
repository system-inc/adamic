"""Check the entry's imported reach, checker boundary, and independent accounting."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

from entry_overlay import prepare
from report import entry_summary, ENTRY_MEASUREMENT


def write_entry(tree, run, diagnostics=(), sources=None, findings=None):
    run.mkdir(parents=True, exist_ok=True)
    entry = (tree / 'src/tsc/tsc.ts').resolve()
    row = {'roots': [str(entry)], 'kind': 'checker' if diagnostics else 'accepted',
           'diagnostics': list(diagnostics)}
    (run / 'census.jsonl').write_text((json.dumps(row) + '\n') * 2)
    header = {'roots': [str(entry)], 'status': 'blocked' if diagnostics else 'measurement',
              'diagnostics': list(diagnostics), 'checker_rejected': bool(diagnostics)}
    records = [header]
    if not diagnostics:
        sources = sources or [entry]
        header.update(measurement=ENTRY_MEASUREMENT, sources=[str(p) for p in sources])
        records += [{'file': str(p), 'measurement': ENTRY_MEASUREMENT,
                     'findings': findings or []} for p in sources]
    (run / 'latent.jsonl').write_text(''.join(json.dumps(row) + '\n' for row in records))


class EntryReportTests(unittest.TestCase):
    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory()
        self.addCleanup(self.scratch.cleanup)
        self.tree = Path(self.scratch.name) / 'tree'
        self.run = Path(self.scratch.name) / 'run'
        self.entry = self.tree / 'src/tsc/tsc.ts'
        self.entry.parent.mkdir(parents=True)
        self.entry.write_text('export {};\n')
        self.dependency = self.tree / 'src/executeCommandLine.a'
        self.dependency.write_text('export {};\n')

    def test_imported_and_global_diagnostics_block_lowering(self):
        diagnostics = [f'{self.dependency}:3:7: error TS2322: imported mismatch',
                       'error TS2318: missing global']
        write_entry(self.tree, self.run, diagnostics=diagnostics)
        result = entry_summary(self.tree, self.run)
        self.assertFalse(result['checker_whole_program'])
        self.assertEqual(result['whole_program_diagnostics'], 2)
        self.assertEqual(result['diagnostics'], diagnostics)
        self.assertFalse(result['lowering_attempted'])
        self.assertIsNone(result['lowering_census'])

    def test_clean_reach_ranks_unique_sites_with_owners(self):
        finding = {'kind': 'NotYet', 'where': f'{self.dependency}:2:1',
                   'reason': '||=', 'text': 'not yet', 'measurement': ENTRY_MEASUREMENT}
        write_entry(self.tree, self.run, sources=[self.entry, self.dependency], findings=[finding, finding])
        result = entry_summary(self.tree, self.run)
        self.assertTrue(result['checker_whole_program'])
        self.assertEqual(result['whole_program_diagnostics'], 0)
        self.assertEqual(result['lowering_census']['source_files'], 2)
        self.assertEqual(result['lowering_census']['totals']['NotYet'], 1)
        self.assertEqual(result['lowering_census']['top_reasons'][0]['owner'], '01a113e3-a058')

    def test_equivalent_symlink_entry_roots_are_accepted(self):
        alias = Path(self.scratch.name) / 'alias'
        alias.symlink_to(self.tree, target_is_directory=True)
        for diagnostics in ([], ['error TS2318: missing global']):
            with self.subTest(diagnostics=diagnostics):
                write_entry(self.tree, self.run, diagnostics=diagnostics,
                            sources=[self.entry, self.dependency])
                for name in ('census.jsonl', 'latent.jsonl'):
                    path = self.run / name
                    rows = [json.loads(line) for line in path.read_text().splitlines()]
                    for row in rows:
                        if 'roots' in row:
                            row['roots'] = [str(alias / 'src/tsc/tsc.ts')]
                    path.write_text(''.join(json.dumps(row) + '\n' for row in rows))
                result = entry_summary(alias, self.run)
                self.assertEqual(result['checker_whole_program'], not diagnostics)
                if not diagnostics:
                    self.assertEqual(result['lowering_census']['source_files'], 2)

    def test_wrong_root_is_rejected(self):
        write_entry(self.tree, self.run)
        path = self.run / 'census.jsonl'
        rows = [json.loads(line) for line in path.read_text().splitlines()]
        for row in rows:
            row['roots'] = [str(self.dependency)]
        path.write_text(''.join(json.dumps(row) + '\n' for row in rows))
        with self.assertRaisesRegex(ValueError, 'roots or coverage'):
            entry_summary(self.tree, self.run)

    def test_missing_reachable_file_record_is_rejected(self):
        write_entry(self.tree, self.run, sources=[self.entry, self.dependency])
        path = self.run / 'latent.jsonl'
        path.write_text('\n'.join(path.read_text().splitlines()[:-1]) + '\n')
        with self.assertRaisesRegex(ValueError, 'coverage incomplete'):
            entry_summary(self.tree, self.run)

    def test_blocked_stream_cannot_claim_lowering(self):
        write_entry(self.tree, self.run, diagnostics=['error TS2318: missing global'])
        with (self.run / 'latent.jsonl').open('a') as output:
            output.write(json.dumps({'file': str(self.entry), 'findings': []}) + '\n')
        with self.assertRaisesRegex(ValueError, 'ran despite checker diagnostics'):
            entry_summary(self.tree, self.run)

    def test_missing_entry_is_rejected(self):
        write_entry(self.tree, self.run)
        self.entry.unlink()
        with self.assertRaisesRegex(ValueError, 'tsc entry missing'):
            entry_summary(self.tree, self.run)

    def test_checker_and_lowering_diagnostics_must_agree(self):
        write_entry(self.tree, self.run, diagnostics=['error TS2318: missing global'])
        path = self.run / 'latent.jsonl'
        header = json.loads(path.read_text())
        header['diagnostics'] = []
        path.write_text(json.dumps(header) + '\n')
        with self.assertRaisesRegex(ValueError, 'diagnostics disagree'):
            entry_summary(self.tree, self.run)

    def test_clean_reach_cannot_carry_rejected_program_label(self):
        write_entry(self.tree, self.run)
        path = self.run / 'latent.jsonl'
        path.write_text(path.read_text().replace(ENTRY_MEASUREMENT, 'measured on a checker-rejected program'))
        with self.assertRaisesRegex(ValueError, 'invalid latent measurement header'):
            entry_summary(self.tree, self.run)

    def test_overlay_missing_hook_fails_loudly(self):
        source = self.run / 'source'
        source.mkdir(parents=True)
        (source / 'overlay.json').write_text(json.dumps({'Replace': {}}))
        with self.assertRaisesRegex(ValueError, 'internal/load/latent_hook.go, found 0'):
            prepare(source, self.run / 'overlay')


@unittest.skipUnless(os.environ.get('ENTRY_CENSUS_BINARY') and os.environ.get('CENSUS_BINARY'),
                     'set ENTRY_CENSUS_BINARY and CENSUS_BINARY for the real entry reach probe')
class EntryCensusTests(unittest.TestCase):
    def test_reach_includes_dependency_outside_compiler_and_excludes_unused_file(self):
        with tempfile.TemporaryDirectory() as scratch:
            tree = Path(scratch) / 'tree'
            entry = tree / 'src/tsc/tsc.ts'
            entry.parent.mkdir(parents=True)
            dependency = tree / 'src/executeCommandLine.a'
            dependency.write_text('export function target(): number { debugger; return 1; }\n')
            unused = tree / 'src/unused.a'
            unused.write_text('export const value: number = "unused error";\n')
            entry.write_text('import { target } from "../executeCommandLine.a"; target();\n')
            observations = []
            for name in ['clean', 'planted']:
                run = Path(scratch) / name
                run.mkdir()
                if name == 'planted':
                    dependency.write_text('export function target(): number { return "wrong"; }\n')
                with (run / 'census.log').open('w') as log:
                    subprocess.run([os.environ['CENSUS_BINARY'], str(entry), str(run / 'census.jsonl')],
                                   check=True, stdout=log, stderr=log)
                with (run / 'latent.log').open('w') as log:
                    subprocess.run([os.environ['ENTRY_CENSUS_BINARY'], str(entry), str(run / 'latent.jsonl')],
                                   check=True, env=dict(os.environ, LATENT_ASSERT_NO_OUTPUT='1'), stdout=log, stderr=log)
                observations.append(entry_summary(tree, run))
            clean, planted = observations
            self.assertTrue(clean['checker_whole_program'])
            self.assertEqual(clean['lowering_census']['source_files'], 2)
            self.assertGreater(clean['lowering_census']['totals']['NotYet'], 0)
            self.assertFalse(planted['checker_whole_program'])
            self.assertEqual(planted['whole_program_diagnostics'], 1)
            self.assertFalse(planted['lowering_attempted'])
            self.assertIsNone(planted['lowering_census'])


@unittest.skipUnless(os.environ.get('ENTRY_CENSUS_BINARY') and os.environ.get('CENSUS_BINARY'),
                     'set ENTRY_CENSUS_BINARY and CENSUS_BINARY for full nested entry measurement')
class FullEntryCensusTests(unittest.TestCase):
    def test_rejected_entry_measures_clean_nested_body_and_catches_first_error_mutant(self):
        with tempfile.TemporaryDirectory() as scratch:
            tree = Path(scratch) / 'tree'
            entry = tree / 'src/tsc/tsc.ts'
            entry.parent.mkdir(parents=True)
            dependency = tree / 'src/executeCommandLine.a'
            dependency.write_text('''export function root(text: string, count: number): void {
 const impossible: number = "wrong";
 function safe(): void { var x = 1; if (text) { } if (count) { } }
}
''')
            entry.write_text('import { root } from "../executeCommandLine.a"; root("x", 1);\n')
            observations = []
            for name, extra in [('full', {}), ('mutant', {'LATENT_MUTANT_FIRST_ERROR_ONLY': '1'})]:
                run = Path(scratch) / name
                run.mkdir()
                with (run / 'census.log').open('w') as log:
                    subprocess.run([os.environ['CENSUS_BINARY'], str(entry), str(run / 'census.jsonl')],
                                   check=True, stdout=log, stderr=log)
                with (run / 'latent.log').open('w') as log:
                    subprocess.run([os.environ['ENTRY_CENSUS_BINARY'], str(entry), str(run / 'latent.jsonl')],
                                   check=True, env=dict(os.environ, LATENT_FULL='1', LATENT_ASSERT_NO_OUTPUT='1', **extra), stdout=log, stderr=log)
                observations.append(entry_summary(tree, run))
            full, mutant = observations
            self.assertFalse(full['checker_whole_program'])
            self.assertEqual(full['whole_program_diagnostics'], 1)
            self.assertFalse(full['lowering_attempted'])
            self.assertTrue(full['measurement_attempted'])
            self.assertEqual(full['lowering_census']['source_files'], 2)
            self.assertEqual(full['lowering_census']['measurement'], 'measured on a checker-rejected entry-root program')
            self.assertEqual(full['lowering_census']['totals']['Refused'], 3)
            with self.assertRaises(AssertionError):
                self.assertEqual(mutant['lowering_census']['totals']['Refused'], 3)


if __name__ == '__main__':
    unittest.main()
