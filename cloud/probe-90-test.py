#!/usr/bin/env python3
"""The 90-second probe's offline parts: the phase mapping from sample records (a pool record's events, a box record's
steps_seconds), the status line's format and length, the branch naming and skip check, the plist, and a --dry-run of
cloud/probe-90.sh against local Git remotes only. No push, no network, no ahra."""
from datetime import datetime, timezone
import gzip
import importlib.util
import json
import os
from pathlib import Path
import plistlib
import re
import shutil
import subprocess
import tempfile
import unittest

here = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('probe', here / 'probe-90' / 'record.py')
probe = importlib.util.module_from_spec(spec)
spec.loader.exec_module(probe)

os.environ['GIT_CONFIG_GLOBAL'] = os.devnull


def epoch(text):
    return datetime.strptime(text, '%Y%m%dT%H%M%SZ').replace(tzinfo=timezone.utc).timestamp()


# Loom's pool record of devtools/setup-tsc 318eef6a (gate-logs/318eef6a4ccc/20261009T154811Z/fast), trimmed to the
# events the mapping reads; its record commit is 15:51:07Z.
poolFast = {'finished': True, 'verdict': 'green', 'runner': 'pool', 'pool': 'codex-side',
            'pool_run': 'adamic-verify-20261009T154944-cb9f12bd', 'covers': 'go-tests',
            'sha': '318eef6a4cccf979bc4d7026fad7f7838993282f', 'branch': 'devtools/setup-tsc'}
poolEvents = [
    {'run': 'adamic-verify-20261009T154944-cb9f12bd', 'verdict': {'status': 'green'}},
    {'unit': 'tests-00', 'time': '2026-10-09T15:49:44.948Z', 'type': 'started'},
    {'unit': 'build-vet', 'time': '2026-10-09T15:49:44.911Z', 'type': 'started'},
    {'unit': 'tests-00', 'time': '2026-10-09T15:49:46.521Z', 'type': 'output',
     'text': 'loom-pilot: 473509fb286d slot ? cpus ? tree 318eef6a setup 2 s, 2 packages at a time'},
    {'unit': 'build-vet', 'time': '2026-10-09T15:49:46.534Z', 'type': 'output', 'text': 'loom-build: tree 318eef6a setup 2 s'},
    {'unit': 'tests-00', 'time': '2026-10-09T15:49:54.846Z', 'type': 'output', 'text': 'loom-pilot: tests took 10 s in all'},
    {'unit': 'tests-00', 'time': '2026-10-09T15:49:54.847Z', 'type': 'exit', 'code': 0, 'wallSeconds': 9.898},
    {'unit': 'tests-00', 'time': '2026-10-09T15:49:55.836Z', 'type': 'finished', 'status': 'passed'},
    {'unit': 'build-vet', 'time': '2026-10-09T15:50:40.092Z', 'type': 'output', 'text': 'loom-build: go build ./... passed'},
    {'unit': 'build-vet', 'time': '2026-10-09T15:50:40.093Z', 'type': 'exit', 'code': 0, 'wallSeconds': 55.181},
    {'unit': 'build-vet', 'time': '2026-10-09T15:50:40.715Z', 'type': 'finished', 'status': 'passed'},
]
poolStamp, poolCommit = epoch('20261009T154811Z'), epoch('20261009T155107Z')

# A box record's fields (canary/main e77a4ae4 on Home, gate-logs/e77a4ae41f47/20261009T162356Z/fast).
boxFast = {'sha': 'e77a4ae41f473c149aee910c51b73637686a804a', 'tools_sha': 'eaa0dc8bdbc4bfef446f4167615ee58a2bdfd2a6',
           'wall_seconds': 29.4, 'steps_seconds': {'coverage': 0.0, 'tools': 0.1, 'vet': 2.1, 'smoke': 9.8, 'determinism': 14.0,
                                                   'build': 14.8, 'stage3': 26.2, 'products': 26.131, 'tests': 27.8}}
boxStamp = epoch('20261009T162356Z')
boxCommit = boxStamp + 60


def writeRecord(directory, fast, events=None, status='green: abc fast gate', gz=False):
    directory = Path(directory)
    (directory / 'fast.json').write_text(json.dumps(fast))
    (directory / 'status.txt').write_text(status + '\n')
    if events is not None:
        text = ''.join(json.dumps(event) + '\n' for event in events)
        if gz:
            with gzip.open(directory / 'record.jsonl.gz', 'wt') as handle:
                handle.write(text)
        else:
            (directory / 'record.jsonl').write_text(text)


