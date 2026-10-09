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
        shutil.copy(script.with_name('fast-gate-classify.sh'), self.work / 'cloud' / 'fast-gate-classify.sh')
        self.state = root / 'state'
        self.state.mkdir()
        self.code = self.commit({'compiler.go': 'package compiler\n'})
        self.records = self.commit({'stage3/meter/runs/r.json': '{}\n', 'stage3/progress.json': '{}\n'})
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

    def record(self, sha, status, finished=True, stamp='20261008T220000Z', runner=None):
        """A published whole-gate record of sha, as publish() leaves it."""
        tree = Path(self.tmp.name) / ('record-' + sha[:12] + stamp)
        tree.mkdir()
        (tree / 'status.txt').write_text(status + '\n')
        content = {'finished': True} if finished else {}
        if runner:
            content['runner'] = runner
        (tree / 'full.json').write_text(json.dumps(content, indent=2) + '\n')
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
                                env=dict(os.environ, ADAMIC_FULL_GATE_STATE=str(self.state), ADAMIC_FAST_GATE_WATCH_STATE=str(self.state),
                                         ADAMIC_FULL_GATE_BOXES=str(self.state / 'boxes')))
        return result.returncode, result.stdout.strip()

    def test_a_record_only_main_is_confirmed_by_its_parent_s_green_whole_gate(self):
        self.assertEqual(self.call('confirmedBy %s' % self.records), (1, ''), 'no green record yet')
        self.record(self.code, 'green: %s full gate in 2400 s' % self.code)
        self.assertEqual(self.call('confirmedBy %s' % self.records), (0, self.code))
        # A main that changed code past the green one is never confirmed by it.
        self.assertEqual(self.call('confirmedBy %s' % self.changed)[0], 1)

    def test_a_landing_merge_with_the_gated_tree_is_confirmed_by_the_landed_sha_s_green_record(self):
        # Integration, Oct 9 (#8teb9b7): a landing is one merge commit, first parent old main, second the landed sha,
        # tree exactly the gated tree. Its green record confirms the merge; nothing is re-gated.
        oldMain = self.code
        git(self.work, 'checkout', '-q', '-b', 'side', oldMain)
        landed = self.commit({'feature.go': 'package compiler // feature\n'})
        git(self.work, 'checkout', '-q', '--detach', oldMain)
        tree = git(self.work, 'rev-parse', landed + '^{tree}')
        merge = git(self.work, 'commit-tree', tree, '-p', oldMain, '-p', landed, '-m', 'Land feature', '-m', 'Old-main: ' + oldMain)
        git(self.work, 'push', '-q', 'origin', merge + ':refs/heads/landing')
        self.assertEqual(self.call('confirmedBy %s' % merge), (1, ''), 'no green record of the landed sha yet')
        self.record(landed, 'green: %s full gate in 2400 s' % landed)
        self.assertEqual(self.call('confirmedBy %s' % merge), (0, landed))
        # A merge whose tree adds a test the landed sha never gated isn't confirmed by it.
        git(self.work, 'checkout', '-q', '--detach', merge)
        extra = self.commit({'other_test.go': 'package compiler\n'})
        git(self.work, 'push', '-q', 'origin', extra + ':refs/heads/landing-extra')
        self.assertEqual(self.call('confirmedBy %s' % extra)[0], 1)

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

    def test_a_request_that_left_the_file_is_dropped_and_stopped_by_its_exact_sha(self):
        (self.state / 'requests').write_text(self.code + '\n')
        self.assertEqual(self.call('requestDropped %s' % self.code)[0], 1)
        self.assertEqual(self.call('requestDropped %s' % self.changed)[0], 0)
        bin = Path(self.tmp.name) / 'bin'
        bin.mkdir()
        (bin / 'ssh').write_text('#!/bin/bash\nprintf "%%s\\n" "$*" >> %s\n' % (Path(self.tmp.name) / 'ssh-calls'))
        (bin / 'ssh').chmod(0o755)
        environment = dict(os.environ, PATH=str(bin) + ':' + os.environ['PATH'], ADAMIC_FULL_GATE_STATE=str(self.state), ADAMIC_FAST_GATE_WATCH_STATE=str(self.state))
        for sha, code in (('', 1), ('abc', 1), (self.changed, 0)):
            result = subprocess.run(['bash', '-c', 'set +e; ADAMIC_FULL_GATE_LIBRARY=1 source %s; box=threadripper; stopRemote "%s"' % (self.work / 'cloud' / 'full-gate-main.sh', sha)], env=environment)
            self.assertEqual(result.returncode, code, sha)
        self.assertEqual((Path(self.tmp.name) / 'ssh-calls').read_text(), "threadripper pkill -TERM -f 'run.py .*--sha %s'\n" % self.changed)

    def test_a_red_or_running_pre_gate_keeps_a_candidate_off_the_box(self):
        red, running, green = self.code, self.records, self.changed
        (self.state / 'requests').write_text('%s\n%s\n%s\n' % (red, running, green))
        (self.state / 'pregate').mkdir()
        (self.state / 'pregate' / red).write_text('red: 3 units failed\n')
        (self.state / 'pregate' / running).write_text('running\n')
        (self.state / 'pregate' / green).write_text('green\n')
        self.assertEqual(self.call('nextRequest'), (0, green))
        # No pre-gate file at all: eligible.
        (self.state / 'pregate' / red).unlink()
        self.call('release %s' % green)
        self.assertEqual(self.call('nextRequest'), (0, red))

    def test_the_pool_s_record_counts_only_after_promotion_and_then_a_fifth_still_take_a_box(self):
        shas = [self.code, self.records, self.changed]
        due = [sha for sha in shas if int(sha[:8], 16) % 5 == 0]
        undue = [sha for sha in shas if int(sha[:8], 16) % 5 != 0]
        if not undue:
            self.skipTest('no commit here falls outside the spot-check')
        sha = undue[0]
        self.record(sha, 'green: %s full gate on the pool' % sha, runner='pool')
        (self.state / 'requests').write_text(sha + '\n')
        self.assertEqual(self.call('recordState %s box' % sha), (0, 'none'))
        self.assertEqual(self.call('recordState %s pool' % sha), (0, 'green'))
        # Parity phase: the box runs it beside the pool.
        self.assertEqual(self.call('nextRequest'), (0, sha))
        self.call('release %s' % sha)
        # Promoted: the pool's record stands, unless the sha is due a spot-check.
        (self.state / 'pool-promoted').write_text('parity proven\n')
        self.assertEqual(self.call('nextRequest')[0], 1)
        if due:
            self.record(due[0], 'green: %s full gate on the pool' % due[0], runner='pool', stamp='20261008T221000Z')
            (self.state / 'requests').write_text(due[0] + '\n')
            self.assertEqual(self.call('nextRequest'), (0, due[0]))

    def test_after_promotion_the_pool_gates_main_first_and_a_box_only_spot_checks(self):
        # Kirk, Oct 9: "we always need full gates to be running on loom". Before promotion the box gates every main.
        undue = due = None
        for attempt in range(60):
            sha = self.commit({'compiler.go': 'package compiler // %d\n' % attempt})
            if int(sha[:8], 16) % 5 == 0:
                due = due or sha
            else:
                undue = undue or sha
            if undue and due:
                break
        self.assertEqual(self.call('mainTurn %s' % undue)[0], 0)
        (self.state / 'pool-promoted').write_text('parity proven\n')
        # Promoted, the pool goes first: a main it hasn't taken waits its grace, then a box takes it.
        self.assertEqual(self.call('mainTurn %s' % undue)[0], 1)
        rows = (self.state / 'mains-seen').read_text().split()
        (self.state / 'mains-seen').write_text('%s %d\n' % (undue, int(rows[1]) - 601))
        self.assertEqual(self.call('mainTurn %s' % undue)[0], 0)
        # Taken by the pool, it's the pool's.
        self.record(undue, 'running: full gate of %s on the pool' % undue, finished=False, runner='pool')
        self.assertEqual(self.call('mainTurn %s' % undue)[0], 1)
        # A sha the boxes spot-check runs on a box beside the pool's green.
        self.record(due, 'green: %s full gate on the pool' % due, runner='pool', stamp='20261008T221500Z')
        self.assertEqual(self.call('mainTurn %s' % due)[0], 0)
        # A box record of its own ends a main's turn either way.
        self.record(due, 'green: %s full gate' % due, stamp='20261008T223000Z')
        self.assertEqual(self.call('mainTurn %s' % due)[0], 1)

    def test_one_whole_gate_per_box_ever(self):
        # Oct 9 00:13Z: the requests loop took a second whole gate onto the Threadripper beside a hand-started one.
        script = self.work / 'cloud' / 'full-gate-main.sh'
        environment = dict(os.environ, ADAMIC_FULL_GATE_STATE=str(self.state), ADAMIC_FAST_GATE_WATCH_STATE=str(self.state),
                           ADAMIC_FULL_GATE_BOXES=str(self.state / 'boxes'))
        holder = subprocess.Popen(['bash', '-c', 'set +e; ADAMIC_FULL_GATE_LIBRARY=1 source %s; box=threadripper; claimBox %s; sleep 30' % (script, self.code)],
                                  env=environment)
        self.addCleanup(holder.kill)
        for _ in range(100):
            if (self.state / 'boxes' / 'threadripper' / 'holder').exists():
                break
            subprocess.run(['sleep', '0.1'])
        self.assertEqual(self.call('box=threadripper; claimBox %s' % self.changed)[0], 1, 'a held box refuses a second gate')
        self.assertEqual(self.call('box=threadripper; boxFree')[0], 1)
        self.assertEqual(self.call('box=home; claimBox %s' % self.changed)[0], 0, 'another box is free')
        holder.kill()
        holder.wait()
        self.assertEqual(self.call('box=threadripper; boxFree')[0], 0, "a dead holder's claim is free")
        self.assertEqual(self.call('box=threadripper; claimBox %s; releaseBox' % self.changed)[0], 0)
        self.assertFalse((self.state / 'boxes' / 'threadripper').exists())

    def runRequest(self, sha):
        """runOnBox of a request, the box a directory on this machine and integration's inbox a file."""
        root = Path(self.tmp.name)
        bin = root / 'bin'
        bin.mkdir(exist_ok=True)
        box = root / 'box'
        box.mkdir(exist_ok=True)
        # The run itself: records what it gated and finishes green at once.
        (bin / 'ssh').write_text("""#!/bin/bash
shift
case "$*" in
  *pkill*) exit 0 ;;
  'bash -s -- '*)
    cat > /dev/null
    echo "$4" > "$TEST_ROOT/gated"
    mkdir -p "$TEST_ROOT/box/$6"
    echo "green: $4 full gate in 1.0 s" > "$TEST_ROOT/box/$6/status.txt"
    echo '{"finished": true}' > "$TEST_ROOT/box/$6/full.json" ;;
  *) HOME="$TEST_ROOT/box" bash -c "$*" ;;
esac
""")
        (bin / 'rsync').write_text('#!/bin/bash\nsource=${@: -2:1}\nmkdir -p "${@: -1}"\ncp -R "$TEST_ROOT/box/${source#*:}." "${@: -1}"\n')
        (bin / 'sleep').write_text('#!/bin/bash\n')
        (bin / 'ahra').write_text('#!/bin/bash\nprintf "%s\\n" "$*" >> "$TEST_ROOT/sent"\n')
        for name in ('ssh', 'rsync', 'sleep', 'ahra'):
            (bin / name).chmod(0o755)
        command = 'set +e; ADAMIC_FULL_GATE_LIBRARY=1 source %s; box=home; heartbeat() { :; }; runOnBox %s request'
        environment = dict(os.environ, PATH=str(bin) + ':' + os.environ['PATH'], TEST_ROOT=str(root), ADAMIC_FULL_GATE_STATE=str(self.state),
                           ADAMIC_FAST_GATE_WATCH_STATE=str(self.state), ADAMIC_FULL_GATE_BOXES=str(self.state / 'boxes'))
        return subprocess.run(['bash', '-c', command % (self.work / 'cloud' / 'full-gate-main.sh', sha)], env=environment, capture_output=True, text=True, timeout=60)

    def published(self, sha):
        refs = git(self.origin, 'for-each-ref', '--format=%(refname)', 'refs/heads/gate-logs/%s/' % sha[:12]).split()
        self.assertEqual(len(refs), 1, refs)
        return git(self.origin, 'show', '%s:status.txt' % refs[0])

    def request(self, files):
        """The star's candidate, forked from main's parent, so main's own change is one it doesn't hold."""
        git(self.work, 'checkout', '-q', '--detach', self.records)
        sha = self.commit(files)
        git(self.work, 'push', '-q', 'origin', '%s:refs/heads/cloud/land-request' % sha)
        git(self.work, 'checkout', '-q', '--detach', self.changed)
        return sha

    def test_a_request_is_gated_merged_onto_main_s_tip(self):
        # #11ymb02: a candidate forked before main's change gates with it, and its record names both.
        sha = self.request({'feature.go': 'package feature\n'})
        result = self.runRequest(sha)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        gated = (Path(self.tmp.name) / 'gated').read_text().strip()
        self.assertEqual(git(self.work, 'rev-list', '--parents', '-n', '1', gated).split()[1:], [self.changed, sha])
        self.assertEqual(git(self.work, 'show', '%s:compiler.go' % gated), 'package compiler // changed')
        status = self.published(sha)
        self.assertEqual(status, 'green: %s full gate in 1.0 s (gated merged onto main %s as %s)' % (sha, self.changed[:12], gated))
        self.assertEqual(self.published(gated), status, 'push-main lands the merge by its own record')

    def test_a_conflicting_request_is_red_at_merge_without_a_run(self):
        sha = self.request({'compiler.go': 'package compiler // the request\'s own\n'})
        result = self.runRequest(sha)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertFalse((Path(self.tmp.name) / 'gated').exists(), 'a conflict took the box')
        self.assertTrue(self.published(sha).startswith('red: %s full gate, first failure at merge' % sha), self.published(sha))
        self.assertIn('compiler.go', self.published(sha))
        self.assertIn('system_adamic_integration', (Path(self.tmp.name) / 'sent').read_text())

    def test_record_paths_are_push_main_s_three(self):
        self.assertEqual(self.call('recordOnly %s %s' % (self.code, self.records))[0], 0)
        self.assertEqual(self.call('recordOnly %s %s' % (self.records, self.changed))[0], 1)


if __name__ == '__main__':
    unittest.main()
