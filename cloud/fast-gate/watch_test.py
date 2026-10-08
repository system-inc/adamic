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
    def __init__(self, count=6):
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        self.repo = self.root / 'repo'
        cloud = self.repo / 'cloud'
        cloud.mkdir(parents=True)
        for name in ('fast-gate-watch.sh', 'fast-gate-classify.sh', 'first-step-branches.sh'):
            shutil.copy(ROOT / 'cloud' / name, cloud / name)
        self.state = self.root / 'state'
        self.state.mkdir()
        self.bin = self.root / 'bin'
        self.bin.mkdir()
        self.put('clock', '1000')
        self.put('head', 'tools-one')
        self.put('mode', 'void')
        self.put('initial', 'hold')
        self.put('canary', 'hold')
        self.tips = [(f'codex/test{i}', f'{i+1:012x}' + '0' * 28) for i in range(count)]
        self.put('tips', ''.join(f'{sha}\trefs/heads/{b}\n' for b, sha in self.tips))
        (self.state / 'seen').write_text(''.join(f'{b} {sha}\n' for b, sha in self.tips))
        (self.state / 'queue').write_text(''.join(f'S 900 {b} {sha}\n' for b, sha in self.tips))
        (self.state / 'slots').write_text(''.join(f'box{i % 2} S\n' for i in range(max(count, 1))))
        self.script(self.bin / 'git', '''case "$*" in
*rev-parse*) cat "$TEST_ROOT/head" ;;
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
        env = dict(os.environ, PATH=str(self.bin) + ':' + os.environ['PATH'],
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

    def wait(self, predicate):
        deadline = time.monotonic() + 8
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

    def test_reserved_gate_without_failure_keeps_running(self):
        self.stale_red(red=False)

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
        self.assertTrue(any(x.startswith('codex/small ') and x.endswith(' other') for x in starts))
        time.sleep(.15)
        self.assertEqual(len(w.read('starts').splitlines()), 3)
        w.put('mode', 'pass')
        w.wait(lambda: 'done cloud/land-other' in w.read('output'))

    def test_reserved_slot_rejects_other_tip_but_box_serves_small(self):
        w = self.reservation('server B cloud/land-area-next*\nserver S\n',
                             [('cloud/land-other', 'B'), ('codex/small', 'S')])
        w.wait(lambda: len(w.read('starts').splitlines()) == 2)
        self.assertTrue(w.read('starts').splitlines()[1].startswith('codex/small '))
        self.assertTrue(w.read('starts').splitlines()[1].endswith(' server'))
        time.sleep(.15)
        self.assertEqual(len(w.read('starts').splitlines()), 2)
        w.put('mode', 'pass')
        w.wait(lambda: 'done cloud/land-other' in w.read('output'))
        self.assertIn('cloud/land-other ' + w.tips[0][1] + ' S server', w.read('starts'))

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
        self.assertNotIn('show bbbbbbb', w.read('task-calls'))
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
