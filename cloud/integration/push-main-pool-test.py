#!/usr/bin/env python3
"""push-main's verdict on a whole-gate record from Loom's pool (@system_adamic, Oct 9 02:33Z): it lands only once the
pool is promoted, only when it covers every stage a box's whole gate runs, and for a fifth of shas only beside a green
box record. Runs the verdict push-main.sh itself carries, against planted records in a scratch repository.

usage: python3 cloud/integration/push-main-pool-test.py
"""
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
import unittest

script = Path(__file__).with_name('push-main.sh')
verdict = re.search(r"<<'VERDICT'\n(.*?)\nVERDICT\n", script.read_text(), re.S).group(1)
stages = ["coverage", "tools", "build", "vet", "tests", "wasi", "stage3", "catalog", "determinism", "census"]


def git(directory, *arguments):
    return subprocess.run(['git', '-C', str(directory), '-c', 'user.name=t', '-c', 'user.email=t@t'] + list(arguments),
                          check=True, capture_output=True, text=True).stdout.strip()


class PoolRecordTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        root = Path(self.tmp.name)
        self.origin = root / 'origin.git'
        git(root, 'init', '-q', '--bare', str(self.origin))
        self.repository = root / 'repository'
        git(root, 'init', '-q', str(self.repository))
        git(self.repository, 'remote', 'add', 'origin', str(self.origin))
        (self.repository / 'a.txt').write_text('a\n')
        git(self.repository, 'add', '.')
        git(self.repository, 'commit', '-qm', 'main')
        git(self.repository, 'push', '-q', 'origin', 'HEAD:refs/heads/main')
        git(self.repository, 'fetch', '-q', 'origin')
        self.promoted = root / 'pool-promoted'

    def commit(self, spotChecked):
        # Commits until the sha falls on the side of the spot check the test wants.
        for attempt in range(200):
            (self.repository / 'b.txt').write_text('%d\n' % attempt)
            git(self.repository, 'add', '.')
            git(self.repository, 'commit', '-qm', 'candidate %d' % attempt)
            sha = git(self.repository, 'rev-parse', 'HEAD')
            if (int(sha[:8], 16) % 5 == 0) == spotChecked:
                return sha
        raise AssertionError('no sha on that side of the spot check')

    def record(self, sha, runner='pool', planned=stages, covers=None):
        record = {'sha': sha, 'base': sha, 'fail': 0, 'pass': 10, 'skip': 0, 'build_ok': True, 'vet_ok': True,
                  'uncached_tests': True, 'packages': 'all', 'package_list': [], 'finished': True, 'wall_seconds': 600,
                  'steps_seconds': {stage: 1.0 for stage in planned}, 'stages_exit': {stage: 0 for stage in planned},
                  'planned_stages': list(planned), 'runner': runner}
        if covers is not None:
            record['covers'] = covers
        return record

    def judge(self, sha, record):
        path = Path(self.tmp.name) / 'full.json'
        path.write_text(json.dumps(record))
        return subprocess.run(['python3', '-', sha, 'green: %s full gate' % sha, 'gate-logs/x/full-main', str(path)], input=verdict,
                              cwd=self.repository, capture_output=True, text=True,
                              env=dict(os.environ, GATE_KIND='full', PUSH_MAIN_POOL_PROMOTED=str(self.promoted),
                                       RERUN_MERGE=str(Path(__file__).with_name('rerun_merge.py'))))

    def publishBox(self, sha, record=None, status=None):
        tree = Path(self.tmp.name) / 'box'
        tree.mkdir(exist_ok=True)
        (tree / 'full.json').write_text(json.dumps(record or self.record(sha, runner='box')))
        (tree / 'status.txt').write_text(status or 'green: %s full gate in 3000 s\n' % sha)
        index = str(Path(self.tmp.name) / 'index')
        environment = dict(os.environ, GIT_INDEX_FILE=index)
        gitDirectory = git(self.repository, 'rev-parse', '--absolute-git-dir')
        subprocess.run(['git', '--git-dir', gitDirectory, '--work-tree', str(tree), 'add', '-A', '.'], env=environment, check=True)
        treeSha = subprocess.run(['git', '--git-dir', gitDirectory, 'write-tree'], env=environment, check=True, capture_output=True, text=True).stdout.strip()
        commit = git(self.repository, 'commit-tree', treeSha, '-m', 'record')
        git(self.repository, 'push', '-q', 'origin', '%s:refs/heads/gate-logs/%s/20261009T000000Z/full-main' % (commit, sha[:12]))

    def test_a_pool_record_lands_only_once_promoted(self):
        sha = self.commit(spotChecked=False)
        refused = self.judge(sha, self.record(sha))
        self.assertNotEqual(refused.returncode, 0)
        self.assertIn("the pool isn't promoted to land yet", refused.stdout)
        self.promoted.touch()
        self.assertEqual(self.judge(sha, self.record(sha)).returncode, 0)
        # A box record never needed the switch.
        self.promoted.unlink()
        self.assertEqual(self.judge(sha, self.record(sha, runner='box')).returncode, 0)

    def test_a_pool_record_covers_every_stage_of_a_whole_gate(self):
        self.promoted.touch()
        sha = self.commit(spotChecked=False)
        partial = self.judge(sha, self.record(sha, planned=[stage for stage in stages if stage != 'catalog']))
        self.assertIn("doesn't cover every stage of a whole gate (missing catalog", partial.stdout)
        goTestsOnly = self.judge(sha, self.record(sha, covers='go-tests'))
        self.assertIn("covers 'go-tests'", goTestsOnly.stdout)

    def test_a_spot_checked_sha_needs_a_green_box_record_too(self):
        self.promoted.touch()
        sha = self.commit(spotChecked=True)
        alone = self.judge(sha, self.record(sha))
        self.assertIn('the boxes spot-check', alone.stdout)
        self.publishBox(sha)
        self.assertEqual(self.judge(sha, self.record(sha)).returncode, 0, self.judge(sha, self.record(sha)).stdout)


    def test_a_rerun_lands_on_its_base_runs_kept_verdicts_and_is_refused_when_a_kept_units_inputs_moved(self):
        # #hpjftdj: the base run killed TestSlow at 90 s; the fix reruns only it, on a descendant sha.
        self.promoted.touch()
        base = self.commit(spotChecked=False)
        baseRecord = dict(self.record(base), units=[{'id': 'p/TestA', 'input_sha256': 'h1', 'verdict': 'passed'},
                                                    {'id': 'p/TestSlow', 'input_sha256': 'h2', 'verdict': 'killed'}])
        self.publishBox(base, record=baseRecord, status='red: %s full gate\n' % base)
        # The fix: a commit on top of the base (commit() alone could rewrite b.txt with what it already holds).
        (self.repository / 'fix.txt').write_text('split\n')
        sha = self.commit(spotChecked=False)
        ref = 'gate-logs/%s/20261009T000000Z/full-main' % base[:12]
        rerun = dict(self.record(sha), rerun_of=ref, units=[{'id': 'p/TestA', 'input_sha256': 'h1', 'verdict': 'kept'},
                                                            {'id': 'p/TestSlow', 'input_sha256': 'h2b', 'verdict': 'passed'}])
        landed = self.judge(sha, rerun)
        self.assertEqual(landed.returncode, 0, landed.stdout + landed.stderr)
        rerun['units'][0]['input_sha256'] = 'h1-moved'
        refused = self.judge(sha, rerun)
        self.assertNotEqual(refused.returncode, 0)
        self.assertIn('p/TestA is kept but its inputs changed', refused.stdout)
        missing = self.judge(sha, dict(rerun, rerun_of='gate-logs/000000000000/none/full-main'))
        self.assertIn('has no readable full.json on origin', missing.stdout)


if __name__ == '__main__':
    unittest.main()
