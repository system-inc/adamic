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
        for name in ('fast-gate-watch.sh', 'fast-gate-classify.sh'):
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
        self.script(self.bin / 'ahra', '''text=$(printf '%s' "$4" | tr '\\n' ' ')
printf '%s|%s|%s|%s\\n' "$3" "$text" "$5" "$6" >> "$TEST_ROOT/messages"
''')
        self.script(self.bin / 'date', '''case "$*" in
*+%s*) cat "$TEST_ROOT/clock" ;;
*) /bin/date -u +%H:%M:%S ;;
esac
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
        self.proc = subprocess.Popen(['bash', str(cloud / 'fast-gate-watch.sh')], env=env,
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
