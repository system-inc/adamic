#!/usr/bin/env python3
"""The star-idle alarm against today's two cases, replayed from stand-in watcher files and a stand-in ahra."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import time
import unittest

script = Path(__file__).with_name('star-idle-alarm.py')
star, other = 'c' * 40, 'd' * 40


class StarIdleAlarmTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        (self.root / 'running').mkdir()
        (self.root / 'bin').mkdir()
        waterfall = {'nodes': [{'id': 'a03mesg', 'wave': 0}, {'id': 'later', 'wave': 1}], 'criticalPath': ['a03mesg', 'later']}
        (self.root / 'waterfall.json').write_text(json.dumps(waterfall))
        (self.root / 'bin/ahra').write_text('''#!/bin/bash
case "$1 $2" in
  "tasks waterfall") cat "$ROOT/waterfall.json" ;;
  "tasks show") printf 'owner     @system_adamic_compiler (direct)\\n    Branches: cloud/land-*views-slice* compiler/area-stack*\\n' ;;
  "os send") printf '%s|%s\\n' "$3" "$4" >> "$ROOT/sends" ;;
esac
''')
        (self.root / 'bin/ahra').chmod(0o755)
        self.env = dict(os.environ, PATH=str(self.root / 'bin') + ':' + os.environ['PATH'], ROOT=str(self.root),
                        ADAMIC_FAST_GATE_WATCH_STATE=str(self.root), ADAMIC_FAST_GATE_WATCH_LOG=str(self.root / 'watch.log'),
                        ADAMIC_FAST_GATE_AHRA_DIR=str(self.root), ADAMIC_FULL_GATE_LOG=str(self.root / 'full.log'),
                        ADAMIC_MAIN_REDS=str(self.root / 'main-reds.tsv'))
        self.now = int(time.time())
        for name in ('queue', 'queue.ranked', 'seen', 'watch.log', 'full.log', 'main-reds.tsv'):
            (self.root / name).write_text('')

    def check(self):
        subprocess.run(['python3', str(script), '--once'], env=self.env, check=True, capture_output=True)
        sends = self.root / 'sends'
        return sends.read_text().splitlines() if sends.exists() else []

    def queueStar(self, waited, ahead=7):
        # Case one (Oct 8, about 20:2xZ): the star queued eighth, behind scouts, every slot it could take busy.
        lines = ['B %d codex/scout-%d %040x' % (self.now - 9000, i, i + 1) for i in range(ahead)]
        lines.append('B %d cloud/land-stack-x1-views-slice1 %s' % (self.now - waited, star))
        (self.root / 'queue').write_text('\n'.join(lines) + '\n')
        ranked = ['%d %d B %s %s no' % (400 + i, self.now - 9000, line.split()[2], line.split()[3]) for i, line in enumerate(lines)]
        (self.root / 'queue.ranked').write_text('\n'.join(ranked) + '\n')

    def test_a_star_queued_behind_lower_ranked_tips_pages_its_owner_and_the_parent_once(self):
        self.queueStar(waited=30)
        self.assertEqual(self.check(), [], 'thirty seconds queued is not yet idle')
        self.queueStar(waited=400)
        sends = self.check()
        self.assertEqual([line.split('|')[0] for line in sends], ['system_adamic_compiler', 'system_adamic'])
        self.assertIn('★a03mesg has no live turn', sends[0])
        self.assertIn('8 of 8 in the ranked queue', sends[0])
        self.assertEqual(len(self.check()), 2, 'the same idle state pages once')
        # A gate starting on the star's branch is a live turn and re-arms the alarm.
        (self.root / 'queue').write_text('')
        (self.root / 'running/123').write_text('cloud/land-stack-x1-views-slice1 %s S server B token log\n' % star)
        self.assertEqual(len(self.check()), 2)
        (self.root / 'running/123').unlink()

    def test_a_red_with_no_newer_push_pages_and_a_push_rearms(self):
        # Case two (Oct 8, 20:29Z): the star's first red sat until a tick read the log.
        (self.root / 'seen').write_text('cloud/land-views-slice1 %s\ncodex/other %s\n' % (star, other))
        (self.root / 'watch.log').write_text(
            '20:28:00 done codex/other: red: %s fast gate, first failure at tests\n'
            '20:29:21 done cloud/land-views-slice1: red: %s fast gate, first failure at a-check after 40 s\n' % (other, star))
        sends = self.check()
        self.assertEqual(len(sends), 2)
        self.assertIn('first failure at a-check', sends[0])
        # A newer push on the branch is the owner's turn taken: no page while it waits its first minute.
        (self.root / 'seen').write_text('cloud/land-views-slice1 %s\n' % other)
        (self.root / 'queue').write_text('B %d cloud/land-views-slice1 %s\n' % (self.now - 5, other))
        self.assertEqual(len(self.check()), 2)

    def test_a_green_or_a_running_star_is_quiet(self):
        (self.root / 'seen').write_text('cloud/land-views-slice1 %s\n' % star)
        (self.root / 'watch.log').write_text('20:29:21 done cloud/land-views-slice1: green: %s fast gate in 900 s\n' % star)
        self.assertEqual(self.check(), [])
        (self.root / 'running/9').write_text('compiler/area-stack %s B threadripper B token log\n' % other)
        self.queueStar(waited=4000)
        self.assertEqual(self.check(), [], 'a gate running on one of its branches is a live turn')


    def mainIsRed(self):
        main = 'e' * 40
        (self.root / 'full.log').write_text('published gate-logs/eeeeeeeeeeee/20261008T202422Z/full-main (abc)\n'
                                            '20:35:28 red: %s first failure at tests after 597.9 s (still running for triage)\n' % main)
        return main

    def test_a_red_main_with_no_fix_forward_named_pages_integration_and_the_parent(self):
        # Oct 8: d72728e5 red at TestSplitTSGoAgrees, before main-reds.tsv named its fix-forward.
        self.mainIsRed()
        sends = self.check()
        self.assertEqual([line.split('|')[0] for line in sends], ['system_adamic_integration', 'system_adamic'])
        self.assertIn('names no fix-forward', sends[0])
        self.assertEqual(len(self.check()), 2)

    def test_a_named_fix_forward_must_be_gating(self):
        self.mainIsRed()
        (self.root / 'main-reds.tsv').write_text('gate-logs/eeeeeeeeeeee/20261008T202422Z/full-main\tsplit build: owner @system_adamic_compiler, fix-forward cloud/land-split-fix*\n')
        sends = self.check()
        self.assertEqual([line.split('|')[0] for line in sends], ['system_adamic_compiler', 'system_adamic'])
        self.assertIn('cloud/land-split-fix* has no live turn', sends[0])
        # Gating is a live turn: quiet, and re-armed.
        (self.root / 'running/77').write_text('cloud/land-split-fix %s S server S token log\n' % other)
        self.assertEqual(len(self.check()), 2)
        self.assertFalse((self.root / 'main-red-alarmed').exists())
        # An explained red (LIFTED, FIXED FORWARD, CLOSED) is not paged.
        (self.root / 'running/77').unlink()
        (self.root / 'main-reds.tsv').write_text('gate-logs/eeeeeeeeeeee/20261008T202422Z/full-main\tFIXED FORWARD by compiler\n')
        self.assertEqual(len(self.check()), 2)

    def test_a_green_main_is_quiet(self):
        (self.root / 'full.log').write_text('20:35:28 red: %s first failure at tests\n21:30:00 green: %s full gate in 3000 s\n' % ('e' * 40, 'f' * 40))
        self.assertEqual(self.check(), [])


if __name__ == '__main__':
    unittest.main()