class MappingTests(unittest.TestCase):
    def test_pool_record_maps_from_its_events_and_clocks(self):
        with tempfile.TemporaryDirectory() as directory:
            writeRecord(directory, poolFast, poolEvents)
            route, _, mapped = probe.phases(directory, poolStamp, poolCommit)
        self.assertEqual(route, 'pool')
        self.assertAlmostEqual(mapped['select'], 93.911, places=3)  # serve start to the first unit placed
        self.assertEqual(mapped['fetch'], 2.0)  # the longest unit setup
        self.assertAlmostEqual(mapped['build'], 53.181, places=3)  # build-vet's wall less its setup
        self.assertEqual(mapped['test'], 10.0)  # the longest "tests took"
        self.assertAlmostEqual(mapped['publish'], 26.285, places=3)  # last unit finished to the record commit
        self.assertEqual(mapped['wall'], 176.0)

    def test_gzipped_events_read_the_same(self):
        with tempfile.TemporaryDirectory() as directory:
            writeRecord(directory, poolFast, poolEvents, gz=True)
            _, _, mapped = probe.phases(directory, poolStamp, poolCommit)
        self.assertEqual(mapped['test'], 10.0)

    def test_a_test_unit_without_its_took_line_uses_its_wall_less_setup(self):
        events = [event for event in poolEvents if 'tests took' not in event.get('text', '')]
        self.assertAlmostEqual(probe.poolPhases(events, poolStamp, poolCommit)['test'], 7.898, places=3)

    def test_an_earlier_attempts_run_times_nothing_against_this_stamp(self):
        mapped = probe.poolPhases(poolEvents, poolStamp + 3600, poolCommit + 3600)
        self.assertIsNone(mapped['select'])
        self.assertIsNone(mapped['publish'])
        self.assertEqual((mapped['fetch'], mapped['test']), (2.0, 10.0))

    def test_box_record_maps_from_steps_seconds(self):
        with tempfile.TemporaryDirectory() as directory:
            writeRecord(directory, boxFast)
            route, _, mapped = probe.phases(directory, boxStamp, boxCommit)
        self.assertEqual(route, 'box')
        self.assertAlmostEqual(mapped['select'], 1.6, places=3)  # run.py's wall less its longest step
        self.assertAlmostEqual(mapped['fetch'], 30.6, places=3)  # stamp to commit less run.py's wall
        self.assertEqual(mapped['build'], 14.8)
        self.assertEqual(mapped['test'], 27.8)
        self.assertIsNone(mapped['publish'])
        self.assertEqual(mapped['wall'], 60.0)

    def test_a_record_with_no_events_maps_to_unknowns_not_a_crash(self):
        with tempfile.TemporaryDirectory() as directory:
            writeRecord(directory, poolFast)
            _, _, mapped = probe.phases(directory, poolStamp, poolCommit)
        self.assertEqual([mapped[name] for name in ('select', 'fetch', 'build', 'test', 'publish')], [None] * 5)

    def test_a_record_with_no_fast_json_reads_as_an_unknown_box_run(self):
        with tempfile.TemporaryDirectory() as directory:
            (Path(directory) / 'status.txt').write_text('red: abc fast gate, first failure at tests\n')
            route, fast, mapped = probe.phases(directory, boxStamp, boxCommit)
        self.assertEqual((route, fast, mapped['wall'], mapped['test']), ('box', {}, 60.0, None))


