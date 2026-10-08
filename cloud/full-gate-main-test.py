#!/usr/bin/env python3
"""The whole-gate loop's choices, against a real git origin: a record-only main is confirmed by its parent's green
whole gate and never run, a main with code is, and the star's requests run bottom first, skipping finished ones.
Run: python3 cloud/full-gate-main-test.py"""
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

script = Path(__file__).with_name('full-gate-main.sh')


def git(directory, *arguments):
    return subprocess.run(['git', '-C', str(directory), *arguments], check=True, capture_output=True, text=True).stdout.strip()


class FullGateLoopTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        root = Path(self.tmp.name)
        self.origin = root / 'origin.git'
        git(root, 'init', '-q', '--bare', str(self.origin))
        self.work = root / 'work'
        git(root, 'clone', '-q', str(self.origin), str(self.work))
        (self.work / 'cloud').mkdir()
        shutil.copy(script, self.work / 'cloud' / 'full-gate-main.sh')
        self.state = root / 'state'
        self.state.mkdir()
        self.code = self.commit({'compiler.go': 'package compiler\n'})
        self.records = self.commit({'documentation/velocity/landings.csv': 'row\n', 'stage3/progress.json': '{}\n'})
        self.changed = self.commit({'compiler.go': 'package compiler // changed\n'})
        git(self.work, 'push', '-q', 'origin', 'HEAD:refs/heads/main')

    def commit(self, files):
        for name, text in files.items():
            path = self.work / name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(text)
        git(self.work, 'add', '-A')
        git(self.work, '-c', 'user.name=t', '-c', 'user.email=t@t', 'commit', '-qm', 'c')
        return git(self.work, 'rev-parse', 'HEAD')

    def record(self, sha, status, finished=True, stamp='20261008T220000Z'):
        """A published whole-gate record of sha, as publish() leaves it."""
        tree = Path(self.tmp.name) / ('record-' + sha[:12] + stamp)
        tree.mkdir()
        (tree / 'status.txt').write_text(status + '\n')
        (tree / 'full.json').write_text(json.dumps({'finished': True} if finished else {}, indent=2) + '\n')
        index = str(tree) + '.index'
        environment = dict(os.environ, GIT_INDEX_FILE=index)
        gitDirectory = git(self.work, 'rev-parse', '--absolute-git-dir')
        subprocess.run(['git', '--git-dir', gitDirectory, '--work-tree', str(tree), 'add', '-A', '-f', '.'], env=environment, check=True)
        written = subprocess.run(['git', '--git-dir', gitDirectory, 'write-tree'], env=environment, check=True, capture_output=True, text=True).stdout.strip()
        commit = git(self.work, 'commit-tree', written, '-m', 'record')
        git(self.work, 'push', '-q', 'origin', '%s:refs/heads/gate-logs/%s/%s/full-main' % (commit, sha[:12], stamp))

    def call(self, expression):
        """Source the loop as a library and evaluate one expression; (exit code, stdout)."""
        command = 'set +e; ADAMIC_FULL_GATE_LIBRARY=1 source %s; %s' % (self.work / 'cloud' / 'full-gate-main.sh', expression)
        result = subprocess.run(['bash', '-c', command], capture_output=True, text=True,
                                env=dict(os.environ, ADAMIC_FULL_GATE_STATE=str(self.state), ADAMIC_FAST_GATE_WATCH_STATE=str(self.state)))
        return result.returncode, result.stdout.strip()

    def test_a_record_only_main_is_confirmed_by_its_parent_s_green_whole_gate(self):
        self.assertEqual(self.call('confirmedBy %s' % self.records), (1, ''), 'no green record yet')
        self.record(self.code, 'green: %s full gate in 2400 s' % self.code)
        self.assertEqual(self.call('confirmedBy %s' % self.records), (0, self.code))
        # A main that changed code past the green one is never confirmed by it.
        self.assertEqual(self.call('confirmedBy %s' % self.changed)[0], 1)

    def test_a_red_or_unfinished_record_confirms_nothing(self):
        self.record(self.code, 'red: %s full gate, first failure at tests' % self.code)
        self.assertEqual(self.call('confirmedBy %s' % self.records)[0], 1)
        self.record(self.code, 'green: %s full gate' % self.code, finished=False, stamp='20261008T230000Z')
        self.assertEqual(self.call('recordState %s' % self.code), (0, 'running'))
        self.assertEqual(self.call('confirmedBy %s' % self.records)[0], 1)

    def test_requests_run_bottom_first_skipping_finished_and_dropping_on_a_finished_record(self):
        red, void, fresh = self.code, self.records, self.changed
        self.record(red, 'red: %s full gate, first failure at census' % red)
        self.record(void, 'void: %s full gate, box lacks a declared tool' % void)
        (self.state / 'requests').write_text('%s\n%s\n%s\n' % (red, void, fresh))
        self.assertEqual(self.call('nextRequest'), (0, void), 'a void runs again; a finished red does not')
        self.call('dropRequest %s' % void)
        self.call('dropRequest %s' % red)
        self.assertEqual((self.state / 'requests').read_text(), '%s\n%s\n' % (void, fresh), 'only a finished record drops its line')

    def test_two_loops_never_take_the_same_request(self):
        first, second = self.code, self.changed
        (self.state / 'requests').write_text('%s\n%s\n' % (first, second))
        # Two loops on two boxes, in one shell each: the first claims the bottom line, the second the next.
        home = self.call('box=home; nextRequest; sleep 3')
        self.assertEqual(home, (0, first))
        self.assertEqual(self.call('box=threadripper; nextRequest'), (0, first), "a dead loop's claim is taken over")
        holder = (self.state / 'claims' / first / 'holder').read_text().split()
        self.assertEqual(holder[0], 'threadripper')
        # While the holder lives, the other loop takes the next line.
        live = subprocess.Popen(['bash', '-c', 'set +e; ADAMIC_FULL_GATE_LIBRARY=1 source %s; box=home; nextRequest; sleep 5' % (self.work / 'cloud' / 'full-gate-main.sh')],
                                stdout=subprocess.PIPE, text=True, env=dict(os.environ, ADAMIC_FULL_GATE_STATE=str(self.state), ADAMIC_FAST_GATE_WATCH_STATE=str(self.state)))
        self.addCleanup(live.kill)
        for _ in range(100):
            if (self.state / 'claims' / first / 'holder').read_text().split()[0] == 'home':
                break
            subprocess.run(['sleep', '0.1'])
        self.assertEqual(self.call('box=threadripper; nextRequest'), (0, second))
        self.call('release %s' % second)
        self.assertFalse((self.state / 'claims' / second).exists())

    def test_a_whole_gate_takes_every_line_of_its_box_and_gives_them_back(self):
        bin = Path(self.tmp.name) / 'bin'
        bin.mkdir()
        (bin / 'ssh').write_text('#!/bin/bash\nexit 0\n')
        (bin / 'ssh').chmod(0o755)
        (self.state / 'slots').write_text('threadripper B\nthreadripper S\nthreadripper S\nworkshop B\n')
        command = 'set +e; ADAMIC_FULL_GATE_LIBRARY=1 source %s; box=threadripper; %s'
        environment = dict(os.environ, PATH=str(bin) + ':' + os.environ['PATH'], ADAMIC_FULL_GATE_STATE=str(self.state), ADAMIC_FAST_GATE_WATCH_STATE=str(self.state))
        subprocess.run(['bash', '-c', command % (self.work / 'cloud' / 'full-gate-main.sh', 'reclaim')], env=environment, check=True)
        self.assertEqual((self.state / 'slots').read_text(), 'workshop B\n')
        subprocess.run(['bash', '-c', command % (self.work / 'cloud' / 'full-gate-main.sh', 'lend; lend')], env=environment, check=True)
        self.assertEqual((self.state / 'slots').read_text(), 'workshop B\nthreadripper B\nthreadripper S\nthreadripper S\n')

    def test_record_paths_are_push_main_s_three(self):
        self.assertEqual(self.call('recordOnly %s %s' % (self.code, self.records))[0], 0)
        self.assertEqual(self.call('recordOnly %s %s' % (self.records, self.changed))[0], 1)


if __name__ == '__main__':
    unittest.main()
