#!/usr/bin/env python3
"""The standalone stall alarm against a stand-in queue, slot-wait record and ahra."""
import os
from pathlib import Path
import subprocess
import tempfile
import time
import unittest

script = Path(__file__).with_name('gate-stall-alarm.sh')


def iso(seconds):
    return time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime(seconds))


class StallAlarmTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        (self.root / 'bin').mkdir()
        (self.root / 'bin/ahra').write_text('#!/bin/bash\nprintf "%s|%s\\n" "$3" "$4" >> "$SENDS"\n')
        (self.root / 'bin/ahra').chmod(0o755)
        self.env = dict(os.environ, PATH=str(self.root / 'bin') + ':' + os.environ['PATH'], SENDS=str(self.root / 'sends'),
                        ADAMIC_FAST_GATE_WATCH_STATE=str(self.root), ADAMIC_FAST_GATE_WAITS=str(self.root / 'waits.csv'),
                        ADAMIC_FAST_GATE_AHRA_DIR=str(self.root))
        self.now = int(time.time())
        (self.root / 'slots').write_text('box0 B\nbox0 S\nbox1 S\n')
        (self.root / 'running').mkdir()
        (self.root / 'reserved-running').mkdir()
        self.sleeper = subprocess.Popen(['sleep', '60'])
        self.addCleanup(self.sleeper.wait)
        self.addCleanup(self.sleeper.kill)

    def running(self, box, pid=None, reserved=False):
        pid = pid or self.sleeper.pid
        (self.root / 'running' / str(pid)).write_text('codex/r %s S %s S token log\n' % ('e' * 40, box))
        if reserved:
            (self.root / 'reserved-running' / str(pid)).write_text(box + '\n')

    def state(self, queuedAgo, startedAgo):
        (self.root / 'queue').write_text('' if queuedAgo is None else 'S %d codex/old %s\nB %d codex/new %s\n' % (self.now - queuedAgo, 'a' * 40, self.now - 10, 'b' * 40))
        rows = 'sha,branch,class,queued_utc,started_utc,waited_seconds,outcome,box\n'
        rows += '%s,codex/x,S,%s,%s,5,started,threadripper\n' % ('c' * 40, iso(self.now - startedAgo - 5), iso(self.now - startedAgo))
        rows += '%s,codex/y,S,,%s,,void:no_verdict,threadripper\n' % ('d' * 40, iso(self.now - 1))
        (self.root / 'waits.csv').write_text(rows)

    def check(self):
        subprocess.run(['/bin/bash', str(script), '--once'], env=self.env, check=True, capture_output=True)
        sends = self.root / 'sends'
        return sends.read_text().splitlines() if sends.exists() else []

    def test_a_stall_alarms_both_once_and_a_start_rearms(self):
        self.state(queuedAgo=400, startedAgo=360)
        sends = self.check()
        self.assertEqual([line.split('|')[0] for line in sends], ['system_adamic_developer_tools', 'system_adamic_integration'])
        self.assertIn('no gate has started anywhere for 36', sends[0])
        self.assertIn('codex/old', sends[0])
        self.assertEqual(len(self.check()), 2)
        self.state(queuedAgo=400, startedAgo=20)
        self.assertEqual(len(self.check()), 2)
        self.state(queuedAgo=400, startedAgo=360)
        self.assertEqual(len(self.check()), 4)

    def test_no_alarm_while_young_or_starting_or_empty(self):
        for queued, started in [(100, 360), (400, 100), (None, 900)]:
            with self.subTest(queued=queued, started=started):
                self.state(queued, started)
                self.assertEqual(self.check(), [])

    def test_a_void_row_is_not_a_start(self):
        # The record's newest row is a void one second ago: still a stall.
        self.state(queuedAgo=400, startedAgo=360)
        self.assertEqual(len(self.check()), 2)


    def test_a_full_fleet_is_not_a_stall(self):
        # box0 is drained for a reserved run (its slots aren't usable) and box1's one slot is busy.
        self.state(queuedAgo=400, startedAgo=360)
        self.running('box0', reserved=True)
        other = subprocess.Popen(['sleep', '60'])
        self.addCleanup(other.wait)
        self.addCleanup(other.kill)
        self.running('box1', pid=other.pid)
        self.assertEqual(self.check(), [])
        # A dead watcher's leftover doesn't fill a slot: box1's gate is gone, so its slot is free.
        other.kill()
        other.wait()
        self.assertEqual(len(self.check()), 2)


    def test_a_slot_held_for_a_family_with_none_waiting_is_not_free(self):
        # Oct 8 20:38Z: Server's area slot held for runtime's step, no runtime tip queued, every other slot busy.
        (self.root / 'slots').write_text('box1 S\nserver B\n')
        (self.root / 'first-step-globs.poll').write_text('area/runtime,\ncloud/land-runtime-*\n')
        self.state(queuedAgo=400, startedAgo=360)
        other = subprocess.Popen(['sleep', '60'])
        self.addCleanup(other.wait)
        self.addCleanup(other.kill)
        self.running('box1', pid=other.pid)
        self.assertEqual(self.check(), [], 'the held slot is not free while its family is absent')
        # A tip of its family waiting makes it free, and the idle slot is a stall again.
        with open(self.root / 'queue', 'a') as handle:
            handle.write('B %d cloud/land-runtime-slice3 %s\n' % (self.now - 400, 'f' * 40))
        self.assertEqual(len(self.check()), 2)


if __name__ == '__main__':
    unittest.main()
