#!/usr/bin/env python3
"""merge-queue.py stacks against a scratch origin (#dyre8as): members in tier order become main+A, main+A+B, ... as
merge chains on main's tip, a conflicting member is left out with its files, a member already on main is skipped,
depth caps the stacks, an unchanged run pushes nothing, and a moved main restacks and withdraws the old stacks.

usage: python3 cloud/integration/merge-queue-test.py
"""
import json, os, subprocess, tempfile, unittest
from pathlib import Path

script = Path(os.environ.get('MERGE_QUEUE_SCRIPT', Path(__file__).with_name('merge-queue.py')))


def git(directory, *arguments):
    return subprocess.run(['git', '-C', str(directory), '-c', 'user.name=t', '-c', 'user.email=t@t', *arguments],
                          check=True, capture_output=True, text=True, timeout=25).stdout.strip()


class StackTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.origin = self.root / 'origin.git'
        git(self.root, 'init', '-q', '--bare', '-b', 'main', str(self.origin))
        self.repository = self.root / 'repository'
        git(self.root, 'init', '-q', '-b', 'main', str(self.repository))
        git(self.repository, 'remote', 'add', 'origin', str(self.origin))
        self.commit({'shared.txt': 'base\n'}, 'main')
        git(self.repository, 'push', '-q', 'origin', 'main')
        self.main = git(self.repository, 'rev-parse', 'HEAD')
        self.candidates = self.root / 'candidates'
        self.withdrawn = self.root / 'withdrawn'
        withdraw = self.root / 'withdraw.sh'
        withdraw.write_text('#!/usr/bin/env bash\necho "$1" >> %s\n' % self.withdrawn)
        withdraw.chmod(0o755)
        self.environment = dict(os.environ, QUEUE_CANDIDATES=str(self.candidates), QUEUE_STATE=str(self.root / 'state'),
                                QUEUE_LOCK=str(self.root / 'lock'), QUEUE_WITHDRAW=str(withdraw))

    def commit(self, files, message, parent=None):
        if parent:
            git(self.repository, 'checkout', '-q', '--detach', parent)
        for name, text in files.items():
            (self.repository / name).write_text(text)
        git(self.repository, 'add', '.')
        git(self.repository, 'commit', '-qm', message)
        return git(self.repository, 'rev-parse', 'HEAD')

    def cut(self, name, files, tier):
        """A candidate like cut.sh makes: a change on main, then an empty commit carrying Gate-tier, on its branch."""
        change = self.commit(files, name, parent=self.main)
        cut = git(self.repository, 'commit-tree', change + '^{tree}', '-p', change, '-m', '%s\n\nGate-runs: deferred\nGate-tier: %d\n' % (name, tier))
        branch = 'cloud/land-stack-%s-%s' % (name, cut[:8])
        git(self.repository, 'push', '-q', 'origin', '%s:refs/heads/%s' % (cut, branch))
        return branch, cut

    def write(self, *lines):
        self.candidates.write_text('# header\n' + ''.join('%s\t%s\t%s\t\n' % (branch, sha, task) for branch, sha, task in lines))

    def run_queue(self, *arguments):
        result = subprocess.run(['python3', str(script), 'stacks', *arguments], cwd=self.repository, capture_output=True,
                                text=True, timeout=120, env=self.environment)
        self.assertEqual(result.returncode, 0, result.stderr)
        return result.stdout

    def state(self):
        return json.loads((self.root / 'state' / 'state.json').read_text())

    def remoteQueueBranches(self):
        return sorted(line.split('\t')[1] for line in git(self.repository, 'ls-remote', 'origin', 'refs/heads/cloud/land-queue-*').splitlines())

    def test_stacks_in_tier_order_skip_the_conflict_and_cap_depth(self):
        low = self.cut('low', {'low.txt': 'low\n'}, 20)
        clash = self.cut('clash', {'shared.txt': 'theirs\n'}, 30)
        first = self.cut('first', {'shared.txt': 'ours\n'}, 40)
        middle = self.cut('middle', {'middle.txt': 'middle\n'}, 30)
        extra = self.cut('extra', {'extra.txt': 'extra\n'}, 30)
        self.write((low[0], low[1], 'tlow'), (clash[0], clash[1], 'tclash'), (first[0], first[1], 'tfirst'),
                   (middle[0], middle[1], 'tmiddle'), (extra[0], extra[1], 'textra'))
        output = self.run_queue()
        self.assertIn('conflict %s %s with the prefix before it: shared.txt' % (clash[0], clash[1][:8]), output)
        stacks = self.state()['stacks']
        # Tier 40 first, then the 30s in file order without the clash; depth 3 leaves extra's 30 and low's 20 out.
        self.assertEqual([s['members'] for s in stacks], [[first[1]], [first[1], middle[1]], [first[1], middle[1], extra[1]]])
        self.assertEqual(self.state()['conflicts'][0]['files'], ['shared.txt'])
        previous = self.main
        for stack in stacks:
            parents = git(self.repository, 'log', '-1', '--format=%P', stack['sha']).split()
            self.assertEqual(parents, [previous, stack['members'][-1]])
            body = git(self.repository, 'log', '-1', '--format=%B', stack['sha'])
            self.assertIn('Gate-tier: 40', body)
            self.assertIn('Queue-main: %s' % self.main, body)
            previous = stack['sha']
        files = git(self.repository, 'ls-tree', '--name-only', stacks[-1]['sha']).split()
        self.assertEqual(sorted(files), ['extra.txt', 'middle.txt', 'shared.txt'])
        self.assertEqual(git(self.repository, 'show', '%s:shared.txt' % stacks[-1]['sha']), 'ours')
        self.assertEqual(self.remoteQueueBranches(), sorted('refs/heads/' + s['branch'] for s in stacks))

    def test_unchanged_run_pushes_nothing_and_a_moved_main_restacks(self):
        a = self.cut('a', {'a.txt': 'a\n'}, 30)
        b = self.cut('b', {'b.txt': 'b\n'}, 30)
        self.write((a[0], a[1], 'ta'), (b[0], b[1], 'tb'))
        self.run_queue()
        first = self.state()
        self.assertIn('unchanged: 2 stacks', self.run_queue())
        self.assertFalse(self.withdrawn.exists())
        moved = self.commit({'other.txt': 'other\n'}, 'main moves', parent=self.main)
        git(self.repository, 'push', '-q', 'origin', '%s:refs/heads/main' % moved)
        self.run_queue()
        second = self.state()
        self.assertEqual(second['main'], moved)
        self.assertEqual(git(self.repository, 'log', '-1', '--format=%P', second['stacks'][0]['sha']).split()[0], moved)
        self.assertEqual(sorted(self.withdrawn.read_text().split()), sorted(s['sha'] for s in first['stacks']))
        self.assertEqual(self.remoteQueueBranches(), sorted('refs/heads/' + s['branch'] for s in second['stacks']))

    def test_a_member_already_on_main_is_skipped(self):
        a = self.cut('a', {'a.txt': 'a\n'}, 30)
        git(self.repository, 'push', '-q', 'origin', '%s:refs/heads/main' % a[1])
        b = self.cut('b', {'b.txt': 'b\n'}, 30)
        self.write((a[0], a[1], 'ta'), (b[0], b[1], 'tb'))
        output = self.run_queue()
        self.assertIn('on main already: %s' % a[0], output)
        self.assertEqual([s['members'] for s in self.state()['stacks']], [[b[1]]])

    def test_dry_run_writes_and_pushes_nothing(self):
        a = self.cut('a', {'a.txt': 'a\n'}, 30)
        self.write((a[0], a[1], 'ta'))
        self.assertIn('stack 1 cloud/land-queue-1-', self.run_queue('--dry-run'))
        self.assertFalse((self.root / 'state' / 'state.json').exists())
        self.assertEqual(self.remoteQueueBranches(), [])


if __name__ == '__main__':
    unittest.main()
