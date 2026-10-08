#!/usr/bin/env python3
"""Behavior tests and deliberately broken implementations for each comparison check."""
import contextlib
import io
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('records', Path(__file__).with_name('records-check.py'))
records = importlib.util.module_from_spec(spec)
spec.loader.exec_module(records)

ROW = dict(package='stage1/cohere/example', test='TestGap', source='gap_test.go',
           records=['stage1/cohere/example/GAPS.md'], refresh='refresh the gap')


def result(row=ROW, outcomes=None, gaps=None):
    return {'runs': [dict(row=row, outcomes=outcomes or {}, gaps=gaps or {}, output={}, elapsed={}, seconds=0)]}


class Checks(unittest.TestCase):
    def comparison(self, before, after, trailer='', file='', row=ROW):
        return records.compare(result(row, before), result(row, after), trailer, file, Path('.'), Path('.'))

    def test_count_and_gap_moves_and_refreshed_records(self):
        for row in [ROW, dict(ROW, package='internal/oracle', test='TestCountsAreRecorded', records=['internal/oracle/counts.md'])]:
            test = row['test']
            changed = self.comparison({test: 'pass'}, {test: 'fail'}, row=row)
            self.assertEqual(changed['undeclared'], 1)
            self.assertEqual(changed['moved'][0]['record'], row['records'][0])
            self.assertEqual(changed['moved'][0]['refresh'], row['refresh'])
            self.assertEqual((changed['moved'][0]['old'], changed['moved'][0]['new']), ('pass', 'fail'))
            self.assertFalse(self.comparison({test: 'pass'}, {test: 'pass'}, row=row)['moved'])

    def test_removed_fixture_and_declaration_sources(self):
        row = dict(ROW, package='stage3/fixtures', test='TestFixtures')
        old = {'TestFixtures/objects/old.a': 'pass'}
        new = {'TestFixtures/objects/new.a': 'pass'}
        moves = self.comparison(old, new, row=row)
        self.assertEqual(moves['undeclared'], 2)
        self.assertEqual(moves['moved'][1]['new'], 'absent')
        self.assertEqual(moves['moved'][1]['record'], 'stage3/fixtures/objects/status.json')
        for source in ['trailer', 'file']:
            declaration = 'TestFixtures/objects -> TestFixtures/objects/new.a'
            report = self.comparison(old, new, **{source: declaration}, row=row)
            self.assertEqual(report['undeclared'], 0)
            self.assertTrue(all(source in move['declaration'] if source == 'trailer' else
                                move['declaration'] == 'moved-results.txt' for move in report['moved']))
        self.assertEqual(self.comparison(old, {}, file='# TestFixtures -> removed', row=row)['undeclared'], 1)
        self.assertEqual(self.comparison(old, {}, file='TestFixtures/object -> removed', row=row)['undeclared'], 1)

    def test_skip_and_empty_run_are_named_gaps(self):
        event = json.dumps(dict(Action='skip', Test='TestGap', Package='example'))
        self.assertIn('TestGap', records.parse_run(event, '', ROW, 0)['gaps'])
        self.assertIn('TestGap', records.parse_run('', 'build failed', ROW, 1)['gaps'])
        self.assertFalse(records.parse_run(json.dumps(dict(Action='pass', Test='TestGap')), '', ROW, 0)['gaps'])

    def test_platform_guard_lifted_without_modifying_input(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch) / 'tree'
            root.mkdir()
            source = root / 'fixture_test.go'
            original = 'if entry.Platform != "" && entry.Platform != runtime.GOOS {\n t.Skipf("platform")\n}\nif runtime.GOOS != "linux" {\n t.Skip("linux only")\n}'
            source.write_text(original)
            out = Path(scratch) / 'overlay'
            out.mkdir()
            path, lifted = records.platform_overlay(root, out)
            mapping = json.loads(path.read_text())['Replace']
            self.assertEqual(lifted, [str(source)])
            modified = Path(mapping[str(source)]).read_text()
            self.assertEqual(modified.count('false &&'), 2)
            self.assertEqual(source.read_text(), original)

    def test_cache_identity_includes_sha_go_node_and_inputs(self):
        baseline = records.cache_key('a', 'go1', 'node1')
        for args in [('b', 'go1', 'node1'), ('a', 'go2', 'node1'), ('a', 'go1', 'node2')]:
            self.assertNotEqual(baseline, records.cache_key(*args))
        with patch.dict(records.os.environ, {'ADAMIC_JSON_PRETTIER': '/new/oracle'}):
            self.assertNotEqual(baseline, records.cache_key('a', 'go1', 'node1'))

    def test_tools_inventory_follows_the_input_tree(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            self.assertFalse(records.inventory(root))
            pkg = root / 'stage1/new-area'
            pkg.mkdir(parents=True)
            (pkg / 'gaps_test.go').write_text('func TestNewGap(t *testing.T) {}')
            rows = records.inventory(root)
            self.assertEqual(len(rows), 1)
            self.assertEqual(rows[0]['package'], 'stage1/new-area')
            self.assertEqual(rows[0]['test'], 'TestNewGap')

    def test_base_runs_once_and_candidate_runs_each_time(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            @contextlib.contextmanager
            def checked_tree(repository, target):
                yield root, target
            with patch('records.tree', checked_tree), patch('records.sha', lambda root, ref: ref), \
                 patch('records.git', lambda *args: ''), patch('records.version', lambda *args: 'v1'), \
                 patch('records.run_tree', return_value=result(ROW, {'TestGap': 'pass'})) as run, \
                 contextlib.redirect_stdout(io.StringIO()), contextlib.redirect_stderr(io.StringIO()):
                arguments = ['candidate-one', '--base', 'fixed-base', '--cache', str(root / 'cache'), '--json']
                self.assertEqual(records.main(arguments), 0)
                arguments[0] = 'candidate-two'
                self.assertEqual(records.main(arguments), 0)
                self.assertEqual(run.call_count, 3)

    def test_parent_failure_is_not_reported_twice(self):
        report = self.comparison({'TestGap': 'pass', 'TestGap/a': 'pass'}, {'TestGap': 'fail', 'TestGap/a': 'fail'})
        self.assertEqual(len(report['moved']), 1)
        self.assertEqual(report['moved'][0]['test'], 'TestGap/a')


class Mutants(unittest.TestCase):
    def killed(self, method, target, replacement):
        suite = unittest.TestSuite([Checks(method)])
        with patch(target, replacement):
            outcome = unittest.TestResult()
            suite.run(outcome)
        self.assertFalse(outcome.wasSuccessful(), 'mutant survived: ' + target)

    def test_move_mutant(self):
        self.killed('test_count_and_gap_moves_and_refreshed_records', 'records.feedback.moved', lambda a, b: [])

    def test_declaration_mutant(self):
        self.killed('test_removed_fixture_and_declaration_sources', 'records.feedback.undeclared', lambda a, b: [])

    def test_skip_mutant(self):
        self.killed('test_skip_and_empty_run_are_named_gaps', 'records.parse_run', lambda *a: {'gaps': {}})

    def test_guard_mutant(self):
        original = records.platform_overlay
        def broken(root, directory):
            path, lifted = original(root, directory)
            path.write_text('{"Replace": {}}')
            return path, lifted
        self.killed('test_platform_guard_lifted_without_modifying_input', 'records.platform_overlay', broken)

    def test_cache_mutant(self):
        self.killed('test_cache_identity_includes_sha_go_node_and_inputs', 'records.cache_key', lambda *a: 'same')

    def test_tools_tree_mutant(self):
        self.killed('test_tools_inventory_follows_the_input_tree', 'records.inventory', lambda root: [ROW])

    def test_repeated_base_mutant(self):
        self.killed('test_base_runs_once_and_candidate_runs_each_time', 'records.Path.exists', lambda path: False)

    def test_duplicate_parent_mutant(self):
        original = records.compare
        def broken(*args):
            report = original(*args)
            report['moved'] *= 2
            return report
        self.killed('test_parent_failure_is_not_reported_twice', 'records.compare', broken)


if __name__ == '__main__':
    import sys
    sys.modules['records'] = records
    unittest.main()
