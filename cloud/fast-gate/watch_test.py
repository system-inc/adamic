#!/usr/bin/env python3
"""Exercise the real watcher with isolated state and no external side effects."""
import os
from pathlib import Path
import shutil
import signal
import subprocess
import tempfile
import time
import unittest

ROOT = Path(__file__).resolve().parents[2]
MAIN = 'a' * 40


class Watcher:
    def __init__(self, count=6, staleLock=False, canaryBox=None, mode='void', slots=None, boxSides=None):
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        self.repo = self.root / 'repo'
        cloud = self.repo / 'cloud'
        cloud.mkdir(parents=True)
        for name in ('fast-gate-watch.sh', 'fast-gate-classify.sh', 'first-step-branches.sh', 'stop-gate.sh'):
            shutil.copy(ROOT / 'cloud' / name, cloud / name)
        self.state = self.root / 'state'
        self.state.mkdir()
        if staleLock:
            # As a watcher killed mid-pass leaves it, with or without the holder line it records now.
            (self.state / 'slot-table.lock').mkdir()
            if isinstance(staleLock, str):
                (self.state / 'slot-table.lock' / 'holder').write_text(staleLock + ' 1000\n')
        self.bin = self.root / 'bin'
        self.bin.mkdir()
        self.put('clock', '1000')
        self.put('head', 'tools-one')
        self.put('mode', mode)
        self.put('initial', 'hold')
        self.put('canary', 'hold')
        # A tools commit's box-side fingerprint (git ls-tree), its own name unless given.
        for commit, side in (boxSides or {}).items():
            self.put('box-' + commit, side)
        self.tips = [(f'codex/test{i}', f'{i+1:012x}' + '0' * 28) for i in range(count)]
        self.put('tips', ''.join(f'{sha}\trefs/heads/{b}\n' for b, sha in self.tips))
        (self.state / 'seen').write_text(''.join(f'{b} {sha}\n' for b, sha in self.tips))
        (self.state / 'queue').write_text(''.join(f'S 900 {b} {sha}\n' for b, sha in self.tips))
        (self.state / 'slots').write_text(slots if slots is not None else ''.join(f'box{i % 2} S\n' for i in range(max(count, 1))))
        self.script(self.bin / 'git', '''case "$*" in
*rev-parse*) cat "$TEST_ROOT/head" ;;
*ls-tree*) echo "$(cat "$TEST_ROOT/box-$4" 2>/dev/null || echo "$4")" ;;
*'branch -r --contains'*) echo origin/devtools/fast-gate ;;
*'ls-remote'*refs/heads/main*) printf '%s\\trefs/heads/main\\n' aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa ;;
*'ls-remote'*refs/heads/codex*) cat "$TEST_ROOT/tips" ;;
*merge-base*) exit 1 ;;
esac
''')
        self.script(self.bin / 'ahra', '''if [ "$1" = tasks ]; then
  printf '%s\\n' "$*" >> "$TEST_ROOT/task-calls"
  if [ "$2" = ready ]; then cat "$TEST_ROOT/ready" 2>/dev/null; else cat "$TEST_ROOT/task-$3" 2>/dev/null; fi
  exit 0
fi
text=$(printf '%s' "$4" | tr '\\n' ' ')
printf '%s|%s|%s|%s\\n' "$3" "$text" "$5" "$6" >> "$TEST_ROOT/messages"
''')
        self.script(self.bin / 'date', '''case "$*" in
*+%s*) cat "$TEST_ROOT/clock" ;;
*) /bin/date -u +%H:%M:%S ;;
esac
''')
        self.script(self.bin / 'ssh', '''box=$1; shift
cat > "$TEST_ROOT/stop-command"
sha=${4%% *}
printf '%s %s\\n' "$box" "$*" >> "$TEST_ROOT/stops"
touch "$TEST_ROOT/stopped-$sha"
''')
        self.script(self.bin / 'sleep', '/bin/sleep 0.03\n')
        self.script(cloud / 'auto-area-merge.sh', '''printf '%s\\n' "$*" >> "$TEST_ROOT/merges"
''')
        (self.state / 'auto-area-merge').touch()
        self.script(cloud / 'fast-gate.sh', '''sha=$1; branch=$3
printf '%s %s %s %s\\n' "$branch" "$sha" "$5" "$ADAMIC_FAST_GATE_BOX" >> "$TEST_ROOT/starts"
[ "${6:-}" = --whole-box ] && printf '%s\\n' "$branch" >> "$TEST_ROOT/whole"
if [ "$branch" = canary/main ]; then
  control=canary
  if [ ! -f "$TEST_ROOT/initial-started" ]; then
    touch "$TEST_ROOT/initial-started"
    control=initial
  fi
  while [ "$(cat "$TEST_ROOT/$control")" = hold ]; do /bin/sleep 0.01; done
  mode=$(cat "$TEST_ROOT/$control")
else
  if [ -f "$TEST_ROOT/step" ]; then
    awk -v step="$(cat "$TEST_ROOT/step")" '{print $1 + step}' "$TEST_ROOT/clock" > "$TEST_ROOT/clock.next"
    mv "$TEST_ROOT/clock.next" "$TEST_ROOT/clock"
  fi
  [ ! -f "$TEST_ROOT/red-so-far" ] || echo 'FIRST FAILURE (tests, at 174.0 s):'
  while [ "$(cat "$TEST_ROOT/mode")" = hold ]; do
    if [ -f "$TEST_ROOT/stopped-$sha" ]; then
      /bin/sleep 0.15
      echo "red: $sha first failure at tests after 174.0 s"
      echo 'stopped: superseded by newer candidate'
      exit 1
    fi
    /bin/sleep 0.01
  done
  mode=$(cat "$TEST_ROOT/mode")
fi
if [ "$mode" = pass ]; then echo "green: $sha passed"
elif [ "$mode" = red ]; then echo "red: $sha failed"
else
  echo 'VOID FAILURE'; echo 'void: no result'
fi
''')
        extra = {}
        if canaryBox:
            # Staged tools: the good version is tools-zero, the checkout (head) is tools-one.
            (self.state / 'canary-box').write_text(canaryBox + '\n')
            (self.state / 'tools-good').write_text('tools-zero\n')
            good = self.root / 'good' / 'cloud'
            good.mkdir(parents=True)
            (good / 'fast-gate.sh').write_text((cloud / 'fast-gate.sh').read_text().replace('"$TEST_ROOT/starts"', '"$TEST_ROOT/good-starts"'))
            extra['ADAMIC_FAST_GATE_GOOD_TREE'] = str(self.root / 'good')
        env = dict(os.environ, **extra, PATH=str(self.bin) + ':' + os.environ['PATH'],
                   TEST_ROOT=str(self.root), ADAMIC_FAST_GATE_WATCH_STATE=str(self.state),
                   ADAMIC_FAST_GATE_AHRA_DIR=str(self.root))
        self.output = open(self.root / 'output', 'w')
        self.proc = subprocess.Popen([os.environ.get('WATCH_TEST_BASH', 'bash'), str(cloud / 'fast-gate-watch.sh')], env=env,
                                     stdout=self.output, stderr=self.output, start_new_session=True)

    def put(self, name, text):
        path = self.root / name
        tmp = self.root / (name + '.tmp')
        tmp.write_text(text)
        tmp.replace(path)

    @staticmethod
    def script(path, text):
        path.write_text('#!/usr/bin/env bash\n' + text)
        path.chmod(0o755)

    def read(self, name):
        p = self.root / name
        return p.read_text() if p.exists() else ''

    def wait(self, predicate, seconds=8):
        deadline = time.monotonic() + seconds
        while time.monotonic() < deadline:
            if predicate():
                return
            if self.proc.poll() is not None:
                break
            time.sleep(.02)
        raise AssertionError(self.read('output'))

    def close(self):
        os.killpg(self.proc.pid, signal.SIGTERM)
        self.proc.wait(timeout=5)
        self.output.close()
        # A gate child the signal reached mid-write can add a file while the tree is removed; retry briefly.
        for attempt in range(20):
            try:
                self.tmp.cleanup()
                return
            except OSError:
                time.sleep(.05)
        self.tmp.cleanup()


