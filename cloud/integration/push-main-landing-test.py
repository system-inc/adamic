#!/usr/bin/env python3
"""push-main's landings, end to end in a scratch repository (@system_adamic, Oct 9): each landing is one commit on
main, first parent the old main, second the landed sha, its tree exactly the landing's and its numbers as trailers,
and landings.py reads them back as the velocity table. Main moving by test-only commits doesn't spend a candidate's
gate unless they touch a package where the candidate changes code.

usage: python3 cloud/integration/push-main-landing-test.py
"""
import csv
import io
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

directory = Path(__file__).parent
identity = {'GIT_AUTHOR_NAME': 't', 'GIT_AUTHOR_EMAIL': 't@t', 'GIT_COMMITTER_NAME': 't', 'GIT_COMMITTER_EMAIL': 't@t'}


def git(where, *arguments):
    return subprocess.run(['git', '-C', str(where)] + list(arguments), check=True, capture_output=True, text=True,
                          env=dict(os.environ, **identity)).stdout.strip()


class LandingTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        root = Path(self.tmp.name)
        origin = root / 'origin.git'
        git(root, 'init', '-q', '--bare', str(origin))
        self.repository = root / 'repository'
        git(root, 'init', '-q', str(self.repository))
        git(self.repository, 'remote', 'add', 'origin', str(origin))
        self.write('code/a.go', 'package code\n')
        self.write('other/b.go', 'package other\n')
        self.main = self.commit('main')
        git(self.repository, 'push', '-q', 'origin', 'HEAD:refs/heads/main')
        # Developer tools' declared tools, which the lane's checks read.
        self.write('cloud/fast-gate/tools.txt', 'go\tall\tgo version\n')
        git(self.repository, 'push', '-q', 'origin', self.commit('tools') + ':refs/heads/devtools/fast-gate')
        git(self.repository, 'checkout', '-q', '--detach', self.main)
        # The scripts under test, beside a merge-back that does nothing.
        self.scripts = root / 'scripts'
        self.scripts.mkdir()
        for name in ('push-main.sh', 'landings.py', 'lane-checks.py'):
            shutil.copy(directory / name, self.scripts / name)
        (self.scripts / 'merge-back.sh').write_text('#!/usr/bin/env bash\n')
        (self.scripts / 'merge-back.sh').chmod(0o755)

    def write(self, path, text):
        (self.repository / path).parent.mkdir(parents=True, exist_ok=True)
        (self.repository / path).write_text(text)

    def commit(self, message, on=None):
        if on:
            git(self.repository, 'checkout', '-q', '--detach', on)
        git(self.repository, 'add', '-A')
        git(self.repository, 'commit', '-qm', message)
        return git(self.repository, 'rev-parse', 'HEAD')

    def change(self, on, path, text, message):
        git(self.repository, 'checkout', '-q', '--detach', on)
        self.write(path, text)
        return self.commit(message)

    def push(self, *arguments):
        # The star's train and the fast gate's watcher keep their state in scratch directories here.
        root = Path(self.tmp.name)
        return subprocess.run(['bash', str(self.scripts / 'push-main.sh')] + list(arguments), cwd=self.repository,
                              capture_output=True, text=True, env=dict(os.environ, ADAMIC_STAR_FILE=str(root / 'star'),
                                                                       ADAMIC_FAST_GATE_STATE=str(root / 'watch'),
                                                                       ADAMIC_LANE_TREE=str(root / 'lane'), **identity))

    def main_now(self):
        git(self.repository, 'fetch', '-q', 'origin')
        return git(self.repository, 'rev-parse', 'origin/main')

    def assertLanded(self, landed, old, sha):
        self.assertEqual(landed.returncode, 0, landed.stdout + landed.stderr)
        new = self.main_now()
        self.assertEqual(git(self.repository, 'rev-list', '--parents', '-n', '1', new).split()[1:], [old, sha])
        return new

    def test_a_gated_landing_is_one_commit_with_its_numbers_as_trailers(self):
        sha = self.change(self.main, 'code/a.go', 'package code\n\n// new\n', 'code change')
        new = self.assertLanded(self.push(sha, '12', '40', '0', '3', 'code/branch, at 1234'), self.main, sha)
        self.assertEqual(git(self.repository, 'rev-parse', new + '^{tree}'), git(self.repository, 'rev-parse', sha + '^{tree}'))
        trailers = git(self.repository, 'log', '-1', '--format=%(trailers:only,unfold)', new)
        for line in ('Old-main: ' + self.main, 'Landed-commits: 2', 'Gate-minutes: 12', 'Pass: 40', 'Fail: 0', 'Skip: 3', 'Branches: code/branch, at 1234'):
            self.assertIn(line, trailers)
        # Main moved once: nothing on top of the landing commit.
        self.assertEqual(git(self.repository, 'rev-list', '--count', sha + '..' + new), '1')
        rows = list(csv.DictReader(io.StringIO(subprocess.run(['python3', str(self.scripts / 'landings.py')], cwd=self.repository,
                                                              capture_output=True, text=True, check=True).stdout)))
        self.assertEqual([(row['new_main'], row['old_main'], row['gate_minutes']) for row in rows], [(new, self.main, '12')])
        self.assertTrue(rows[0]['branches_landed'].startswith('code/branch; at 1234; lane checks '), rows[0]['branches_landed'])

    def test_a_test_only_landing_is_one_commit_too(self):
        sha = self.change(self.main, 'code/a_test.go', 'package code\n', 'a test')
        new = self.assertLanded(self.push('--test-only', sha, 'a split'), self.main, sha)
        self.assertIn('Gate-minutes: 0', git(self.repository, 'log', '-1', '--format=%B', new))
        refused = self.push('--test-only', self.change(new, 'code/a.go', 'package code\n\n// x\n', 'code'), 'not a split')
        self.assertIn('not test-only', refused.stderr)
        # The lane's checks refuse a test that shells out to a tool the gate doesn't declare, or isn't gofmt'd.
        tool = self.change(new, 'code/tool_test.go', 'package code\n\nimport "os/exec"\n\nvar _ = exec.Command("timeout", "1")\n', 'a tool')
        self.assertIn("runs timeout, which cloud/fast-gate/tools.txt doesn't declare", self.push('--test-only', tool, 'tool').stderr)
        messy = self.change(new, 'code/messy_test.go', 'package code\nvar  x = 1\n', 'messy')
        self.assertIn("code/messy_test.go isn't gofmt-formatted", self.push('--test-only', messy, 'messy').stderr)
        # A branch whose net diff is tests but whose history changes code is refused by that commit: landing it
        # would record the code commit as merged with its change dropped (cohere's estree split, Oct 9).
        code = self.change(new, 'code/a.go', 'package code\n\n// carried\n', 'a code change')
        undone = self.change(code, 'code/a.go', git(self.repository, 'show', new + ':code/a.go') + '\n', 'undo it')
        carried = self.change(undone, 'code/carried_test.go', 'package code\n', 'a test on top')
        refused = self.push('--test-only', carried, 'carried')
        self.assertIn('carries non-test history: %s' % code[:8], refused.stderr)

    def test_a_candidate_built_ahead_lands_over_the_landing_commit_below_it(self):
        lower = self.change(self.main, 'code/a.go', 'package code\n\n// lower\n', 'lower')
        upper = self.change(lower, 'other/b.go', 'package other\n\n// upper\n', 'upper')
        landed = self.assertLanded(self.push(lower, '1', '1', '0', '0', 'lower'), self.main, lower)
        top = self.assertLanded(self.push(upper, '1', '1', '0', '0', 'upper'), landed, upper)
        self.assertEqual(git(self.repository, 'rev-parse', top + '^{tree}'), git(self.repository, 'rev-parse', upper + '^{tree}'))

    def test_main_moving_by_tests_keeps_a_gate_outside_the_candidates_packages(self):
        candidate = self.change(self.main, 'code/a.go', 'package code\n\n// candidate\n', 'candidate')
        elsewhere = self.change(self.main, 'other/b_test.go', 'package other\n', 'a test elsewhere')
        moved = self.assertLanded(self.push('--test-only', elsewhere, 'elsewhere'), self.main, elsewhere)
        landed = self.assertLanded(self.push(candidate, '1', '1', '0', '0', 'candidate'), moved, candidate)
        self.assertEqual(git(self.repository, 'show', landed + ':other/b_test.go'), 'package other')
        # A test in the candidate's own package meets its new code, so the gate is spent.
        second = self.change(landed, 'code/a.go', 'package code\n\n// second\n', 'second')
        same = self.change(landed, 'code/a_test.go', 'package code\n', 'a test beside it')
        self.assertLanded(self.push('--test-only', same, 'beside'), landed, same)
        refused = self.push(second, '1', '1', '0', '0', 'second')
        self.assertIn('beyond record and test-only commits (code/a_test.go)', refused.stderr)


    def test_a_test_beside_the_stars_code_waits_while_its_gate_runs(self):
        root = Path(self.tmp.name)
        star = self.change(self.main, 'code/a.go', 'package code\n\n// the star\n', 'the star')
        git(self.repository, 'push', '-q', 'origin', star + ':refs/heads/cloud/land-train-9-slice-' + star[:8])
        (root / 'star').write_text('cloud/land-train-9-slice-%s\n' % star[:8])
        (root / 'watch' / 'running').mkdir(parents=True)
        (root / 'watch' / 'running' / '123').write_text('cloud/land-train-9-slice-%s %s S box B x:1 log\n' % (star[:8], star))
        beside = self.change(self.main, 'code/a_test.go', 'package code\n', 'a test beside the star')
        held = self.push('--test-only', beside, 'beside')
        self.assertEqual(held.returncode, 3, held.stdout + held.stderr)
        self.assertIn('held for the star: cloud/land-train-9-slice-%s %s has its fast gate running and changes code in code' % (star[:8], star[:8]), held.stderr)
        # Elsewhere keeps flowing.
        elsewhere = self.change(self.main, 'other/b_test.go', 'package other\n', 'elsewhere')
        moved = self.assertLanded(self.push('--test-only', elsewhere, 'elsewhere'), self.main, elsewhere)
        # The gate finishes: the held change lands.
        (root / 'watch' / 'running' / '123').unlink()
        self.assertLanded(self.push('--test-only', beside, 'beside'), moved, beside)


    def publish(self, sha, name, record):
        # A gate-logs record on origin: status.txt and fast.json in a commit of their own.
        tree = Path(self.tmp.name) / ('record-' + name)
        tree.mkdir()
        (tree / 'fast.json').write_text(json.dumps(record))
        (tree / 'status.txt').write_text('green: %s fast gate\n' % sha)
        index = str(Path(self.tmp.name) / ('index-' + name))
        environment = dict(os.environ, GIT_INDEX_FILE=index, **identity)
        gitDirectory = git(self.repository, 'rev-parse', '--absolute-git-dir')
        subprocess.run(['git', '--git-dir', gitDirectory, '--work-tree', str(tree), 'add', '-A', '.'], env=environment, check=True)
        treeSha = subprocess.run(['git', '--git-dir', gitDirectory, 'write-tree'], env=environment, check=True, capture_output=True, text=True).stdout.strip()
        reference = 'gate-logs/%s/20261009T000000Z/%s' % (sha[:12], name)
        git(self.repository, 'push', '-q', 'origin', '%s:refs/heads/%s' % (git(self.repository, 'commit-tree', treeSha, '-m', 'record'), reference))
        return reference

    def test_a_go_tests_only_record_lands_only_beside_a_record_of_the_other_stages(self):
        sha = self.change(self.main, 'other/b.go', 'package other\n\n// pooled\n', 'pooled')
        common = {'sha': sha, 'base': self.main, 'finished': True, 'skip': 0, 'packages': ['other']}
        tests = self.publish(sha, 'fast', dict(common, runner='pool', covers='go-tests', fail=0, **{'pass': 10}, uncached_tests=True, wall_seconds=100,
                                               steps_seconds={'tests': 90}, stages_exit={'tests': 0}, planned_stages=['tests']))
        stages = self.publish(sha, 'stages', dict(common, fail=0, **{'pass': 0}, build_ok=True, vet_ok=True, wall_seconds=50,
                                                  steps_seconds={stage: 5 for stage in ('build', 'vet', 'smoke', 'census')},
                                                  stages_exit={stage: 0 for stage in ('build', 'vet', 'smoke', 'census')},
                                                  planned_stages=['build', 'vet', 'smoke', 'census']))
        alone = self.push('--fast-gate', tests, sha, 'pooled')
        self.assertNotEqual(alone.returncode, 0)
        self.assertIn('go build or go vet failed', alone.stderr)
        self.assertLanded(self.push('--fast-gate', tests, '--also-gate', stages, sha, 'pooled'), self.main, sha)


if __name__ == '__main__':
    unittest.main()
