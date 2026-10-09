#!/usr/bin/env python3
"""gate-lane.py against planted records in a scratch repository: it picks a sha's newest finished record set, hands it
to push-main with the candidate's infra names, posts a landing and drops the line, posts a refusal once and keeps it,
and waits on a hold (#43kay4z).

usage: python3 cloud/integration/gate-lane-test.py
"""
import os, subprocess, tempfile, unittest
from pathlib import Path

script = Path(__file__).with_name('gate-lane.py')


def git(directory, *arguments):
    return subprocess.run(['git', '-C', str(directory), '-c', 'user.name=t', '-c', 'user.email=t@t', *arguments],
                          check=True, capture_output=True, text=True).stdout.strip()


class GateLaneTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        origin = self.root / 'origin.git'
        git(self.root, 'init', '-q', '--bare', str(origin))
        self.repository = self.root / 'repository'
        git(self.root, 'init', '-q', str(self.repository))
        git(self.repository, 'remote', 'add', 'origin', str(origin))
        (self.repository / 'a.txt').write_text('a\n')
        git(self.repository, 'add', '.')
        git(self.repository, 'commit', '-qm', 'candidate')
        self.sha = git(self.repository, 'rev-parse', 'HEAD')
        self.calls = self.root / 'calls'
        self.posts = self.root / 'posts'
        self.candidates = self.root / 'candidates'

    def record(self, stamp, kind, status, sha=None):
        sha = sha or self.sha
        tree = self.root / ('record-%s-%s-%s' % (sha[:8], stamp, kind))
        tree.mkdir()
        (tree / 'status.txt').write_text(status + '\n')
        index = str(tree) + '.index'
        environment = dict(os.environ, GIT_INDEX_FILE=index)
        gitDirectory = git(self.repository, 'rev-parse', '--absolute-git-dir')
        subprocess.run(['git', '--git-dir', gitDirectory, '--work-tree', str(tree), 'add', '-A', '.'], env=environment, check=True)
        written = subprocess.run(['git', '--git-dir', gitDirectory, 'write-tree'], env=environment, check=True, capture_output=True, text=True).stdout.strip()
        commit = git(self.repository, 'commit-tree', written, '-m', 'record')
        git(self.repository, 'push', '-q', 'origin', '%s:refs/heads/gate-logs/%s/%s/%s' % (commit, sha[:12], stamp, kind))

    def gateMerge(self):
        """Main moves past the candidate, and Loom gates the candidate as its merge onto main (main first, candidate second)."""
        git(self.repository, 'checkout', '-q', '-b', 'main-moved', self.sha + '~0')
        (self.repository / 'b.txt').write_text('b\n')
        git(self.repository, 'add', '.')
        git(self.repository, 'commit', '-qm', 'main moved')
        main = git(self.repository, 'rev-parse', 'HEAD')
        (self.repository / 'c.txt').write_text('c\n')
        git(self.repository, 'add', '.')
        git(self.repository, 'commit', '-qm', 'unrelated')
        unrelated = git(self.repository, 'rev-parse', 'HEAD')
        merge = git(self.repository, 'commit-tree', main + '^{tree}', '-p', main, '-p', self.sha, '-m', 'Gate merge')
        other = git(self.repository, 'commit-tree', main + '^{tree}', '-p', main, '-p', unrelated, '-m', 'Gate merge of another')
        for sha in (merge, other):
            git(self.repository, 'push', '-q', 'origin', '%s:refs/gate-merges/%s' % (sha, sha))
        return merge, other

    def run_lane(self, exitCode):
        fake = self.root / 'push-main.sh'
        fake.write_text('echo "$@" >> %s\n[ %d = 0 ] && echo "Pushed main aaaa..bbbb"\n[ %d = 1 ] && echo "refused: a real reason" >&2\nexit %d\n'
                        % (self.calls, exitCode, exitCode, exitCode))
        poster = self.root / 'ahra'
        # Like ahra: the comment comes from the file after --text-file, and success says "Comment added".
        poster.write_text('#!/usr/bin/env bash\nfile=""; while [ $# -gt 0 ]; do [ "$1" = --text-file ] && file=$2; shift; done\n'
                          'cat "$file" >> %s; echo >> %s; echo "Comment added"\n' % (self.posts, self.posts))
        poster.chmod(0o755)
        return subprocess.run(['python3', str(script)], cwd=self.repository, capture_output=True, text=True,
                              env=dict(os.environ, GATE_LANE_CANDIDATES=str(self.candidates), GATE_LANE_STATE=str(self.root / 'state'),
                                       GATE_LANE_PUSH_MAIN=str(fake), GATE_LANE_AHRA=str(poster)))

    def test_it_lands_the_newest_finished_record_and_drops_the_line(self):
        self.record('20261009T100000Z', 'fast', 'red: old')
        self.record('20261009T110000Z', 'fast', 'red: newer')
        self.record('20261009T110000Z', 'fast-phases', 'green: phases')
        self.record('20261009T120000Z', 'fast', 'running: newest, unfinished')
        self.candidates.write_text('cloud/land-x\t%s\ttask1\tpkg TestKilled=#t4b9j71\n' % self.sha)
        self.run_lane(0)
        call = self.calls.read_text()
        self.assertIn('--fast-gate gate-logs/%s/20261009T110000Z/fast --also-gate gate-logs/%s/20261009T110000Z/fast-phases' % (self.sha[:12], self.sha[:12]), call)
        self.assertIn('--infra-red pkg TestKilled=#t4b9j71', call)
        self.assertIn('Gate lane landed', self.posts.read_text())
        self.assertNotIn(self.sha, self.candidates.read_text())

    def test_a_gate_merge_of_the_candidate_lands_on_its_newer_record(self):
        merge, other = self.gateMerge()
        self.record('20261009T100000Z', 'fast', 'red: the candidate alone, older')
        self.record('20261009T110000Z', 'fast', 'green: the gate merge', merge)
        self.record('20261009T120000Z', 'fast', 'green: another candidate\'s merge, newest', other)
        self.candidates.write_text('cloud/land-m\t%s\ttask3\t\n' % self.sha)
        self.run_lane(0)
        call = self.calls.read_text()
        self.assertIn('--fast-gate gate-logs/%s/20261009T110000Z/fast %s' % (merge[:12], merge), call)
        self.assertNotIn(other[:12], call)
        self.assertIn('Gate lane landed %s (gate merge of %s)' % (merge[:8], self.sha[:8]), self.posts.read_text())
        self.assertNotIn(self.sha, self.candidates.read_text())

    def test_the_candidates_own_newer_record_beats_an_older_gate_merge(self):
        merge, _ = self.gateMerge()
        self.record('20261009T100000Z', 'fast', 'green: the gate merge, older', merge)
        self.record('20261009T110000Z', 'fast', 'green: the candidate itself')
        self.candidates.write_text('cloud/land-n\t%s\ttask4\t\n' % self.sha)
        self.run_lane(0)
        self.assertIn('--fast-gate gate-logs/%s/20261009T110000Z/fast %s' % (self.sha[:12], self.sha), self.calls.read_text())

    def test_a_refusal_is_posted_once_and_a_hold_waits(self):
        self.record('20261009T110000Z', 'full-main', 'red: whole')
        self.candidates.write_text('cloud/land-y\t%s\ttask2\t\n' % self.sha)
        self.run_lane(1)
        self.run_lane(1)
        self.assertEqual(self.posts.read_text().count("doesn't land: refused: a real reason"), 1)
        self.assertIn('--full-gate gate-logs/%s/20261009T110000Z/full-main' % self.sha[:12], self.calls.read_text())
        self.assertIn(self.sha, self.candidates.read_text())
        self.posts.write_text('')
        self.run_lane(3)
        self.assertEqual(self.posts.read_text(), '')
        self.assertIn(self.sha, self.candidates.read_text())


if __name__ == '__main__':
    unittest.main()
