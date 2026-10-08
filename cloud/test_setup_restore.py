#!/usr/bin/env python3
"""Exercise the actual shell exit handler against Git worktrees and submodules."""
from pathlib import Path
import subprocess
import tempfile
import unittest

SOURCE = Path(__file__).with_name('setup.sh')


class RestoreOnExit(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix='setup-restore-')
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.log = self.root / 'commands.log'
        self.repository = self.root / 'repository'
        self.repository.mkdir()
        self.git(self.repository, 'init')
        self.git(self.repository, 'commit', '--allow-empty', '-m', 'Initial')
        self.worktree = self.root / 'worktree'
        self.git(self.repository, 'worktree', 'add', '--detach', str(self.worktree))
        source = SOURCE.read_text()
        self.handler = source[source.index('# Restore workspace sums'):source.index('# End workspace sum exit handler.')]

    def git(self, directory, *args):
        with self.log.open('ab') as output:
            subprocess.run(['git', '-c', 'user.name=Setup proof', '-c',
                            'user.email=setup@example.invalid', '-C', str(directory), *args],
                           stdout=output, stderr=output, check=True)

    def run_handler(self, status=73, prefix='', mutant=False):
        handler = self.handler.replace('trap restoreOnExit EXIT', ':') if mutant else self.handler
        script = self.root / 'exit.sh'
        script.write_text('set -euo pipefail\nrepository=$1\n' + handler + '\n' + prefix +
                          '\necho "original step failure" >&2\nexit ' + str(status) + '\n')
        log = self.root / 'exit.log'
        with log.open('wb') as output:
            result = subprocess.run(['bash', str(script), str(self.worktree)],
                                    stdout=output, stderr=output)
        self.assertEqual(result.returncode, status, log.read_text())
        self.assertIn('original step failure', log.read_text())
        return log.read_text()

    def test_untracked_removed_on_failure_and_success(self):
        for status in [73, 0]:
            (self.worktree / 'go.work.sum').write_text('downloaded\n')
            self.run_handler(status)
            self.assertFalse((self.worktree / 'go.work.sum').exists())

    def test_tracked_submodule_restored_not_removed(self):
        child = self.root / 'child'
        child.mkdir()
        self.git(child, 'init')
        (child / 'go.work.sum').write_text('committed\n')
        self.git(child, 'add', 'go.work.sum')
        self.git(child, 'commit', '-m', 'Tracked sums')
        self.git(self.worktree, '-c', 'protocol.file.allow=always', 'submodule', 'add', str(child), 'cohere')
        self.git(self.worktree, 'commit', '-am', 'Submodule')
        (self.worktree / 'cohere/go.work.sum').write_text('changed\n')
        (self.worktree / 'go.work.sum').write_text('new\n')
        self.run_handler()
        self.assertEqual((self.worktree / 'cohere/go.work.sum').read_text(), 'committed\n')
        status = subprocess.check_output(['git', '-C', str(self.worktree), 'status', '--porcelain',
                                          '--ignore-submodules=none'])
        self.assertEqual(status, b'')

    def test_ignored_file_is_not_removed(self):
        (self.worktree / '.gitignore').write_text('go.work.sum\n')
        sums = self.worktree / 'go.work.sum'
        sums.write_text('ignored\n')
        self.run_handler()
        self.assertEqual(sums.read_text(), 'ignored\n')

    def test_git_failure_keeps_file_and_original_status(self):
        sums = self.worktree / 'go.work.sum'
        sums.write_text('keep\n')
        for status in [73, 0]:
            text = self.run_handler(status, 'timeout() { return 42; }')
            self.assertIn('restoration failed', text)
            self.assertIn('preserving setup exit status ' + str(status), text)
            self.assertEqual(sums.read_text(), 'keep\n')

    def test_staged_new_tracked_file_is_never_removed(self):
        sums = self.worktree / 'go.work.sum'
        sums.write_text('staged\n')
        self.git(self.worktree, 'add', 'go.work.sum')
        self.run_handler()
        self.assertEqual(sums.read_text(), 'staged\n')

    def test_missing_exit_trap_mutant_is_caught(self):
        sums = self.worktree / 'go.work.sum'
        sums.write_text('downloaded\n')
        self.run_handler(mutant=True)
        with self.assertRaises(AssertionError):
            self.assertFalse(sums.exists(), 'missing EXIT trap leaves the tree dirty')


if __name__ == '__main__':
    unittest.main()
