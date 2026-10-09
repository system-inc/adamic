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
  "tasks show") cat "$ROOT/show-$3" 2>/dev/null || printf 'owner     @system_adamic_compiler (direct)\\n    Branches: cloud/land-*views-slice* compiler/area-stack*\\n' ;;
  "os send") printf '%s|%s\\n' "$3" "$4" >> "$ROOT/sends" ;;
esac
''')
        (self.root / 'bin/ahra').chmod(0o755)
        self.env = dict(os.environ, PATH=str(self.root / 'bin') + ':' + os.environ['PATH'], ROOT=str(self.root),
                        ADAMIC_FAST_GATE_WATCH_STATE=str(self.root), ADAMIC_FAST_GATE_WATCH_LOG=str(self.root / 'watch.log'),
                        ADAMIC_FAST_GATE_AHRA_DIR=str(self.root), ADAMIC_FULL_GATE_LOG=str(self.root / 'full.log'),
                        ADAMIC_MAIN_REDS=str(self.root / 'main-reds.tsv'), ADAMIC_LANDED_SHAS=str(self.root / 'landed'),
                        ADAMIC_MAIN_HEAD=str(self.root / 'main-head'), ADAMIC_FULL_GATE_RUNNING=str(self.root / 'full-running'),
                        ADAMIC_BRANCH_COMMITS=str(self.root / 'branch-commits'),
                        ADAMIC_FULL_GATE_REQUESTS=str(self.root / 'requests'),
                        ADAMIC_WHOLE_GATES=str(self.root / 'whole-gates'), ADAMIC_FULL_GATE_PREGATES=str(self.root / 'pregate'))
        self.now = int(time.time())
        for name in ('queue', 'queue.ranked', 'seen', 'watch.log', 'full.log', 'main-reds.tsv', 'landed', 'main-head', 'full-running', 'branch-commits', 'requests', 'whole-gates'):
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

    def test_a_whole_gate_running_on_the_star_is_a_live_turn(self):
        # Oct 8 22:21Z: the star's fast gate went red, and its one-off full-main run on Home was the live turn.
        (self.root / 'seen').write_text('cloud/land-train-1-views-slice1 %s\n' % star)
        (self.root / 'watch.log').write_text('22:21:39 done cloud/land-train-1-views-slice1: red: %s fast gate, first failure at deferred\n' % star)
        (self.root / 'full-running').write_text(other + '\n' + star + '\n')
        self.assertEqual(self.check(), [])
        (self.root / 'full-running').write_text(other + '\n')
        self.assertEqual(len(self.check()), 2, 'a whole gate of another commit is not the star\'s turn')

    def test_a_whole_gate_the_loops_run_is_a_live_turn_by_its_record(self):
        (self.root / 'seen').write_text('cloud/land-train-1-views-slice1 %s\n' % star)
        (self.root / 'watch.log').write_text('22:21:39 done cloud/land-train-1-views-slice1: red: %s fast gate, first failure at tests\n' % star)
        (self.root / 'whole-gates').write_text('%s running %d\n' % (star, self.now - 300))
        self.assertEqual(self.check(), [])
        (self.root / 'whole-gates').write_text('')
        self.assertEqual(len(self.check()), 2)

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

    def test_a_red_page_says_under_load_when_its_record_does(self):
        # The witness (Oct 9): the red's own record names load, and the page reads it first.
        label = ', under load (load 178.0 on 64 cores at first failure)'
        (self.root / 'seen').write_text('cloud/land-views-slice1 %s\n' % star)
        (self.root / 'watch.log').write_text('20:29:21 done cloud/land-views-slice1: red: %s fast gate, first failure at tests after 40 s%s (steps), 1 fail, 9 pass\n' % (star, label))
        sends = self.check()
        self.assertIn('ran under load (load 178.0 on 64 cores): read it as load', sends[0])
        # A whole gate's red on a train slice, by its record.
        (self.root / 'seen').write_text('cloud/land-train-2-views-slice1 %s\n' % other)
        (self.root / 'watch.log').write_text('%s done cloud/land-train-2-views-slice1: green: %s fast gate in 600 s\n' % (self.clock(30), other))
        (self.root / 'whole-gates').write_text('%s red %d red: %s full gate, first failure at tests after 1042.6 s%s, cancelled after first failure (steps)\n' % (other, self.now - 60, other, label))
        sends = self.check()
        self.assertIn('whole gate is red, so it can\'t land. Its record says the first failure ran under load (load 178.0 on 64 cores)', sends[-2])
        # Main's red, by its log line; one without the label pages without the note.
        for main, tail, loaded in (('e' * 40, '', False), ('9' * 40, label, True)):
            (self.root / 'full.log').write_text('20:35:28 red: %s full gate, first failure at tests after 597.9 s%s (steps)\n' % (main, tail))
            (self.root / 'main-head').write_text(main + '\n')
            sends = self.check()
            self.assertIn('Main %s is red at tests' % main[:12], sends[-2])
            self.assertEqual('ran under load (load 178.0 on 64 cores)' in sends[-2], loaded)

    def test_a_quiet_worker_on_the_chain_pages_its_owner_once_and_a_push_rearms(self):
        # @system_adamic, Oct 8: V1's worker was quiet for two hours on the critical path and nobody knew.
        waterfall = {'nodes': [{'id': 'a03mesg', 'wave': 0}, {'id': 'v1', 'wave': 1}, {'id': 'v2', 'wave': 2}, {'id': 'v3', 'wave': 3}],
                     'criticalPath': ['a03mesg', 'v1', 'v2', 'v3']}
        (self.root / 'waterfall.json').write_text(json.dumps(waterfall))
        (self.root / 'show-v1').write_text('owner     @system_adamic_typescript (direct)\n    Branches: codex/views-v1-*\n')
        (self.root / 'show-v2').write_text('owner     @system_adamic_typescript (direct)\n    Branches: codex/views-v2-*\n')
        (self.root / 'show-v3').write_text('owner     @system_adamic_runtime (direct)\n    Branches: codex/views-v3-*\n')
        self.assertEqual(self.check(), [], 'a step just seen on the chain is not yet quiet')
        # On the chain for 25 minutes: v1 pushed 21 minutes ago is quiet, v2 pushed 5 minutes ago is not, and the
        # fourth step is off the chain.
        (self.root / 'chain-first-seen.json').write_text(json.dumps({'a03mesg': self.now, 'v1': self.now - 1500, 'v2': self.now - 1500, 'v3': self.now - 1500}))
        (self.root / 'watch.log').write_text('%s queued codex/views-v1-frame %s (B)\n%s queued codex/views-v2-unions %s (S)\n' % (self.clock(1260), star, self.clock(300), other))
        sends = self.check()
        self.assertEqual([line.split('|')[0] for line in sends], ['system_adamic_typescript', 'system_adamic'])
        self.assertIn('#v1 is on the critical path and its worker is quiet', sends[0])
        self.assertIn('codex/views-v1-frame %s, 21 min ago' % star[:12], sends[0])
        self.assertEqual(len(self.check()), 2, 'one quiet spell pages once')
        # In the gate is not quiet, and a newer push re-arms.
        (self.root / 'running/7').write_text('codex/views-v1-frame %s S server S token log\n' % star)
        self.assertEqual(len(self.check()), 2)
        self.assertFalse((self.root / 'chain-quiet-alarmed-v1').exists())
        (self.root / 'running/7').unlink()
        with (self.root / 'watch.log').open('a') as handle:
            handle.write('%s queued codex/views-v1-frame %s (B)\n' % (self.clock(10), other))
        self.assertEqual(len(self.check()), 2)
        # Never pushed and on the chain over 20 minutes is quiet too.
        (self.root / 'chain-first-seen.json').write_text(json.dumps({'a03mesg': self.now, 'v1': self.now, 'v2': self.now - 1300}))
        (self.root / 'watch.log').write_text('')
        sends = self.check()
        self.assertIn('#v2 is on the critical path', sends[-2])
        self.assertIn('Last push: none seen', sends[-2])

    def test_the_chain_reads_running_turns_through_the_star_s_probe_and_shared_branches_by_trailer(self):
        # @system_adamic, Oct 8 22:38Z: two false pages. kvmcfr1's one-off full gate was running the whole time, and
        # V2's work lands on compiler/views-rehearsal, a branch V1 to V6 all name, one slice per Train-slice trailer.
        waterfall = {'nodes': [{'id': 'kvmcfr1', 'wave': 0}, {'id': 'v1', 'wave': 1}, {'id': 'v2', 'wave': 2}],
                     'criticalPath': ['kvmcfr1', 'v1', 'v2']}
        (self.root / 'waterfall.json').write_text(json.dumps(waterfall))
        (self.root / 'show-v1').write_text('owner     @system_adamic_compiler (direct)\n    Branches: cloud/land-*views-v1* compiler/views-v1* compiler/views-rehearsal\n')
        (self.root / 'show-v2').write_text('owner     @system_adamic_compiler (direct)\n    Branches: cloud/land-*views-v2* compiler/views-v2* compiler/views-rehearsal\n')
        (self.root / 'chain-first-seen.json').write_text(json.dumps({'kvmcfr1': self.now - 3000, 'v1': self.now - 3000, 'v2': self.now - 3000}))
        (self.root / 'seen').write_text('cloud/land-train-1r-views-slice1 %s\n' % star)
        (self.root / 'full-running').write_text(star + '\n')
        # One rehearsal push five minutes ago, V1's slice: V1 is alive, V2 isn't.
        (self.root / 'branch-commits').write_text('%d compiler/views-rehearsal %s views-v1\n%d compiler/views-rehearsal %s\n' % (self.now - 300, other, self.now - 100, 'e' * 40))
        sends = self.check()
        self.assertEqual([line.split('|')[0] for line in sends], ['system_adamic_compiler', 'system_adamic'])
        self.assertIn('#v2 is on the critical path', sends[0])
        # V2's own slice pushed: quiet ends.
        with (self.root / 'branch-commits').open('a') as handle:
            handle.write('%d compiler/views-rehearsal %s views-v2\n' % (self.now - 60, 'f' * 40))
        self.assertEqual(len(self.check()), 2)
        self.assertFalse((self.root / 'chain-quiet-alarmed-v2').exists())

    def test_a_train_candidate_s_whole_gate_is_never_main_and_a_superseded_one_is_history(self):
        # Oct 8 23:19Z: 'Main 3e88a5766db1 is red' (a superseded train-2, gated by the loop between mains), and a page
        # for V1 off superseded 5b7bd612's red while the current candidate ran.
        main, candidate, current = 'a' * 40, 'b' * 40, 'c' * 40
        (self.root / 'main-head').write_text(main + '\n')
        (self.root / 'full.log').write_text('22:00:00 full gate of main %s (tools t) on home\n22:40:00 green: %s full gate in 2400 s\n'
                                            '23:18:00 red: %s first failure at tests after 87 s (still running for triage)\n' % (main, main, candidate))
        (self.root / 'seen').write_text('cloud/land-train-2-views-slice1-old %s\ncloud/land-train-2-views-slice1-new %s\n' % (candidate, current))
        (self.root / 'requests').write_text(current + '\n')
        (self.root / 'watch.log').write_text('23:10:00 done cloud/land-train-2-views-slice1-old: red: %s fast gate, first failure at tests\n' % candidate)
        self.assertEqual(self.check(), [], 'neither the candidate red nor the superseded one pages')
        # The current candidate's own red still pages.
        with (self.root / 'watch.log').open('a') as handle:
            handle.write('23:20:00 done cloud/land-train-2-views-slice1-new: red: %s fast gate, first failure at tests\n' % current)
        sends = self.check()
        self.assertEqual(len(sends), 2)
        self.assertIn(current[:12], sends[0])
        self.assertNotIn('Main', sends[0])

    def test_a_train_slice_waits_on_its_whole_gate_not_its_fast_green(self):
        # Oct 9 00:04Z: 6b2c73f9's fast gate went green and the alarm paged 'green and not landed' while its whole
        # gate, the record it lands on, still ran on the Threadripper.
        (self.root / 'seen').write_text('cloud/land-train-2-views-slice1-x %s\n' % star)
        (self.root / 'watch.log').write_text('%s done cloud/land-train-2-views-slice1-x: green: %s fast gate in 627 s\n' % (self.clock(400), star))
        (self.root / 'whole-gates').write_text('%s running %d\n' % (star, self.now - 900))
        self.assertEqual(self.check(), [], 'a running whole gate is the slice\'s turn')
        # Green whole gate under a minute ago: still the lander's minute.
        (self.root / 'whole-gates').write_text('%s green %d\n' % (star, self.now - 30))
        self.assertEqual(self.check(), [])
        # Green and unlanded past a minute: the handoff pages integration.
        (self.root / 'whole-gates').write_text('%s green %d\n' % (star, self.now - 200))
        sends = self.check()
        self.assertEqual([line.split('|')[0] for line in sends], ['system_adamic_integration', 'system_adamic'])
        self.assertIn('green and not landed', sends[0])

    def clock(self, ago):
        return time.strftime('%H:%M:%S', time.gmtime(self.now - ago))

    def test_a_green_star_waits_a_minute_then_pages_the_lander_and_a_landing_quiets_it(self):
        # @system_adamic, Oct 8: a green tip nobody has landed is a stalled turn; the handoff pages integration.
        (self.root / 'seen').write_text('cloud/land-views-slice1 %s\n' % star)
        (self.root / 'watch.log').write_text('%s done cloud/land-views-slice1: green: %s fast gate in 900 s\n' % (self.clock(30), star))
        self.assertEqual(self.check(), [], 'thirty seconds green is not yet a stalled handoff')
        (self.root / 'watch.log').write_text('%s done cloud/land-views-slice1: green: %s fast gate in 900 s\n' % (self.clock(400), star))
        sends = self.check()
        self.assertEqual([line.split('|')[0] for line in sends], ['system_adamic_integration', 'system_adamic'])
        self.assertIn('green and not landed', sends[0])
        (self.root / 'landed').write_text(star + '\n')
        self.assertEqual(len(self.check()), 2)
        self.assertFalse((self.root / 'star-idle-alarmed').exists())

    def test_a_running_star_is_quiet(self):
        (self.root / 'running/9').write_text('compiler/area-stack %s B threadripper B token log\n' % other)
        self.queueStar(waited=4000)
        self.assertEqual(self.check(), [], 'a gate running on one of its branches is a live turn')


    def test_a_step_that_isnt_ready_waits_on_the_steps_before_it(self):
        # Compiler, Oct 9: V5 paged quiet while it waited on V4. Not ready on the waterfall, it isn't its worker's quiet.
        waterfall = {'nodes': [{'id': 'a03mesg', 'wave': 0}, {'id': 'v5', 'wave': 1, 'branches': ['compiler/views-v5*']}],
                     'criticalPath': ['a03mesg', 'v5'], 'ready': ['a03mesg']}
        (self.root / 'waterfall.json').write_text(json.dumps(waterfall))
        (self.root / 'show-v5').write_text('owner     @system_adamic_compiler (direct)\n')
        (self.root / 'chain-first-seen.json').write_text(json.dumps({'a03mesg': self.now, 'v5': self.now - 3600}))
        self.assertEqual(self.check(), [])
        # Ready, the same step pages, its globs read from the waterfall's branches field.
        waterfall['ready'] = ['a03mesg', 'v5']
        (self.root / 'waterfall.json').write_text(json.dumps(waterfall))
        sends = self.check()
        self.assertIn('#v5 is on the critical path and its worker is quiet', sends[0])
        self.assertIn('compiler/views-v5*', sends[0])

    def test_a_step_in_its_pre_gate_is_not_quiet(self):
        # @system_adamic, Oct 9 01:45Z: V2's complete candidate waited in Loom's pre-gate, and the alarm paged its worker.
        waterfall = {'nodes': [{'id': 'a03mesg', 'wave': 0}, {'id': 'v2', 'wave': 1}], 'criticalPath': ['a03mesg', 'v2']}
        (self.root / 'waterfall.json').write_text(json.dumps(waterfall))
        (self.root / 'show-v2').write_text('owner     @system_adamic_compiler (direct)\n    Branches: compiler/views-v2*\n')
        (self.root / 'chain-first-seen.json').write_text(json.dumps({'a03mesg': self.now, 'v2': self.now - 3600}))
        (self.root / 'branch-commits').write_text('%d compiler/views-v2 %s\n' % (self.now - 1500, other))
        (self.root / 'pregate').mkdir()
        (self.root / 'pregate' / other).write_text('running\npre-gate of %s (the whole Go test set) on codex\n' % other)
        self.assertEqual(self.check(), [], 'a step whose commit is in its pre-gate has a live turn')
        # A finished pre-gate is no longer a turn: quiet pages.
        (self.root / 'pregate' / other).write_text('red: 3 units failed\n')
        self.assertIn('#v2 is on the critical path and its worker is quiet', self.check()[0])

    def mainIsRed(self):
        main = 'e' * 40
        (self.root / 'full.log').write_text('published gate-logs/eeeeeeeeeeee/20261008T202422Z/full-main (abc)\n'
                                            '20:35:28 red: %s first failure at tests after 597.9 s (still running for triage)\n' % main)
        (self.root / 'main-head').write_text(main + '\n')
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
        # A green fix-forward unlanded after a minute is a stalled handoff: it pages the lander, not the author
        # (Oct 8 20:58Z: e96b7d43 green, unlanded).
        (self.root / 'running/77').unlink()
        (self.root / 'seen').write_text('cloud/land-split-fix %s\n' % other)
        (self.root / 'watch.log').write_text('%s done cloud/land-split-fix: green: %s fast gate in 364.2 s\n' % (self.clock(400), other))
        sends = self.check()
        self.assertEqual([line.split('|')[0] for line in sends[2:]], ['system_adamic_integration', 'system_adamic'])
        self.assertIn('green but not landed', sends[2])
        (self.root / 'running/77').write_text('cloud/land-split-fix %s S server S token log\n' % other)
        # An explained red (LIFTED, FIXED FORWARD, CLOSED) is not paged.
        (self.root / 'running/77').unlink()
        (self.root / 'main-reds.tsv').write_text('gate-logs/eeeeeeeeeeee/20261008T202422Z/full-main\tFIXED FORWARD by compiler\n')
        self.assertEqual(len(self.check()), 4)

    def test_a_green_main_is_quiet(self):
        (self.root / 'full.log').write_text('20:35:28 red: %s first failure at tests\n21:30:00 green: %s full gate in 3000 s\n' % ('e' * 40, 'f' * 40))
        self.assertEqual(self.check(), [])


    def test_a_main_with_no_whole_gate_started_within_a_minute_pages_developer_tools(self):
        # Oct 8: main 54cbc125 landed with the full-gate loop paused; nothing confirmed it until a hand found it.
        head = '5' * 40
        (self.root / 'main-head').write_text(head + '\n')
        (self.root / 'main-head-first-seen').write_text('%s %d\n' % (head, self.now - 30))
        self.assertEqual(self.check(), [], 'thirty seconds is not yet a stalled confirmation')
        (self.root / 'main-head-first-seen').write_text('%s %d\n' % (head, self.now - 120))
        sends = self.check()
        self.assertEqual([line.split('|')[0] for line in sends], ['system_adamic_developer_tools', 'system_adamic'])
        self.assertIn('no whole gate has started', sends[0])
        with open(self.root / 'full.log', 'a') as handle:
            handle.write('21:01:48 full gate of main %s (tools t) on home\n' % head)
        self.assertEqual(len(self.check()), 2)
        self.assertFalse((self.root / 'main-confirm-alarmed').exists())


if __name__ == '__main__':
    unittest.main()
