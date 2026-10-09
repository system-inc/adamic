#!/usr/bin/env python3
"""push-main's landings, end to end in a scratch repository (@system_adamic, Oct 9): each landing is one commit on
main, first parent the old main, second the landed sha, its tree exactly the landing's and its numbers as trailers,
and landings.py reads them back as the velocity table. Main moving by test-only commits doesn't spend a candidate's
gate unless they touch a package where the candidate changes code.

usage: python3 cloud/integration/push-main-landing-test.py
"""
import csv
import io
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
        # The scripts under test, beside a merge-back that does nothing.
        self.scripts = root / 'scripts'
        self.scripts.mkdir()
        for name in ('push-main.sh', 'landings.py'):
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
        return subprocess.run(['bash', str(self.scripts / 'push-main.sh')] + list(arguments), cwd=self.repository,
                              capture_output=True, text=True, env=dict(os.environ, **identity))

    def main_now(self):
        git(self.repository, 'fetch', '-q', 'origin')
        return git(self.repository, 'rev-parse', 'origin/main')

    def assertLanded(self, landed, old, sha):
        self.assertEqual(landed.returncode, 0, landed.stdout + landed.stderr)
        new = self.main_now()
        self.assertEqual(git(self.repository, 'rev-list', '--parents', '-n', '1', new).split()[1:], [old, sha])
        return new

    def test_a_gated_landing_is_one_commit_with_its_numbers_as_trailers(self):
        sha = self.change(self.main, 'code/a.go', 'package code\n// new\n', 'code change')
        new = self.assertLanded(self.push(sha, '12', '40', '0', '3', 'code/branch, at 1234'), self.main, sha)
        self.assertEqual(git(self.repository, 'rev-parse', new + '^{tree}'), git(self.repository, 'rev-parse', sha + '^{tree}'))
        trailers = git(self.repository, 'log', '-1', '--format=%(trailers:only,unfold)', new)
        for line in ('Old-main: ' + self.main, 'Landed-commits: 2', 'Gate-minutes: 12', 'Pass: 40', 'Fail: 0', 'Skip: 3', 'Branches: code/branch, at 1234'):
            self.assertIn(line, trailers)
        # Main moved once: nothing on top of the landing commit.
        self.assertEqual(git(self.repository, 'rev-list', '--count', sha + '..' + new), '1')
        rows = list(csv.DictReader(io.StringIO(subprocess.run(['python3', str(self.scripts / 'landings.py')], cwd=self.repository,
                                                              capture_output=True, text=True, check=True).stdout)))
        self.assertEqual([(row['new_main'], row['old_main'], row['gate_minutes'], row['branches_landed']) for row in rows],
                         [(new, self.main, '12', 'code/branch; at 1234')])

    def test_a_test_only_landing_is_one_commit_too(self):
        sha = self.change(self.main, 'code/a_test.go', 'package code\n', 'a test')
        new = self.assertLanded(self.push('--test-only', sha, 'a split'), self.main, sha)
        self.assertIn('Gate-minutes: 0', git(self.repository, 'log', '-1', '--format=%B', new))
        refused = self.push('--test-only', self.change(new, 'code/a.go', 'package code\n// x\n', 'code'), 'not a split')
        self.assertIn('not test-only', refused.stderr)

    def test_a_candidate_built_ahead_lands_over_the_landing_commit_below_it(self):
        lower = self.change(self.main, 'code/a.go', 'package code\n// lower\n', 'lower')
        upper = self.change(lower, 'other/b.go', 'package other\n// upper\n', 'upper')
        landed = self.assertLanded(self.push(lower, '1', '1', '0', '0', 'lower'), self.main, lower)
        top = self.assertLanded(self.push(upper, '1', '1', '0', '0', 'upper'), landed, upper)
        self.assertEqual(git(self.repository, 'rev-parse', top + '^{tree}'), git(self.repository, 'rev-parse', upper + '^{tree}'))

    def test_main_moving_by_tests_keeps_a_gate_outside_the_candidates_packages(self):
        candidate = self.change(self.main, 'code/a.go', 'package code\n// candidate\n', 'candidate')
        elsewhere = self.change(self.main, 'other/b_test.go', 'package other\n', 'a test elsewhere')
        moved = self.assertLanded(self.push('--test-only', elsewhere, 'elsewhere'), self.main, elsewhere)
        landed = self.assertLanded(self.push(candidate, '1', '1', '0', '0', 'candidate'), moved, candidate)
        self.assertEqual(git(self.repository, 'show', landed + ':other/b_test.go'), 'package other')
        # A test in the candidate's own package meets its new code, so the gate is spent.
        second = self.change(landed, 'code/a.go', 'package code\n// second\n', 'second')
        same = self.change(landed, 'code/a_test.go', 'package code\n', 'a test beside it')
        self.assertLanded(self.push('--test-only', same, 'beside'), landed, same)
        refused = self.push(second, '1', '1', '0', '0', 'second')
        self.assertIn('beyond record and test-only commits (code/a_test.go)', refused.stderr)


if __name__ == '__main__':
    unittest.main()
