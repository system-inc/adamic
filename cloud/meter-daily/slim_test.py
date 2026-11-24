#!/usr/bin/env python3
"""slim.py commits a record's summary and reason rows alone, names every file it leaves out by hash, and refuses bulk.

Mutants, each red here:
  slim.py copying `*.log.gz` through, or the whole run: test_record_holds_exactly_the_kept_files, on its file set;
  slim.py dropping reason_rows or per_reason from KEPT: test_record_holds_exactly_the_kept_files, on report.json;
  slim.py naming the slim report.json in kept.json, not the full one: test_record_holds_exactly_the_kept_files, on
    report.json's sha256;
  slim.py without its 1 MB check: test_a_record_past_a_megabyte_is_refused;
  slim.py leaving --job out of kept.json: test_job_lands_in_kept_json.
"""
import gzip
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

SLIM = Path(os.environ.get('METER_SLIM_UNDER_TEST', Path(__file__).resolve().parent / 'slim.py'))
# What the committed record holds, and nothing else.
RECORD = {'README.md', 'report.md', 'main/report.md', 'area/report.md', 'report.json', 'kept.json'}


def slim(*arguments):
    # -I, as submit.sh runs it.
    return subprocess.run([sys.executable, '-I', str(SLIM), *arguments], capture_output=True, text=True, timeout=60)


def rows(count, width=40):
    return [{'reason': f'reason {index} ' + 'x' * width, 'owner': '#jobs', 'NotYet': index, 'Refused': 1, 'count': index + 1}
            for index in range(count)]


def summary(reason_rows):
    """A lowering summary as report.py's lowering_summary writes it, its bulk beside what readers read."""
    return {
        'measurement': 'measured on a checker-rejected program',
        'relation_report_complete': True,
        'totals': {'NotYet': 7, 'Refused': 3, 'SkippedDependency': 0, 'error': 0, 'panic': 0},
        'per_reason': {f"NotYet: {row['reason']}": row['NotYet'] for row in reason_rows},
        'reason_rows': reason_rows,
        'unowned_reasons': [dict(row, owner='OWNER BLANK') for row in reason_rows],
        'top_reasons': reason_rows[:10],
        'panic_sites': [{'where': 'src/compiler/core.ts:1:1', 'reason': 'panic', 'text': 'lowering a call'}],
        'excluded_nested_functions': [],
    }


def full_report(reason_rows, lowering_census=True):
    main = {
        'tree_commit': 'a' * 40,
        'adamic_commit': 'b' * 40,
        'totals': {'files': 80, 'source_files': 79, 'checker': 2, 'lowering': 2, 'checker_whole_program': 2,
                   'checker_own_file': 56},
        'files': [{'file': 'src/compiler/core.ts', 'checker': False, 'lowering': False, 'outcome': 'checker'}],
        'readiness_node_checks': [{'root': 'src/compiler/debug.ts', 'node': {'exit': 1, 'throws': True}}],
        'latent_lowering': summary(reason_rows),
        'latent_time': {'wall_seconds': 12.5, 'cpu_seconds': 11.0, 'files': [], 'slowest_units': []},
        'tsc_entry': {'whole_program_diagnostics': 235, 'diagnostics': ['error TS2322: fixture'],
                      'lowering_census': summary(reason_rows) if lowering_census else None},
    }
    return dict(main, timestamp_utc='20261125T131040Z', unowned_reasons=main['latent_lowering']['unowned_reasons'],
                trees={'main': main, 'area': main}, box={'hostname': 'pupa', 'nproc': 32},
                relation_budget={'status': 'ok', 'assignable': 299624, 'threshold': 1000000, 'why': 'within'})