class LineTests(unittest.TestCase):
    pattern = re.compile(r'^probe-90 (leaf|all) (green|red): wall \d+s = select (\d+|-) fetch (\d+|-) build (\d+|-) '
                         r'test (\d+|-) publish (\d+|-)(; queue \d+s)?; (pool|box), tools [0-9a-f]{9}, run \S.*$')

    def assertGoodLine(self, line):
        self.assertLessEqual(len(line), 160, line)
        self.assertNotIn('\n', line)
        self.assertRegex(line, self.pattern)
        # ahra refuses all-caps words of three letters or more.
        self.assertIsNone(re.search(r'\b[A-Z]{3,}\b', line), line)

    def test_pool_line(self):
        with tempfile.TemporaryDirectory() as directory:
            writeRecord(directory, poolFast, poolEvents)
            route, fast, mapped = probe.phases(directory, poolStamp, poolCommit)
        line = probe.statusLine('leaf', 'green', route, mapped, 41, '95eb1c8fd6a310758b9de0620e54c05389048fc1',
                                probe.runIdentifier(fast, 'gate-logs/318eef6a4ccc/20261009T154811Z/fast'))
        self.assertEqual(line, 'probe-90 leaf green: wall 176s = select 94 fetch 2 build 53 test 10 publish 26; queue 41s; '
                               'pool, tools 95eb1c8fd, run adamic-verify-20261009T154944-cb9f12bd')
        self.assertGoodLine(line)

    def test_box_line_names_its_record(self):
        with tempfile.TemporaryDirectory() as directory:
            writeRecord(directory, boxFast)
            route, fast, mapped = probe.phases(directory, boxStamp, boxCommit)
        line = probe.statusLine('all', 'red', route, mapped, 1234, boxFast['tools_sha'],
                                probe.runIdentifier(fast, 'gate-logs/e77a4ae41f47/20261009T162356Z/fast'))
        self.assertIn('publish -;', line)
        self.assertTrue(line.endswith('run record 20261009T162356Z'), line)
        self.assertGoodLine(line)

    def test_a_long_line_drops_the_queue_then_stays_within_160(self):
        mapped = {'wall': 99999, 'select': 99999, 'fetch': 99999, 'build': 99999, 'test': 99999, 'publish': 99999}
        line = probe.statusLine('all', 'green', 'pool', mapped, 99999, 'a' * 40, 'adamic-verify-20261009T154944-cb9f12bd')
        self.assertNotIn('queue', line)
        self.assertGoodLine(line)
        line = probe.statusLine('all', 'green', 'pool', mapped, 9, 'a' * 40, 'x' * 200)
        self.assertEqual(len(line), 160)

    def test_cli_line_reads_a_record_directory(self):
        with tempfile.TemporaryDirectory() as directory:
            writeRecord(directory, poolFast, poolEvents, status='green: 318eef6a fast gate on Loom\'s side pool')
            out = subprocess.run(['python3', str(here / 'probe-90' / 'record.py'), 'line', directory, '--mode', 'leaf',
                                  '--sha', poolFast['sha'], '--ref', 'gate-logs/318eef6a4ccc/20261009T154811Z/fast',
                                  '--tools', 'b' * 40, '--stamp-epoch', str(poolStamp), '--commit-epoch', str(poolCommit),
                                  '--pushed-epoch', str(poolStamp - 41)], capture_output=True, text=True, check=True).stdout
        self.assertGoodLine(out.strip())
        self.assertTrue(out.startswith('probe-90 leaf green: wall 176s'))


class BranchTests(unittest.TestCase):
    def test_names(self):
        self.assertEqual(probe.branchName('20261009T170000Z', 'leaf'), 'devtools/probe-90-20261009T170000Z')
        self.assertEqual(probe.branchName('20261009T170000Z', 'all'), 'devtools/probe-90-20261009T170000Z-all')
        with self.assertRaises(ValueError):
            probe.branchName('2026-10-09', 'leaf')
        with self.assertRaises(ValueError):
            probe.branchName('20261009T170000Z', 'everything')

    def test_a_run_is_a_candidate_the_watcher_gates_and_never_the_tools_branch(self):
        name = probe.branchName('20261009T170000Z', 'leaf')
        self.assertTrue(name.startswith('devtools/'))  # cloud/fast-gate-watch.sh gates devtools/* tips
        self.assertNotEqual(name, 'devtools/probe-90')
        globs = ('devtools/fast-gate* the gate tools\n'
                 'devtools/probe-90 the probe tools branch, never a candidate\n'
                 '# devtools/* a comment line\n'
                 'devtools/gate-* gate-tool branches\n')
        self.assertIsNone(probe.skippedBy(name, globs))
        self.assertIsNone(probe.skippedBy(probe.branchName('20261009T170000Z', 'all'), globs))
        self.assertEqual(probe.skippedBy('devtools/probe-90', globs), 'devtools/probe-90')

    def test_skip_globs_match_as_the_watcher_does(self):
        # bash [[ == ]]: * crosses slashes.
        self.assertEqual(probe.skippedBy('devtools/probe-90-20261009T170000Z', 'devtools/* all of them'), 'devtools/*')
        self.assertEqual(probe.skippedBy('devtools/probe-90-20261009T170000Z', 'devtools/probe-90* a family'), 'devtools/probe-90*')


class ConfigurationTests(unittest.TestCase):
    def test_plist(self):
        with open(here / 'probe-90' / 'com.adamic.probe-90.plist', 'rb') as handle:
            plist = plistlib.load(handle)
        self.assertEqual(plist['Label'], 'com.adamic.probe-90')
        self.assertEqual(plist['StartInterval'], 1800)
        self.assertEqual(plist['StandardOutPath'], '/Users/kirkouimet/Projects/system/adamic-gate-logs/probe-90.log')
        self.assertEqual(plist['StandardErrorPath'], plist['StandardOutPath'])
        self.assertTrue(plist['ProgramArguments'][-1].endswith('/adamic-gate-watch/cloud/probe-90.sh'))
        self.assertIn('/Users/kirkouimet/Projects/ahra/node_modules/.bin', plist['EnvironmentVariables']['PATH'])


