#!/usr/bin/env python3
"""Planted observations exercise the same checker CLI that the lane uses."""
import copy
import difflib
import os
import gzip
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

import check

LANE = Path(__file__).resolve().parent


class LandingLaneTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.base_scratch = tempfile.TemporaryDirectory(prefix='stage3-lane-api-fixture-')
        cls.addClassCleanup(cls.base_scratch.cleanup)
        cls.base = Path(cls.base_scratch.name) / 'git'
        cls.base.mkdir()
        cls.original = ("declare namespace ts {\n"
                        "    interface Host {\n"
                        "        optional?: number;\n"
                        "        method?(s: string): void;\n"
                        '        trace?: { trace(s: string): void }["trace"];\n'
                        "    }\n"
                        "    type Brand = { brand: any };\n"
                        "}\n")
        cls.adapted = cls.original.replace('optional?: number;', 'optional?: number | undefined;').replace(
            '["trace"];', '["trace"] | undefined;').replace('brand: any', 'brand: undefined')
        reference = cls.base / 'tests/baselines/reference/api/typescript.d.ts'
        reference.parent.mkdir(parents=True)
        reference.write_text(cls.original)
        for command in [['git', 'init', '-q', str(cls.base)],
                        ['git', '-C', str(cls.base), 'add', '.'],
                        ['git', '-C', str(cls.base), '-c', 'user.name=Lane test',
                         '-c', 'user.email=lane-test@example.invalid', 'commit', '-q', '-m', 'API fixture']]:
            subprocess.run(command, check=True, capture_output=True)
        pin = subprocess.check_output(['git', '-C', str(cls.base), 'rev-parse', 'HEAD'], text=True).strip()
        original = Path(cls.base_scratch.name) / 'original.d.ts'
        adapted = Path(cls.base_scratch.name) / 'adapted.d.ts'
        original.write_text(cls.original)
        adapted.write_text(cls.adapted)
        env = dict(os.environ)
        cache = Path(env.get('STAGE3_CACHE', str(Path.home() / '.cache/adamic-stage3')))
        env['NODE_PATH'] = str(cache / 'api/node_modules') + os.pathsep + env.get('NODE_PATH', '')
        changes = json.loads(subprocess.check_output(['node', str(LANE / 'normalize-api.cjs'),
                                                      str(original), str(adapted)], text=True, env=env))
        cls.fixture_sanction = dict(source_commit=pin, normalized_changes=changes)

    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory(prefix='stage3-lane-mutant-')
        self.addCleanup(self.scratch.cleanup)
        self.results = Path(self.scratch.name)
        (self.results / 'oracle').mkdir()
        subprocess.run(['git', 'clone', '-q', '--shared', str(self.base), str(self.results / 'adapted-tree')],
                       check=True, capture_output=True)
        self.expected = json.loads((LANE / 'expected.json').read_text())
        self.sanctioned_file = self.results / 'sanctioned.json'
        self.sanctioned_file.write_text(json.dumps(self.fixture_sanction))
        evidence = LANE.parent / 'adapt/75-optional-widening/evidence/full-apply'
        self.oracle = json.loads((evidence / 'oracle-report.json').read_text())
        # The evidence ran upstream's 4 workers, before the lane fixed eight.
        self.oracle['workers'] = 8
        self.log = gzip.decompress((evidence / 'oracle-tests.log.gz').read_bytes()).decode()
        # Counts and failure title come from the real upstream reporter.
        # API mutants use an isolated, valid declaration snapshot.
        self.reference = self.original
        self.local = self.adapted
        self.diff = None
        self.execution = dict(apply_exit=0, oracle_exit=1,
                              platform=dict(os='linux', arch='x64', node='v24.19.0'))

    def replay(self):
        (self.results / 'execution.json').write_text(json.dumps(self.execution))
        (self.results / 'oracle/report.json').write_text(json.dumps(self.oracle))
        (self.results / 'oracle/tests.log').write_text(self.log)
        tree = self.results / 'adapted-tree'
        (tree / 'tests/baselines/reference/api/typescript.d.ts').write_text(self.reference)
        local = tree / 'tests/baselines/local/api/typescript.d.ts'
        local.parent.mkdir(parents=True, exist_ok=True)
        local.write_text(self.local)
        diff = self.diff if self.diff is not None else ''.join(difflib.unified_diff(
            self.reference.splitlines(True), self.local.splitlines(True),
            fromfile='reference/api/typescript.d.ts', tofile='local/api/typescript.d.ts'))
        (self.results / 'oracle/baseline.diff').write_text(diff)
        result = subprocess.run([sys.executable, str(LANE / 'check.py'), str(self.results),
                                 '--sanctioned', str(self.sanctioned_file)], capture_output=True, text=True)
        report = json.loads((self.results / 'report.json').read_text())
        self.assertEqual(result.returncode, 0 if report['status'] == 'pass' else 1)
        self.assertEqual(result.stdout, report['verdict'] + '\n')
        return report

    def killed(self, label):
        report = self.replay()
        self.assertEqual(report['status'], 'fail')
        self.assertTrue(any(label in error for error in report['errors']), report)
        print(f'caught {self._testMethodName}: {label}')
        return report

    def test_sanctioned_observation_passes(self):
        self.assertEqual(self.replay()['status'], 'pass')
        self.assertEqual(check.failed_tests(self.log), self.expected['failed_tests'])

    def test_unsanctioned_line(self):
        self.local = self.local.replace('interface Host {', 'interface Host {\n        rogue: string;')
        report = self.killed('unsanctioned API declaration')
        self.assertEqual(len(report['errors']), 1)

    def test_missing_sanctioned_line(self):
        self.local = self.local.replace('optional?: number | undefined;', 'optional?: number;')
        report = self.killed('missing sanctioned API declaration')
        self.assertEqual(len(report['errors']), 1)

    def test_changed_sanctioned_type(self):
        self.local = self.local.replace('optional?: number | undefined;', 'optional?: string | undefined;')
        self.killed('changed sanctioned API declaration')

    def test_multiline_printer_form(self):
        self.local = self.local.replace('trace?: { trace(s: string): void }["trace"] | undefined;',
                                        "trace?:\n            | {\n                trace(s: string): void;\n"
                                        "            }['trace']\n            | undefined;")
        self.local = self.local.replace('optional?: number | undefined;', 'optional?: (number | undefined);')
        self.assertEqual(self.replay()['status'], 'pass')

    def test_proved_reference_changes_are_included(self):
        self.reference = self.reference.replace('brand: any', 'brand: undefined')
        report = self.replay()
        self.assertEqual(report['status'], 'pass')
        self.assertEqual(report['api']['reference_declarations'], 1)

    def test_unproved_reference_change(self):
        self.reference = self.reference.replace('optional?: number;', 'optional?: boolean;')
        self.killed('unsanctioned API reference declaration')

    def test_extra_comment_line(self):
        self.local += '// Unreviewed API documentation\n'
        self.killed('unsanctioned API declaration: comments')

    def test_source_pin(self):
        sanctioned = json.loads(self.sanctioned_file.read_text())
        sanctioned['source_commit'] = '0' * 40
        self.sanctioned_file.write_text(json.dumps(sanctioned))
        self.killed('API source pin')

    def test_failed_install(self):
        self.oracle['phases']['install']['exit'] = 1
        self.killed('install exit')

    def test_unexpected_tests_exit(self):
        self.oracle['phases']['tests']['exit'] = 124
        self.killed('tests exit')

    def test_diff_must_describe_snapshots(self):
        self.diff = '--- reference/api/typescript.d.ts\n+++ local/api/typescript.d.ts\n'
        self.killed('baseline.diff does not exactly describe')

    def test_method_conversion_is_not_a_property_union(self):
        self.local = self.local.replace('method?(s: string): void;',
                                        'method?: { method(s: string): void; }["method"] | undefined;')
        report = self.killed('unsanctioned API declaration')
        self.assertEqual(len(report['errors']), 1)

    def test_manifest_includes_only_proved_property_handoffs(self):
        manifest = json.loads((LANE / 'sanctioned-api.json').read_text())
        sites = json.loads((LANE.parent / 'adapt/32-indexed-reads-program/public-host-sites.json').read_text())
        changes = {row['declaration'] for row in manifest['normalized_changes']}
        owners = {tuple(row['path']) for row in manifest['owners'] if row['adaptation'] == '32'}
        self.assertEqual(owners, {('AmdDependency', 'name'), ('CommentRange', 'hasTrailingNewLine'),
                                  ('BuilderProgramHost', 'createHash')})
        methods = [site for site in sites if site['public'] and '?(' in site['before']]
        self.assertEqual(manifest['excluded_method_conversions'], len(methods))
        for site in methods:
            key = 'member:' + json.dumps(['ts', site['interface'], site['name']], separators=(',', ':')) + '#0'
            self.assertNotIn(key, changes)
        self.assertEqual(manifest['counts']['30'], 0)
        self.assertEqual(manifest['counts']['32'], 3)
        self.assertEqual(manifest['counts']['33'], 0)
        self.assertEqual(len(changes), 222)

    def test_unknown_platform(self):
        self.execution['platform']['node'] = 'v24.19.1'
        self.killed('unknown platform')

    def test_unknown_platform_stops_runner_before_apply(self):
        commands = self.results / 'commands'
        commands.mkdir()
        node = commands / 'node'
        node.write_text("#!/usr/bin/env bash\nprintf '%s\\n' "
                        "'{\"os\":\"linux\",\"arch\":\"x64\",\"node\":\"v0.0.0\"}'\n")
        node.chmod(0o755)
        env = dict(os.environ, PATH=str(commands) + os.pathsep + os.environ['PATH'])
        output = self.results / 'unknown-platform-run'
        result = subprocess.run(['bash', str(LANE / 'run.sh'), str(output)],
                                env=env, capture_output=True, text=True)
        self.assertEqual(result.returncode, 1)
        self.assertIn('unknown platform: linux/x64/v0.0.0', result.stdout)
        self.assertFalse((output / 'apply.log').exists())
        self.assertFalse((output / 'adapted-tree').exists())

    def test_measured_macos_counts(self):
        self.execution['platform'].update(os='darwin', arch='arm64')
        self.oracle['counts']['passing'] = 106369
        self.log = self.log.replace('106366 passing', '106369 passing')
        self.assertEqual(self.replay()['status'], 'pass')

    def test_node_version_matches_report(self):
        self.oracle['node'] = 'v24.19.1'
        self.killed('node version')

    def test_wrong_passed_count(self):
        self.oracle['counts']['passing'] += 1
        self.log = self.log.replace('106366 passing', '106367 passing')
        self.killed('test counts')

    def test_different_single_failure(self):
        self.log = self.log.replace('should be acknowledged when they change:',
                                    'should reject an unrelated runtime regression:')
        report = self.killed('failed test names')
        self.assertEqual(len(report['errors']), 1)

    def test_wrong_baseline_path(self):
        self.oracle['baseline_diffs'] = ['api/tsserverlibrary.d.ts']
        self.killed('baseline paths')

    def test_failed_apply(self):
        self.execution['apply_exit'] = 1
        self.killed('apply exit')

    def test_wrong_oracle_exit(self):
        self.execution['oracle_exit'] = 124
        self.killed('oracle exit')

    def test_failed_build(self):
        self.oracle['phases']['build']['exit'] = 1
        self.killed('build exit')

    def test_zero_failure_is_not_the_expected_failure(self):
        self.oracle['counts']['failing'] = 0
        self.killed('test counts')

    def test_pending_test(self):
        self.oracle['counts']['pending'] = 1
        self.killed('test counts')

    def test_reporter_counts_must_agree(self):
        self.log = self.log.replace('106366 passing', '106365 passing')
        report = self.killed('reporter counts')
        self.assertEqual(len(report['errors']), 1)

    def test_filtered_run(self):
        for field, value, label in [('runners', 'compiler', 'runners'),
                                    ('tests', 'Public APIs', 'test filter'),
                                    ('workers', 1, 'workers'), ('workers', 4, 'workers'),
                                    ('workers', 64, 'workers'),
                                    ('status', 'pass', 'oracle status')]:
            with self.subTest(field=field):
                oracle = copy.deepcopy(self.oracle)
                self.oracle[field] = value
                self.killed(label)
                self.oracle = oracle

    def test_missing_evidence(self):
        self.replay()
        (self.results / 'oracle/baseline.diff').unlink()
        report = check.check_results(self.results, sanctioned_file=self.sanctioned_file)
        self.assertEqual(report['status'], 'fail')
        self.assertIn('missing or invalid lane evidence', report['errors'][0])


if __name__ == '__main__':
    unittest.main(verbosity=2)