class Slim(unittest.TestCase):
    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory()
        self.addCleanup(self.scratch.cleanup)
        self.run_directory = Path(self.scratch.name) / '20261125T131040Z.fixture'
        self.out = Path(self.scratch.name) / 'out' / '20261125T131040Z.fixture'
        self.rows = rows(12)

    def write(self, name, data):
        path = self.run_directory / name
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(data if isinstance(data, bytes) else data.encode())

    def full_run(self, report):
        """A finished run as twice-daily.sh and job.sh leave it: Markdown, reports, logs and raw streams."""
        self.write('README.md', '# Stage 3 meter 20261125T131040Z\n')
        self.write('report.md', 'Whole program: main: 2/79; area: 2/79\nOwn file: main: 56/79; area: 56/79\n')
        self.write('report.json', json.dumps(report, indent=2) + '\n')
        for label in ('main', 'area'):
            self.write(f'{label}/report.md', f'{label} report\n')
            self.write(f'{label}/report.json', json.dumps(report['trees'][label], indent=2) + '\n')
            self.write(f'{label}/census.log.gz', gzip.compress(b'census output\n', mtime=0))
            self.write(f'{label}/latent-times.jsonl.gz', gzip.compress(b'{"file": "core.ts"}\n', mtime=0))
            self.write(f'{label}/node-readiness/debug/compiler-options.json', '{"strict": true}\n')
        self.write('job-meter.log', 'meter-job: measuring\n')
        self.write('upstream.log.gz', gzip.compress(b'upstream\n', mtime=0))
        self.write('toolchain.txt', 'node v24.11.0\n')
        self.write('box.json', '{"hostname": "pupa"}\n')

    def held(self, root):
        return {str(path.relative_to(root)): path.read_bytes() for path in root.rglob('*') if path.is_file()}

    def slim(self, *arguments):
        return slim(str(self.run_directory), str(self.out), *arguments)

    def test_record_holds_exactly_the_kept_files(self):
        report = full_report(self.rows)
        self.full_run(report)
        before = self.held(self.run_directory)
        result = self.slim()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.held(self.run_directory), before, 'slim.py wrote to the run')
        record = self.held(self.out)
        self.assertEqual(set(record), RECORD)
        for name in RECORD - {'report.json', 'kept.json'}:
            self.assertEqual(record[name], before[name], name)

        slimmed = json.loads(record['report.json'])
        main = slimmed['trees']['main']
        for census in (main['latent_lowering'], main['tsc_entry']['lowering_census']):
            self.assertEqual(census['reason_rows'], self.rows)
            self.assertEqual(census['totals'], report['trees']['main']['latent_lowering']['totals'])
            self.assertTrue(census['relation_report_complete'])
        self.assertEqual(main['latent_lowering']['per_reason'], report['trees']['main']['latent_lowering']['per_reason'])
        self.assertEqual(main['totals'], report['trees']['main']['totals'])
        self.assertEqual((main['tree_commit'], main['adamic_commit']), ('a' * 40, 'b' * 40))
        self.assertEqual(main['tsc_entry']['whole_program_diagnostics'], 235)
        self.assertEqual(slimmed['relation_budget'], report['relation_budget'])
        self.assertEqual(slimmed['latent_lowering'], {'totals': report['latent_lowering']['totals']})
        # The bulk stays with the job.
        self.assertEqual(set(slimmed['trees']), {'main'})
        for key in ('files', 'latent_time', 'readiness_node_checks'):
            self.assertNotIn(key, main)
        self.assertNotIn('unowned_reasons', slimmed)
        self.assertNotIn('unowned_reasons', main['latent_lowering'])

        kept = json.loads(record['kept.json'])
        self.assertIsNone(kept['job'])
        left = {name: data for name, data in before.items() if name not in RECORD - {'report.json', 'kept.json'}}
        self.assertEqual({entry['path'] for entry in kept['files']}, set(left))
        for entry in kept['files']:
            with self.subTest(path=entry['path']):
                self.assertEqual(entry['bytes'], len(left[entry['path']]))
                self.assertEqual(entry['sha256'], hashlib.sha256(left[entry['path']]).hexdigest())
        self.assertIn('main/census.log.gz', left)
        self.assertIn('job-meter.log', left)

    def test_a_blocked_entry_keeps_its_census_as_null(self):
        self.full_run(full_report(self.rows, lowering_census=False))
        result = self.slim()
        self.assertEqual(result.returncode, 0, result.stderr)
        main = json.loads((self.out / 'report.json').read_text())['trees']['main']
        self.assertIsNone(main['tsc_entry']['lowering_census'])

    def test_a_record_past_a_megabyte_is_refused(self):
        # Two summaries of 2,000 rows of 600 characters: the slim report.json alone is past 2 MB.
        self.full_run(full_report(rows(2000, width=600)))
        result = self.slim()
        self.assertEqual(result.returncode, 1, result.stdout)
        self.assertIn('report.json is', result.stderr)
        self.assertIn('past 1,048,576', result.stderr)
        self.assertFalse(self.out.exists(), 'a refused record left files behind')

    def test_job_lands_in_kept_json(self):
        self.full_run(full_report(self.rows))
        result = self.slim('--job', 'J-08xadbj')
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(json.loads((self.out / 'kept.json').read_text())['job'], 'J-08xadbj')

    def test_a_report_without_a_kept_key_is_refused(self):
        report = full_report(self.rows)
        del report['trees']['main']['latent_lowering']['reason_rows']
        self.full_run(report)
        result = self.slim()
        self.assertEqual(result.returncode, 1, result.stdout)
        self.assertIn('reason_rows', result.stderr)
        self.assertFalse(self.out.exists())

    def test_an_out_directory_already_there_is_refused(self):
        self.full_run(full_report(self.rows))
        self.out.mkdir(parents=True)
        result = self.slim()
        self.assertEqual(result.returncode, 1, result.stdout)
        self.assertEqual(list(self.out.iterdir()), [])

    def test_a_malformed_job_is_a_usage_error(self):
        self.full_run(full_report(self.rows))
        self.assertEqual(self.slim('--job', 'J-08X; rm').returncode, 2)
        self.assertFalse(self.out.exists())


if __name__ == '__main__':
    unittest.main()
