#!/usr/bin/env python3
"""cloud/prune-gate-merges.sh deletes gate merges older than its window from origin and keeps the rest (#6f3b76x)."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import time
import unittest

ROOT = Path(__file__).resolve().parents[2]


def git(directory, *arguments, env=None):
    return subprocess.run(['git', '-C', str(directory), *arguments], check=True, capture_output=True, text=True, env=env).stdout.strip()


class PruneTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        root = Path(self.tmp.name)
        self.origin = root / 'origin.git'
        subprocess.run(['git', 'init', '-q', '--bare', str(self.origin)], check=True)
        self.here = root / 'here'
        subprocess.run(['git', 'clone', '-q', str(self.origin), str(self.here)], check=True, capture_output=True)
        (self.here / 'cloud').mkdir()
        shutil.copy(ROOT / 'cloud' / 'prune-gate-merges.sh', self.here / 'cloud')
        (self.here / 'a.go').write_text('package a\n')
        git(self.here, 'add', '-A')
        git(self.here, '-c', 'user.name=t', '-c', 'user.email=t@t', 'commit', '-qm', 'base')
        git(self.here, 'push', '-q', 'origin', 'HEAD:refs/heads/main')
        self.merges = {}
        for name, age in (('old', 9), ('fresh', 1)):
            date = '@%d +0000' % (time.time() - age * 86400)
            env = dict(os.environ, GIT_AUTHOR_NAME='g', GIT_AUTHOR_EMAIL='g@g', GIT_AUTHOR_DATE=date,
                       GIT_COMMITTER_NAME='g', GIT_COMMITTER_EMAIL='g@g', GIT_COMMITTER_DATE=date)
            sha = git(self.here, 'commit-tree', 'HEAD^{tree}', '-p', 'HEAD', '-m', name, env=env)
            git(self.here, 'push', '-q', 'origin', '%s:refs/gate-merges/%s' % (sha, sha))
            self.merges[name] = sha

    def remaining(self):
        return git(self.origin, 'for-each-ref', '--format=%(refname)', 'refs/gate-merges/').split()

    def test_a_merge_older_than_the_window_goes_and_a_fresh_one_stays(self):
        result = subprocess.run(['bash', str(self.here / 'cloud' / 'prune-gate-merges.sh')], capture_output=True, text=True, timeout=60)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.remaining(), ['refs/gate-merges/' + self.merges['fresh']])
        self.assertIn('deleted 1', result.stdout)

    def test_a_dry_run_deletes_nothing(self):
        result = subprocess.run(['bash', str(self.here / 'cloud' / 'prune-gate-merges.sh'), '--dry-run'], capture_output=True, text=True, timeout=60)
        self.assertIn('refs/gate-merges/' + self.merges['old'], result.stdout)
        self.assertEqual(len(self.remaining()), 2)


if __name__ == '__main__':
    unittest.main()