class WatchTests(unittest.TestCase):
    def start(self, count=6):
        w = Watcher(count)
        self.addCleanup(w.close)
        w.wait(lambda: len(w.read('starts').splitlines()) == 1)
        self.assertEqual(w.read('starts').split()[0], 'canary/main')
        time.sleep(.12)
        self.assertEqual(len(w.read('starts').splitlines()), 1)
        self.assertFalse((w.state / 'gated').read_text())
        return w

    def reservation(self, slots, branches, release=True):
        w = self.start(0)
        (w.state / 'slots').write_text(slots)
        w.tips = [(b, f'{i+1:040x}') for i, (b, c) in enumerate(branches)]
        w.put('tips', ''.join(f'{sha}\trefs/heads/{b}\n' for b, sha in w.tips))
        (w.state / 'seen').write_text(''.join(f'{b} {sha}\n' for b, sha in w.tips))
        (w.state / 'queue').write_text(''.join(f'{c} {900+i} {b} {sha}\n' for i, ((b, c), (_, sha)) in enumerate(zip(branches, w.tips))))
        w.put('mode', 'hold')
        if release:
            w.put('initial', 'pass')
        return w

    def stale_red_clean(self):
        w = self.reservation('server B cloud/land-area-next*\n', [('cloud/land-area-next-11', 'B')], release=False)
        w.put('initial', 'pass')
        w.wait(lambda: len(w.read('starts').splitlines()) == 2)
        old, newer = w.tips[0][1], 'b' * 40
        w.put('clock', '1100')
        w.tips.append(('cloud/land-area-next-12', newer))
        w.put('tips', ''.join(f'{sha}\trefs/heads/{b}\n' for b, sha in w.tips))
        w.wait(lambda: f'stopped cloud/land-area-next-11 {old}: superseded by cloud/land-area-next-12 {newer}, clean so far' in w.read('output'))
        self.assertEqual(len(w.read('stops').splitlines()), 1)
        return w

    def stale_red(self, reserved=True, red=True):
        slots = 'server B' + (' cloud/land-area-next*' if reserved else '') + '\n'
        w = self.reservation(slots, [('cloud/land-area-next-11', 'B')], release=False)
        if red:
            w.put('red-so-far', 'yes')
        w.put('initial', 'pass')
        w.wait(lambda: len(w.read('starts').splitlines()) == 2)
        old = w.tips[0][1]
        newer = 'b' * 40
        w.put('clock', '1100')
        w.tips.append(('cloud/land-area-next-12', newer))
        w.put('tips', ''.join(f'{sha}\trefs/heads/{b}\n' for b, sha in w.tips))
        w.wait(lambda: 'queued cloud/land-area-next-12' in w.read('output'))
        if reserved and red:
            w.wait(lambda: 'done cloud/land-area-next-11: red:' in w.read('output'))
            time.sleep(.15)
            self.assertEqual(len(w.read('stops').splitlines()), 1)
            self.assertIn(f'stopped cloud/land-area-next-11 {old}: superseded by cloud/land-area-next-12 {newer}, red since tests at 174.0 s', w.read('output'))
            self.assertNotIn('void cloud/land-area-next-11', w.read('output'))
            self.assertNotIn(old, (w.state / 'queue').read_text())
            self.assertIn(old, (w.state / 'gated').read_text())
            self.assertIn('pkill -TERM -f "run.py .*--sha ${sha}"', w.read('stop-command'))
            self.assertIn('${out}.stop-reason', w.read('stop-command'))
        else:
            time.sleep(.2)
            self.assertEqual(w.read('stops'), '')
            self.assertEqual(len(w.read('starts').splitlines()), 2)

    def test_red_reserved_gate_stops_once_and_reaps_red(self):
        self.stale_red()

    def test_reserved_gate_still_clean_is_stopped_too(self):
        # The newer candidate carries the older's commits and the fix (@system_adamic, Oct 8).
        w = self.stale_red_clean()

    def test_non_reserved_red_gate_keeps_running(self):
        self.stale_red(reserved=False)

    def test_reserved_priority_and_whole_box(self):
        w = self.reservation('server B cloud/land-area-next*\nserver S\nother S\n',
                             [('cloud/land-area-next1', 'B'), ('cloud/land-other', 'B'),
                              ('codex/small', 'S')])
        w.wait(lambda: len(w.read('starts').splitlines()) >= 3)
        starts = w.read('starts').splitlines()[1:]
        self.assertTrue(starts[0].startswith('cloud/land-area-next1 '))
        self.assertEqual(starts[0].split()[-1], 'server')
        self.assertEqual(len([x for x in starts if x.endswith(' server')]), 1)
        # A landing borrows the free small slot ahead of the small worker tip (integration, Oct 8).
        self.assertTrue(any(x.startswith('cloud/land-other ') and x.endswith(' S other') for x in starts), starts)
        # Only the reserved gate holding its box runs on every CPU of it.
        self.assertEqual(w.read('whole').splitlines(), ['cloud/land-area-next1'])
        time.sleep(.15)
        self.assertEqual(len(w.read('starts').splitlines()), 3)
        w.put('mode', 'pass')
        w.wait(lambda: 'done cloud/land-other' in w.read('output'))
        w.wait(lambda: any(x.startswith('codex/small ') for x in w.read('starts').splitlines()))

    def test_reserved_slot_rejects_other_tip_but_box_serves_small(self):
        w = self.reservation('server B cloud/land-area-next*\nserver S\n',
                             [('cloud/land-other', 'B'), ('codex/small', 'S')])
        w.wait(lambda: len(w.read('starts').splitlines()) == 2)
        # The reserved area slot refuses the other landing; it borrows the box's small slot, ahead of the worker.
        self.assertEqual(w.read('starts').splitlines()[1], 'cloud/land-other ' + w.tips[0][1] + ' S server')
        time.sleep(.15)
        self.assertEqual(len(w.read('starts').splitlines()), 2)
        w.put('mode', 'pass')
        w.wait(lambda: any(x.startswith('codex/small ') and x.endswith(' S server') for x in w.read('starts').splitlines()))

    def test_big_borrowing_obeys_globs_and_limit(self):
        w = self.reservation('server S cloud/land-area-next*\nother S\n',
                             [('cloud/land-other', 'B')], release=False)
        (w.state / 'big-per-box').write_text('other 0\n')
        w.put('initial', 'pass')
        w.wait(lambda: 'done canary:' in w.read('output'))
        time.sleep(.15)
        self.assertEqual(len(w.read('starts').splitlines()), 1)
        (w.state / 'big-per-box').write_text('other 1\n')
        w.wait(lambda: len(w.read('starts').splitlines()) == 2)
        self.assertTrue(w.read('starts').splitlines()[1].endswith(' S other'))

    def test_running_reservation_survives_table_change(self):
        w = self.reservation('server B cloud/land-area-next*\nserver S\n',
                             [('cloud/land-area-next1', 'B'), ('codex/small', 'S')])
        w.wait(lambda: len(w.read('starts').splitlines()) == 2)
        (w.state / 'slots').write_text('server B\nserver S\n')
        time.sleep(.15)
        self.assertEqual(len(w.read('starts').splitlines()), 2)
        w.put('mode', 'pass')
        w.wait(lambda: 'done codex/small' in w.read('output'))

    def test_waiting_reservation_drains_its_box(self):
        # Server's small slots must not refill under a reserved tip waiting for the whole box,
        # and the reserved big tip must not borrow one of them either.
        w = self.reservation('server B cloud/land-area-next*\nserver S\nserver S\nother S\n',
                             [('codex/first', 'S')])
        w.wait(lambda: len(w.read('starts').splitlines()) == 2)
        self.assertTrue(w.read('starts').splitlines()[1].endswith(' S server'))
        late = [('cloud/land-area-next1', 'B'), ('codex/second', 'S'), ('codex/third', 'S')]
        shas = [f'{i+10:040x}' for i in range(len(late))]
        # The landing candidate touches stage3/, so the classifier queues it big.
        git = (w.bin / 'git').read_text().replace('*merge-base*', f"*diff*{shas[0]}*) echo stage3/x.go ;;\n*merge-base*")
        w.script(w.bin / 'git', git.split('\n', 1)[1])
        w.put('tips', ''.join(f'{sha}\trefs/heads/{b}\n' for b, sha in w.tips + [(b, s) for (b, _), s in zip(late, shas)]))
        w.wait(lambda: w.read('output').count('queued ') == 3)
        w.wait(lambda: len(w.read('starts').splitlines()) == 3)
        time.sleep(.2)
        starts = w.read('starts').splitlines()
        self.assertEqual(len(starts), 3, starts)
        self.assertTrue(starts[2].endswith(' other'), starts)
        # Other's slot frees first: the third tip takes it, and the reserved tip still waits for Server.
        w.put('mode', 'pass')
        w.wait(lambda: any(x.startswith('cloud/land-area-next1 ') for x in w.read('starts').splitlines()))
        reserved = [x for x in w.read('starts').splitlines() if x.startswith('cloud/land-area-next1 ')][0]
        self.assertTrue(reserved.endswith(' B server'), reserved)
        w.wait(lambda: any(x.startswith('codex/third ') and x.endswith(' other') for x in w.read('starts').splitlines()))

    def test_canary_takes_a_big_slot_when_no_small_one_is_free(self):
        w = Watcher(0)
        self.addCleanup(w.close)
        (w.state / 'slots').write_text('busy S\nhome B\n')
        # A gate already on busy's only small slot, from before this watcher started.
        holder = subprocess.Popen(['sleep', '30'])
        self.addCleanup(holder.kill)
        (w.state / 'running' ).mkdir(exist_ok=True)
        (w.state / 'running' / str(holder.pid)).write_text('codex/old ' + 'c' * 40 + ' S busy S x ' + str(w.state / 'logs/old.log') + '\n')
        w.wait(lambda: 'canary/main' in w.read('starts'))
        self.assertTrue(w.read('starts').splitlines()[0].endswith(' B home'), w.read('starts'))

    def test_a_long_waiting_small_tip_takes_an_idle_area_slot(self):
        # The clock reads 1000: a small tip queued at 800 has waited 200 s, one queued at 950 only 50 s.
        w = self.reservation('busy S\nhome B\n', [('codex/small-old', 'S')])
        holder = subprocess.Popen(['sleep', '30'])
        self.addCleanup(holder.kill)
        (w.state / 'running' / str(holder.pid)).write_text('codex/old ' + 'c' * 40 + ' S busy S x ' + str(w.state / 'logs/old.log') + '\n')
        (w.state / 'queue').write_text('S 800 codex/small-old %s\n' % w.tips[0][1])
        w.wait(lambda: any(x.startswith('codex/small-old ') for x in w.read('starts').splitlines()))
        line = [x for x in w.read('starts').splitlines() if x.startswith('codex/small-old ')][0]
        self.assertTrue(line.endswith(' B home'), line)

    def test_a_small_tip_leaves_the_area_slot_while_young_or_a_big_tip_waits(self):
        for queue, why in [('S 950 codex/young {sha}\n', 'young'),
                           ('S 800 codex/young {sha}\nB 990 codex/big {big}\n', 'big queued')]:
            with self.subTest(why=why):
                w = self.reservation('busy S\nbusy S\n', [('codex/young', 'S'), ('codex/big', 'B')])
                (w.state / 'slots').write_text('busy S\nhome B\n')
                holder = subprocess.Popen(['sleep', '30'])
                self.addCleanup(holder.kill)
                (w.state / 'running' / str(holder.pid)).write_text('codex/old ' + 'c' * 40 + ' S busy S x ' + str(w.state / 'logs/old.log') + '\n')
                (w.state / 'queue').write_text(queue.format(sha=w.tips[0][1], big=w.tips[1][1]))
                time.sleep(.3)
                self.assertFalse(any(x.startswith('codex/young ') for x in w.read('starts').splitlines()), w.read('starts'))

    def test_the_deploy_canary_may_use_a_draining_box(self):
        # The only free box drains for a queued reservation: the canary still runs there, then the reserved tip.
        w = Watcher(0)
        self.addCleanup(w.close)
        (w.state / 'slots').write_text('server B cloud/land-area-next*\nserver S\n')
        w.put('tips', '')
        (w.state / 'queue').write_text('B 900 cloud/land-area-next-auto-1 %s\n' % ('d' * 40))
        (w.state / 'seen').write_text('cloud/land-area-next-auto-1 %s\n' % ('d' * 40))
        w.wait(lambda: 'canary/main' in w.read('starts'))
        self.assertTrue(w.read('starts').splitlines()[0].endswith(' S server'), w.read('starts'))
        w.put('initial', 'pass')
        w.wait(lambda: 'cloud/land-area-next-auto-1' in w.read('starts'))

    def test_a_lock_left_by_a_killed_watcher_is_cleared_at_start(self):
        w = Watcher(0, staleLock=True)
        self.addCleanup(w.close)
        w.wait(lambda: 'canary/main' in w.read('starts'))
        self.assertIn('cleared a slot-table lock', w.read('output'))

    def test_a_dead_holder_s_lock_is_taken_over_and_a_live_one_is_not(self):
        dead = subprocess.Popen(['true'])
        dead.wait()
        w = Watcher(0, staleLock=str(dead.pid))
        self.addCleanup(w.close)
        w.wait(lambda: 'canary/main' in w.read('starts'))
        self.assertIn('took over a slot-table lock whose holder %d' % dead.pid, w.read('output'))
        live = subprocess.Popen(['sleep', '30'])
        self.addCleanup(live.kill)
        w2 = Watcher(0, staleLock=str(live.pid))
        self.addCleanup(w2.close)
        time.sleep(.4)
        self.assertNotIn('canary/main', w2.read('starts'))

    def test_a_stall_alarms_once(self):
        w = self.reservation('busy S\n', [('codex/waiting', 'S')])
        holder = subprocess.Popen(['sleep', '30'])
        self.addCleanup(holder.kill)
        (w.state / 'running' / str(holder.pid)).write_text('codex/old ' + 'c' * 40 + ' S busy S x ' + str(w.state / 'logs/old.log') + '\n')
        w.wait(lambda: 'done canary:' in w.read('output'))
        messages = w.read('messages')
        w.put('clock', '1950')
        w.wait(lambda: 'stall alarm sent' in w.read('output'))
        time.sleep(.2)
        new = w.read('messages')[len(messages):]
        self.assertEqual(new.count('fast gate stall'), 2, new)  # developer tools and integration, once
        self.assertIn('nothing dispatched for', new)
        self.assertIn('queue head: codex/waiting', new)

    def test_a_box_held_whole_is_not_a_free_slot_to_the_stall_alarm(self):
        # Server's small slot is free in the table, but a reserved gate holds the box: fifteen minutes, not five.
        w = self.reservation('server B cloud/land-area-next*\nserver S\n', [('codex/waiting', 'S')])
        holder = subprocess.Popen(['sleep', '30'])
        self.addCleanup(holder.kill)
        (w.state / 'running' / str(holder.pid)).write_text('cloud/land-area-next-1 ' + 'c' * 40 + ' B server B x ' + str(w.state / 'logs/old.log') + '\n')
        (w.state / 'reserved-running' / str(holder.pid)).write_text('server\n')
        w.wait(lambda: 'done canary:' in w.read('output') or 'gating canary' not in w.read('output'))
        w.put('clock', '1400')
        time.sleep(.3)
        self.assertNotIn('stall alarm sent', w.read('output'))
        w.put('clock', '2000')
        w.wait(lambda: 'stall alarm sent' in w.read('output'))
        self.assertIn('and 0 slots free', w.read('messages'))

    def test_a_complete_run_s_first_failure_publishes_once_and_a_worker_s_does_not(self):
        w = Watcher(0)
        self.addCleanup(w.close)
        w.put('initial', 'pass')
        w.wait(lambda: 'done canary:' in w.read('output'))
        holders = []
        for branch, sha in [('cloud/land-x', 'e' * 40), ('codex/worker', 'f' * 40)]:
            holder = subprocess.Popen(['sleep', '30'])
            self.addCleanup(holder.kill)
            holders.append(holder)
            (w.state / 'logs' / (sha[:12] + '.log')).write_text('fast gate: %s against main x\nFIRST FAILURE (tests, at 12.5 s):\npkg TestA\n    a_test.go:3: wrong\n' % sha)
            (w.state / 'running' / str(holder.pid)).write_text('%s %s B box0 B x %s\n' % (branch, sha, w.state / 'logs' / (sha[:12] + '.log')))
        w.wait(lambda: 'early red cloud/land-x' in w.read('output'))
        time.sleep(.3)
        self.assertEqual(w.read('output').count('early red'), 1)

    def test_a_running_gate_on_the_skip_list_is_stopped_once(self):
        w = Watcher(0)
        self.addCleanup(w.close)
        w.put('initial', 'pass')
        w.wait(lambda: 'done canary:' in w.read('output'))
        holder = subprocess.Popen(['sleep', '30'])
        self.addCleanup(holder.kill)
        sha = '9' * 40
        (w.state / 'running' / str(holder.pid)).write_text('cloud/land-old %s B box0 B x %s\n' % (sha, w.state / 'logs/old.log'))
        (w.state / 'skip').write_text('cloud/land-old %s\n' % sha)
        w.wait(lambda: 'stopped cloud/land-old %s: on the skip list' % sha in w.read('output'))
        time.sleep(.3)
        self.assertEqual(w.read('output').count('on the skip list'), 1)
        self.assertEqual(w.read('stops').split()[0], 'box0')
        # Its end is no void: nothing counted toward a storm, nothing requeued.
        (w.state / 'logs/old.log').write_text('void: %s stopped before a verdict: on the skip list\n' % sha)
        holder.kill()
        holder.wait()  # a zombie still answers kill -0
        w.wait(lambda: 'before a verdict, as asked: not a void' in w.read('output'))
        self.assertFalse((w.state / 'void-window').exists() and (w.state / 'void-window').read_text().strip())
        self.assertNotIn(sha, (w.state / 'queue').read_text())

    def test_a_landing_borrows_a_small_slot_ahead_of_small_worker_tips(self):
        # No tips on origin, so the queue holds exactly these two; one small slot and no area slot.
        w = Watcher(0)
        self.addCleanup(w.close)
        w.put('initial', 'pass')
        w.wait(lambda: 'done canary:' in w.read('output'))
        w.put('mode', 'hold')
        small, landing = '1' * 40, '2' * 40
        (w.state / 'seen').write_text('cloud/land-x %s\ncodex/small %s\n' % (landing, small))
        w.put('tips', '%s\trefs/heads/cloud/land-x\n%s\trefs/heads/codex/small\n' % (landing, small))
        (w.state / 'slots').write_text('box S\n')
        # The small tip is older; without the landing's exemption it would go first (rank 4 against 11).
        (w.state / 'queue').write_text('S 800 codex/small %s\nB 900 cloud/land-x %s\n' % (small, landing))
        w.wait(lambda: len(w.read('starts').splitlines()) >= 2)
        self.assertTrue(w.read('starts').splitlines()[1].startswith('cloud/land-x %s S box' % landing), w.read('starts'))

    def test_the_queue_runs_in_roadmap_step_order(self):
        # Two steps declare branches (a first, b second); the oldest tip serves neither and goes last.
        w = Watcher(0)
        self.addCleanup(w.close)
        w.put('ready', ' 9  #aaaaaaa  Running  first\n 3  #bbbbbbb  Running  second\n')
        w.put('task-aaaaaaa', 'Branches: codex/step-a*\n')
        w.put('task-bbbbbbb', 'Branches: codex/step-b*\n')
        w.put('initial', 'pass')
        w.wait(lambda: 'done canary:' in w.read('output'))
        w.wait(lambda: (w.state / 'step-globs').exists() and 'codex/step-b*' in (w.state / 'step-globs').read_text())
        w.put('mode', 'pass')
        (w.state / 'slots').write_text('box S\n')
        (w.state / 'queue').write_text('S 700 codex/other %s\nS 800 codex/step-b-x %s\nS 900 codex/step-a-x %s\n' % ('1' * 40, '2' * 40, '3' * 40))
        (w.state / 'seen').write_text('codex/other %s\ncodex/step-a-x %s\ncodex/step-b-x %s\n' % ('1' * 40, '3' * 40, '2' * 40))
        w.put('tips', '%s\trefs/heads/codex/other\n%s\trefs/heads/codex/step-a-x\n%s\trefs/heads/codex/step-b-x\n' % ('1' * 40, '3' * 40, '2' * 40))
        w.wait(lambda: len([x for x in w.read('starts').splitlines() if x.startswith('codex/')]) == 3)
        self.assertEqual([x.split()[0] for x in w.read('starts').splitlines() if x.startswith('codex/')],
                         ['codex/step-a-x', 'codex/step-b-x', 'codex/other'])

    def test_a_realistic_queue_fills_and_refills_every_slot_in_seconds(self):
        # @system_adamic, Oct 8: a watcher change is tested against a realistic queue before it deploys.
        # 150 tips, a run of discards, sixteen slots on four boxes, and every gate finishing at once: each
        # fill and each reap-and-refill must take seconds, not the minutes a per-pick ranking took.
        slots = ''.join(f'box{b} B\nbox{b} S\nbox{b} S\nbox{b} S\n' for b in range(4))
        w = Watcher(150, mode='hold', slots=slots)
        self.addCleanup(w.close)
        discards = ''.join('S 800 codex/stale%d %040x\n' % (i, i + 1) for i in range(40))
        (w.state / 'queue').write_text(discards + (w.state / 'queue').read_text())
        workers = lambda: [x for x in w.read('starts').splitlines() if x.startswith('codex/test')]
        w.put('initial', 'pass')
        w.wait(lambda: 'done canary:' in w.read('output'), seconds=60)
        started = time.monotonic()
        # Twelve small slots; the area slots stay for big tips until a small one has waited two minutes.
        w.wait(lambda: len(workers()) == 12, seconds=60)
        fill = time.monotonic() - started
        started = time.monotonic()
        w.put('mode', 'pass')
        w.wait(lambda: len(workers()) >= 24, seconds=60)
        refill = time.monotonic() - started
        print('fill %.1f s, reap and refill %.1f s' % (fill, refill))
        self.assertLess(fill, 8, 'twelve slots took %.1f s to fill' % fill)
        self.assertLess(refill, 8, 'twelve finished gates took %.1f s to reap and refill' % refill)

    def test_discards_leave_the_queue_in_one_pass_before_ranking(self):
        w = self.start(0)
        live = ('codex/live', '7' * 40)
        stale = [('codex/old%d' % i, '%040x' % (i + 1)) for i in range(20)]
        (w.state / 'skip').write_text('codex/hand %s\n' % ('8' * 40))
        (w.state / 'skip-globs').write_text('codex/views-* why\n')
        (w.state / 'gated').write_text('9' * 40 + '\n')
        (w.state / 'seen').write_text('codex/live %s\ncodex/hand %s\ncodex/views-a %s\n' % ('7' * 40, '8' * 40, '6' * 40))
        w.put('tips', '%s\trefs/heads/codex/live\n%s\trefs/heads/codex/hand\n%s\trefs/heads/codex/views-a\n' % ('7' * 40, '8' * 40, '6' * 40))
        rows = ['S %d %s %s' % (800 + i, b, sha) for i, (b, sha) in enumerate(stale)]
        rows += ['S 900 codex/hand %s' % ('8' * 40), 'S 901 codex/views-a %s' % ('6' * 40), 'S 902 codex/gone %s' % ('9' * 40), 'S 950 codex/live %s' % ('7' * 40)]
        (w.state / 'queue').write_text('\n'.join(rows) + '\n')
        w.put('mode', 'pass')
        w.put('initial', 'pass')
        w.wait(lambda: 'codex/live ' in w.read('starts'))
        output = w.read('output')
        self.assertEqual(sum('superseded codex/old' in line for line in output.splitlines()), 20)
        self.assertIn('skipped by hand codex/hand', output)
        self.assertIn('skipped by pattern codex/views-* codex/views-a', output)
        self.assertNotIn('codex/gone', w.read('starts'))

    def test_front_outranks_every_step_and_skip_globs_take_a_family_out(self):
        # The steps' globs are known before any slot exists, so the second step's tip would go first without
        # front. (The first step's own tips are reserved and stay ahead of front.)
        w = Watcher(0)
        self.addCleanup(w.close)
        w.put('ready', ' 9  #aaaaaaa  Running  first\n 3  #bbbbbbb  Running  second\n')
        w.put('task-aaaaaaa', 'Branches: codex/step-a*\n')
        w.put('task-bbbbbbb', 'Branches: codex/step-b*\n')
        (w.state / 'slots').write_text('')
        # Ranking reads front once per pass, so it is in place before any pass could rank these tips.
        (w.state / 'front').write_text('# ruled to land first\ncloud/land-gate-speed*\n')
        (w.state / 'skip-globs').write_text('codex/views-* until compiler/area-views-next lands\n')
        w.put('initial', 'pass')
        w.wait(lambda: 'done canary:' in w.read('output'))
        w.wait(lambda: (w.state / 'step-globs').exists() and 'codex/step-b*' in (w.state / 'step-globs').read_text())
        tips = [('codex/step-b-x', '1' * 40), ('cloud/land-gate-speed', '2' * 40), ('codex/views-a', '3' * 40)]
        (w.state / 'seen').write_text(''.join('%s %s\n' % t for t in tips))
        w.put('tips', ''.join('%s\trefs/heads/%s\n' % (sha, b) for b, sha in tips))
        (w.state / 'queue').write_text('S 700 cloud/land-gate-speed %s\nS 900 codex/step-b-x %s\nS 950 codex/views-a %s\n' % ('2' * 40, '1' * 40, '3' * 40))
        w.put('mode', 'pass')
        # One slot, so the order is the queue's: the slot opens last.
        (w.state / 'slots').write_text('box S\n')
        w.wait(lambda: 'skipped by pattern codex/views-* codex/views-a' in w.read('output'))
        w.wait(lambda: len([x for x in w.read('starts').splitlines() if not x.startswith('canary/')]) == 2)
        self.assertEqual([x.split()[0] for x in w.read('starts').splitlines() if not x.startswith('canary/')],
                         ['cloud/land-gate-speed', 'codex/step-b-x'])

    def test_the_star_preempts_side_work_on_server_and_runs_there_whole(self):
        # @system_adamic, Oct 8 21:47Z: the star owns its box. A side gate already on Server is stopped on purpose
        # and queued again (not a void), the star runs big and whole on Server even though it was queued small,
        # and the side gate finds another box.
        w = self.reservation('server B\nserver S\nother S\n', [('codex/side', 'S')], release=False)
        (w.state / 'front').write_text('# the star\ncloud/land-train-*\n')
        w.put('initial', 'pass')
        w.wait(lambda: any(x.startswith('codex/side ') for x in w.read('starts').splitlines()))
        side = w.tips[0][1]
        self.assertTrue([x for x in w.read('starts').splitlines() if x.startswith('codex/side ')][0].endswith(' server'))
        star = 'c' * 40
        w.tips.append(('cloud/land-train-1-views-slice1', star))
        w.put('tips', ''.join(f'{sha}\trefs/heads/{b}\n' for b, sha in w.tips))
        w.wait(lambda: f'preempted codex/side {side} on server for the star' in w.read('output'))
        self.assertIn(f'server bash -s -- {side} ', w.read('stops'))
        w.wait(lambda: f'preempted codex/side {side} stopped for the star, queued again' in w.read('output'))
        w.wait(lambda: any(x.startswith('cloud/land-train-1-views-slice1 ') for x in w.read('starts').splitlines()))
        starStart = [x for x in w.read('starts').splitlines() if x.startswith('cloud/land-train-1-views-slice1 ')][0]
        self.assertTrue(starStart.endswith(' B server'), starStart)
        self.assertIn('cloud/land-train-1-views-slice1', w.read('whole'))
        w.wait(lambda: [x for x in w.read('starts').splitlines() if x.startswith('codex/side ')][-1].endswith(' other'))
        self.assertEqual((w.state / 'star-boxes').read_text(), 'server\n')
        self.assertNotIn('void codex/side', w.read('output'))

    def test_a_box_running_the_star_takes_no_other_gate(self):
        # No Server in the table: the star runs wherever a big slot is free, and that box is then its own.
        w = self.reservation('workshop B\nworkshop S\n',
                             [('cloud/land-train-1-views-slice1', 'B'), ('codex/side', 'S')], release=False)
        (w.state / 'front').write_text('cloud/land-train-*\n')
        w.put('initial', 'pass')
        w.wait(lambda: any(x.startswith('cloud/land-train-1-views-slice1 ') for x in w.read('starts').splitlines()))
        time.sleep(.3)
        self.assertNotIn('codex/side', w.read('starts'))
        self.assertEqual((w.state / 'star-boxes').read_text(), 'workshop\n')
        w.put('mode', 'pass')
        w.wait(lambda: 'codex/side' in w.read('starts'))

    def test_stop_gate_refuses_anything_but_a_whole_sha(self):
        w = self.start(0)
        for sha in ('', 'abc', 'g' * 40, 'a' * 39, 'A' * 40):
            result = subprocess.run(['bash', str(w.repo / 'cloud' / 'stop-gate.sh'), 'workshop', sha, 'by hand'],
                                    env=dict(os.environ, PATH=str(w.bin) + ':' + os.environ['PATH'], TEST_ROOT=str(w.root)),
                                    capture_output=True, text=True)
            self.assertEqual(result.returncode, 2, sha)
        self.assertEqual(w.read('stops'), '')
        result = subprocess.run(['bash', str(w.repo / 'cloud' / 'stop-gate.sh'), 'workshop', 'b' * 40, 'by hand'],
                                env=dict(os.environ, PATH=str(w.bin) + ':' + os.environ['PATH'], TEST_ROOT=str(w.root)),
                                capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn('workshop bash -s -- ' + 'b' * 40, w.read('stops'))
        # The box side refuses too, so a caller that skips the script can't widen the match.
        self.assertIn('is not 40 hex digits', w.read('stop-command'))

    def test_an_ahead_tip_outranks_every_step_without_the_star_s_box(self):
        w = Watcher(0)
        self.addCleanup(w.close)
        w.put('ready', ' 9  #aaaaaaa  Running  first\n 3  #bbbbbbb  Running  second\n')
        w.put('task-aaaaaaa', 'Branches: codex/step-a*\n')
        w.put('task-bbbbbbb', 'Branches: codex/step-b*\n')
        (w.state / 'slots').write_text('')
        (w.state / 'front').write_text('cloud/land-train-*\n')
        (w.state / 'ahead').write_text('# test-only splits\ncloud/land-test-split-*\n')
        w.put('initial', 'pass')
        w.wait(lambda: 'done canary:' in w.read('output'))
        w.wait(lambda: (w.state / 'step-globs').exists() and 'codex/step-b*' in (w.state / 'step-globs').read_text())
        tips = [('codex/step-b-x', '1' * 40), ('cloud/land-test-split-flow', '2' * 40)]
        (w.state / 'seen').write_text(''.join('%s %s\n' % t for t in tips))
        w.put('tips', ''.join('%s\trefs/heads/%s\n' % (sha, b) for b, sha in tips))
        (w.state / 'queue').write_text('S 700 codex/step-b-x %s\nS 900 cloud/land-test-split-flow %s\n' % ('1' * 40, '2' * 40))
        w.put('mode', 'pass')
        (w.state / 'slots').write_text('box S\n')
        w.wait(lambda: len([x for x in w.read('starts').splitlines() if not x.startswith('canary/')]) == 2)
        self.assertEqual([x.split()[0] for x in w.read('starts').splitlines() if not x.startswith('canary/')],
                         ['cloud/land-test-split-flow', 'codex/step-b-x'])
        # Not the star: no whole box, and its box is never a star box.
        self.assertNotIn('cloud/land-test-split-flow', w.read('whole'))
        self.assertEqual((w.state / 'star-boxes').read_text(), '')

    def test_staged_tools_run_on_the_canary_box_until_a_real_green_promotes_them(self):
        # Tips, queue and slots are in place before the watcher starts: with no deploy barrier to hold it,
        # a write after its first poll races the watcher's own rewrite of the queue.
        w = Watcher(2, canaryBox='box1', mode='hold')
        self.addCleanup(w.close)
        # No fleet-wide deploy barrier while staged: both start at once, one per box and version.
        w.wait(lambda: w.read('starts').strip() and w.read('good-starts').strip())
        self.assertNotIn('canary/main', w.read('starts') + w.read('good-starts'))
        self.assertTrue(w.read('starts').strip().endswith(' box1'), w.read('starts'))
        self.assertTrue(w.read('good-starts').strip().endswith(' box0'), w.read('good-starts'))
        w.put('mode', 'pass')
        w.wait(lambda: 'promoted tools tools-one' in w.read('output'))
        self.assertEqual((w.state / 'tools-good').read_text().strip(), 'tools-one')
        # Promoted, every box runs the new tools. The new tip arrives as a pushed branch, which the watcher
        # queues itself, never as a write to the queue it owns.
        w.tips.append(('codex/c', 'c' * 40))
        w.put('tips', ''.join(f'{sha}\trefs/heads/{b}\n' for b, sha in w.tips))
        w.wait(lambda: 'codex/c ' in w.read('starts'))
        self.assertNotIn('codex/c ', w.read('good-starts'))

    def test_a_tools_commit_that_leaves_the_box_side_alone_neither_stages_nor_waits_on_a_canary(self):
        # Staged tools: good is tools-zero, head tools-one, but tools-one changes only the Mac side.
        w = Watcher(2, canaryBox='box1', mode='hold', boxSides={'tools-one': 'tools-zero'})
        self.addCleanup(w.close)
        w.wait(lambda: len(w.read('starts').splitlines()) + len(w.read('good-starts').splitlines()) >= 1)
        w.put('mode', 'pass')
        w.wait(lambda: w.read('output').count('done codex/') == 2)
        self.assertEqual(w.read('good-starts'), '', 'no staging: every box runs the head tools')
        # A Mac-only commit while running: no deploy canary, and the queue keeps flowing.
        w.put('head', 'tools-two')
        w.put('box-tools-two', 'tools-zero')
        w.wait(lambda: 'tools tools-two change only the Mac side' in w.read('output'))
        w.tips.append(('codex/c', 'c' * 40))
        w.put('tips', ''.join(f'{sha}\trefs/heads/{b}\n' for b, sha in w.tips))
        w.wait(lambda: 'done codex/c' in w.read('output'))
        self.assertNotIn('canary/main', w.read('starts'))

    def test_control_without_globs(self):
        w = self.reservation('server B\nserver S\n',
                             [('cloud/land-other', 'B'), ('codex/small', 'S')])
        w.wait(lambda: len(w.read('starts').splitlines()) == 3)
        self.assertEqual([x.split()[2:] for x in w.read('starts').splitlines()[1:]],
                         [['B', 'server'], ['S', 'server']])

    def test_roadmap_refresh_and_file_reservation(self):
        w = self.reservation('server B\nserver S\nother B\n',
                             [('codex/first', 'B'), ('cloud/land-other', 'B')], release=False)
        # Hold the deploy canary while the next five-minute refresh arrives.
        w.put('initial', 'hold')
        w.put('ready', '#aaaaaaa no branches\n#2s8kq6y first\n#bbbbbbb later\n')
        w.put('task-aaaaaaa', 'No branch declaration\n')
        w.put('task-2s8kq6y', 'Branches: codex/first cloud/land-area-next*\n')
        w.put('task-bbbbbbb', 'Branches: codex/later\n')
        w.put('clock', '1300')
        w.wait(lambda: (w.state / 'first-step-globs').read_text() == 'codex/first\ncloud/land-area-next*\n')
        # The same refresh then reads every ready step for the queue's step order.
        w.wait(lambda: (w.state / 'step-globs').exists() and (w.state / 'step-globs').read_text() ==
               '0 codex/first\n0 cloud/land-area-next*\n1 codex/later\n')
        calls = w.read('task-calls')
        time.sleep(.15)
        self.assertEqual(w.read('task-calls'), calls)
        w.put('initial', 'pass')
        w.wait(lambda: len(w.read('starts').splitlines()) == 3)
        self.assertTrue(w.read('starts').splitlines()[1].startswith('codex/first '))
        self.assertTrue(w.read('starts').splitlines()[1].endswith(' server'))

    def test_no_branch_line_clears_roadmap_reservation(self):
        w = self.start(0)
        w.put('ready', '#2s8kq6y first\n')
        w.put('task-2s8kq6y', 'No candidates yet\n')
        (w.state / 'first-step-globs').write_text('cloud/land-old*\n')
        w.put('clock', '1300')
        w.wait(lambda: 'show 2s8kq6y' in w.read('task-calls'))
        w.wait(lambda: (w.state / 'first-step-globs').read_text() == '')

    def test_storm_hold_recovery(self):
        w = self.start()
        # The sixth concurrent void would otherwise exhaust this tip's retries.
        (w.state / 'void-tries').write_text((w.tips[-1][1] + '\n') * 3)
        w.put('initial', 'pass')
        w.wait(lambda: sum(line.startswith('codex/') for line in w.read('starts').splitlines()) == 6)
        w.put('canary', 'hold')
        w.wait(lambda: (w.state / 'storm').exists() and len(w.read('messages').splitlines()) == 2)
        w.wait(lambda: 'held during storm' in w.read('output') and len(w.read('starts').splitlines()) == 8)
        starts = w.read('starts')
        time.sleep(.15)
        self.assertEqual(w.read('starts'), starts)
        self.assertNotIn('three tries, given up', w.read('output'))
        self.assertEqual((w.state / 'void-tries').read_text().count(w.tips[-1][1]), 3)
        storm = (w.state / 'storm').read_text()
        first_log = (w.state / 'void-window').read_text().splitlines()[0].split(' ', 1)[1]
        self.assertIn(first_log, storm)
        for recipient in ('system_adamic_developer_tools', 'system_adamic_integration'):
            self.assertEqual(w.read('messages').count(recipient + '|'), 1)
        self.assertIn('void failure', w.read('messages'))
        self.assertNotRegex(w.read('messages'), r'[A-Z]{3,}')
        self.assertEqual({x.split()[-1] for x in w.read('starts').splitlines() if x.startswith('codex/')}, {'box0', 'box1'})
        self.assertIn('--from|system_adamic_developer_tools', w.read('messages'))
        w.put('mode', 'pass')
        w.put('canary', 'pass')
        w.wait(lambda: not (w.state / 'storm').exists() and len(w.read('messages').splitlines()) == 4)
        w.wait(lambda: len((w.state / 'gated').read_text().splitlines()) == 6)
        self.assertNotIn(MAIN, (w.state / 'gated').read_text())
        self.assertNotIn('done canary/main', w.read('output'))
        self.assertIn('done canary:', w.read('output'))
        self.assertNotIn('canary/main', w.read('merges'))
        self.assertNotIn(MAIN, w.read('merges'))

    def test_control_spread_voids_and_normal_retry_limit(self):
        w = self.start(1)
        # Voids more than ten minutes apart cannot accumulate into a storm.
        w.put('step', '601')
        (w.state / 'void-window').write_text('1 old-log\n' * 5)
        w.put('initial', 'pass')
        w.wait(lambda: 'three tries, given up' in w.read('output'))
        self.assertFalse((w.state / 'storm').exists())
        self.assertEqual(len((w.state / 'void-window').read_text().splitlines()), 1)
        self.assertEqual(len((w.state / 'void-tries').read_text().splitlines()), 4)
        self.assertEqual(sum(x.startswith('codex/') for x in w.read('starts').splitlines()), 4)

    def test_void_deploy_canary_and_ten_minute_probes(self):
        w = self.start(1)
        w.put('initial', 'void')
        w.put('canary', 'void')
        w.wait(lambda: (w.state / 'storm').exists() and len(w.read('starts').splitlines()) == 2)
        time.sleep(.12)
        self.assertEqual(len(w.read('starts').splitlines()), 2)
        w.put('clock', '1599')
        time.sleep(.12)
        self.assertEqual(len(w.read('starts').splitlines()), 2)
        w.put('canary', 'red')
        w.put('mode', 'pass')
        w.put('clock', '1600')
        w.wait(lambda: not (w.state / 'storm').exists())
        w.wait(lambda: 'done codex/' in w.read('output'))
        self.assertEqual(len(w.read('messages').splitlines()), 4)

    def test_tools_head_change_blocks_queue(self):
        w = self.start(0)
        w.put('initial', 'pass')
        w.wait(lambda: 'done canary:' in w.read('output'))
        w.put('canary', 'hold')
        w.put('head', 'tools-two')
        w.wait(lambda: len(w.read('starts').splitlines()) == 2)
        sha = 'b' * 40
        w.put('tips', f'{sha}\trefs/heads/codex/new\n')
        w.wait(lambda: 'queued codex/new' in w.read('output'))
        time.sleep(.12)
        self.assertEqual(len(w.read('starts').splitlines()), 2)
        w.put('mode', 'pass')
        w.put('canary', 'pass')
        w.wait(lambda: 'done codex/new' in w.read('output'))


if __name__ == '__main__':
    unittest.main()