class DryRunTests(unittest.TestCase):
    """cloud/probe-90.sh --dry-run in a copy of its tools beside a local origin: the commit it would gate."""

    def setUp(self):
        self.root = Path(tempfile.mkdtemp())
        git = lambda *args, cwd: subprocess.run(['git', *args], cwd=cwd, check=True, capture_output=True, text=True).stdout
        self.git = git
        origin = self.root / 'origin.git'
        git('init', '-q', '--bare', '-b', 'main', str(origin), cwd=self.root)
        seed = self.root / 'seed'
        seed.mkdir()
        git('init', '-q', '-b', 'main', cwd=seed)
        for path in ('cmd/adamic-stage1-progress/main_test.go', 'internal/load/load.go'):
            (seed / path).parent.mkdir(parents=True, exist_ok=True)
            (seed / path).write_text('package x\n\nfunc f() {}\n')
        git('add', '.', cwd=seed)
        git('-c', 'user.name=t', '-c', 'user.email=t@t.invalid', 'commit', '-qm', 'main', cwd=seed)
        git('push', '-q', str(origin), 'main', cwd=seed)
        self.main = git('rev-parse', 'HEAD', cwd=seed).strip()
        self.tools = self.root / 'tools'
        git('clone', '-q', str(origin), str(self.tools), cwd=self.root)
        (self.tools / 'cloud' / 'probe-90').mkdir(parents=True)
        shutil.copy(here / 'probe-90.sh', self.tools / 'cloud' / 'probe-90.sh')
        shutil.copy(here / 'probe-90' / 'record.py', self.tools / 'cloud' / 'probe-90' / 'record.py')
        self.state = self.root / 'state'
        self.state.mkdir()

    def tearDown(self):
        shutil.rmtree(self.root)

    def run_probe(self, *arguments, **environment):
        env = dict(os.environ, ADAMIC_FAST_GATE_WATCH_STATE=str(self.state), ADAMIC_PROBE_POST='0', **environment)
        return subprocess.run(['bash', str(self.tools / 'cloud' / 'probe-90.sh'), *arguments], env=env, capture_output=True, text=True)

    def commitOf(self, output):
        match = re.search(r'(devtools/probe-90-\d{8}T\d{6}Z(?:-all)?) ([0-9a-f]{40}): (leaf|all) probe of (\S+)', output)
        self.assertIsNotNone(match, output)
        return match.groups()

    def test_leaf_dry_run_appends_one_comment_line_to_the_leaf_on_mains_tip(self):
        result = self.run_probe('--dry-run')
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        branch, sha, mode, path = self.commitOf(result.stdout)
        self.assertEqual((mode, path), ('leaf', 'cmd/adamic-stage1-progress/main_test.go'))
        self.assertEqual(self.git('rev-parse', sha + '^', cwd=self.tools).strip(), self.main)
        self.assertEqual(self.git('diff', '--name-only', self.main, sha, cwd=self.tools).split(), [path])
        added = [line for line in self.git('diff', '-U0', self.main, sha, cwd=self.tools).splitlines() if line.startswith('+') and not line.startswith('+++')]
        self.assertEqual(len(added), 2)
        self.assertEqual(added[0], '+')
        self.assertTrue(added[1].startswith('+// probe-90 '), added)
        self.assertIn('Co-Authored-By: Ahra <ahra@ahra.ai>', self.git('log', '-1', '--format=%B', sha, cwd=self.tools))
        # The watcher's poolTier reads these as git trailers: tier 30 at Loom, and the task.
        self.assertEqual(self.git('log', '-1', '--format=%(trailers:key=Gate-tier,valueonly)', sha, cwd=self.tools).strip(), '30')
        self.assertEqual(self.git('log', '-1', '--format=%(trailers:key=Task,valueonly)', sha, cwd=self.tools).strip(), '#c89weq5')
        self.assertIn('dry run: not pushed', result.stdout)
        # Nothing reached origin, and the tools checkout's tree is untouched.
        self.assertNotIn('probe-90', self.git('ls-remote', 'origin', cwd=self.tools))
        self.assertEqual(self.git('status', '--porcelain', '--untracked-files=no', cwd=self.tools), '')

    def test_force_all_probes_the_wide_file(self):
        result = self.run_probe('--dry-run', ADAMIC_PROBE_FORCE_ALL='1')
        branch, _, mode, path = self.commitOf(result.stdout)
        self.assertEqual((mode, path), ('all', 'internal/load/load.go'))
        self.assertTrue(branch.endswith('-all'))

    def test_a_skipped_branch_is_never_made(self):
        (self.state / 'skip-globs').write_text('devtools/probe-90-* paused by hand\n')
        result = self.run_probe('--dry-run')
        self.assertEqual(result.returncode, 2)
        self.assertIn('skipped by devtools/probe-90-*', result.stdout)


if __name__ == '__main__':
    unittest.main()
