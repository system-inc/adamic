#!/usr/bin/env python3
"""Offline wiring tests; --integration also runs integration's live-origin routing test.

No test sends real worker/Circle messages or pushes an area.
"""
import importlib.util
import json
import os
import shlex
from pathlib import Path
import subprocess
import sqlite3
import sys
import tempfile
import unittest
from unittest.mock import patch

INTEGRATION = '--integration' in sys.argv
if INTEGRATION:
    sys.argv.remove('--integration')

spec = importlib.util.spec_from_file_location('auto_area_merge', Path(__file__).with_name('auto-area-merge.py'))
auto = importlib.util.module_from_spec(spec)
spec.loader.exec_module(auto)


class Wiring(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.root = Path(self.temporary.name)
        self.state = self.root / 'state'
        self.state.mkdir()
        self.state_patch = patch.object(auto, 'STATE', self.state)
        self.state_patch.start()
        self.database_patch = patch.dict('os.environ', {'ADAMIC_AI_DATABASE': str(self.root / 'missing-ai.db')})
        self.database_patch.start()
        self.integration = self.root / 'cloud/integration'
        self.integration.mkdir(parents=True)
        (self.integration / 'area-route.py').write_text("def areaNames():\n return {'compiler', 'platforms'}\ndef route(branch):\n if branch == 'codex/host-exception': return ('compiler', 'override from integration')\n if branch.startswith('codex/host-'): return ('platforms', 'prefix from integration')\n return ('hold', 'ask @system_adamic_compiler and system_adamic_typescript')\n")

    def tearDown(self):
        self.database_patch.stop()
        self.state_patch.stop()
        self.temporary.cleanup()

    def test_routing_delegates_to_trusted_integration(self):
        for branch, expected in [('codex/host-exception', 'compiler'),
                                 ('codex/host-new', 'platforms'),
                                 ('codex/unknown', 'hold')]:
            self.assertEqual(auto.route(branch, self.integration)[0], expected)

    def test_switch_off_never_fetches_or_merges(self):
        with patch('sys.argv', ['auto', '--drain']), patch.object(auto, 'checkout') as checkout:
            auto.main()
            checkout.assert_not_called()
        self.assertFalse((self.state / 'auto-area-merge').exists())

    def test_locked_queue_retries_and_completed_queue_deduplicates(self):
        (self.state / 'auto-area-merge').touch()
        branch, sha = 'codex/host-new', 'a' * 40
        with patch('sys.argv', ['auto', '--enqueue', branch, sha]):
            auto.main()
        with patch.object(auto, 'checkout', return_value=self.root), patch.object(auto, 'publish_records'), patch.object(auto, 'hold_job'), patch.object(auto, 'attempt', return_value='locked') as attempt:
            with patch('sys.argv', ['auto', '--drain']):
                auto.main()
            self.assertEqual(len(list((self.state / 'area-queue').glob('*.held'))), 1)
            attempt.return_value = 'merged'
            with patch('sys.argv', ['auto', '--drain']):
                auto.main()
            self.assertEqual(attempt.call_count, 2)
        with patch('sys.argv', ['auto', '--enqueue', branch, sha]):
            auto.main()
        self.assertEqual(list((self.state / 'area-queue').glob('*.json')), [])

    def test_dry_run_passes_no_push_and_never_notifies(self):
        result = subprocess.CompletedProcess([], 3, 'conflict merging branch:\ninternal/foo.go\n')
        with patch.object(auto, 'remote_merge', return_value=result) as remote, patch.object(auto, 'git', return_value='b' * 40 + '\tref'), patch.object(auto, 'notify') as notify:
            self.assertEqual(auto.attempt(self.root, 'codex/host-new', 'a' * 40, True), 'conflict')
            self.assertTrue(remote.call_args.args[4])
            notify.assert_not_called()
        row = (self.state / 'dry-run' / auto.RECORD_FILE).read_text()
        self.assertIn(',platforms,conflict,', row)
        self.assertFalse((self.state / 'auto-area-merge').exists())

    def test_unknown_does_not_run_merge(self):
        with patch.object(auto, 'run') as run:
            self.assertEqual(auto.attempt(self.root, 'codex/unknown', 'a' * 40, True), 'no-area')
            run.assert_not_called()

    def test_red_returns_details_to_worker_and_journals_result(self):
        result = subprocess.CompletedProcess([], 1, 'new failures from merging branch:\n  package TestFailure\n')
        with patch.object(auto, 'remote_merge', return_value=result), patch.object(auto, 'git', return_value='b' * 40 + '\tref'), patch.object(auto, 'notify') as notify:
            self.assertEqual(auto.attempt(self.root, 'codex/host-new', 'a' * 40, False), 'red')
            self.assertIn('TestFailure', notify.call_args.args[4])
        rows = list((self.state / 'area-records-outbox').glob('*.json'))
        self.assertEqual(json.loads(rows[0].read_text())[4], 'red')


    def test_records_failed_push_retries_without_losing_or_duplicating_row(self):
        source = self.root / 'source'
        remote = self.root / 'origin.git'
        source.mkdir()
        subprocess.run(['git', 'init', '-q', str(source)], check=True)
        subprocess.run(['git', 'init', '-q', '--bare', str(remote)], check=True)
        auto.git('remote', 'add', 'origin', str(remote), cwd=source)
        auto.git('commit', '-q', '--allow-empty', '-m', 'Initial', cwd=source)
        (self.state / 'auto-area-merge').touch()
        row = ['2026-10-08T00:00:00Z', 'codex/host-new', 'a' * 40, 'platforms', 'merged', '1.0', 'b' * 40]
        real_run = auto.run
        def fail_push(*args, **kwargs):
            if args[:2] == ('git', 'push'):
                return subprocess.CompletedProcess(args, 1, 'simulated origin failure')
            return real_run(*args, **kwargs)
        with patch.object(auto, 'HERE', source):
            auto.record(row, False)
            with patch.object(auto, 'run', side_effect=fail_push):
                auto.publish_records()
            self.assertEqual(len(list((self.state / 'area-records-outbox').glob('*.json'))), 1)
            auto.publish_records()
            self.assertEqual(list((self.state / 'area-records-outbox').glob('*.json')), [])
            # Replay after a crash following a successful push: exactly one CSV row.
            auto.record(row, False)
            auto.publish_records()
        contents = auto.git('show', auto.RECORD_REF + ':' + auto.RECORD_FILE, cwd=remote)
        self.assertEqual(contents.count('codex/host-new'), 1)

    def test_session_lookup_and_message_file_use_exact_branch_text(self):
        database = self.root / 'ai.db'
        with sqlite3.connect(database) as db:
            db.execute('create table replies (session_id text, text text, fetched_at integer)')
            db.execute('insert into replies values (?, ?, ?)', ('correct', 'pushed codex/host_new', 1))
            db.execute('insert into replies values (?, ?, ?)', ('wrong', 'pushed codex/hostXnew', 2))
        logfile = self.root / 'merge.log'
        with patch.dict('os.environ', {'ADAMIC_AI_DATABASE': str(database)}), patch.object(auto, 'run', return_value=subprocess.CompletedProcess([], 0, 'sent')) as send:
            auto.notify('codex/host_new', 'a' * 40, 'platforms', 'conflict', 'conflict:\ninternal/foo.go\n', logfile)
            self.assertEqual(send.call_args.args[:4], ('ahra', 'ai', 'send', 'correct'))
            self.assertEqual(send.call_args.kwargs['cwd'], '/Users/kirkouimet/Projects/ahra')
            self.assertIn('internal/foo.go', Path(send.call_args.args[-1]).read_text())


    def test_held_job_stays_held_and_reports_to_named_circles_once(self):
        (self.state / 'auto-area-merge').touch()
        branch, sha = 'codex/unknown', 'a' * 40
        with patch('sys.argv', ['auto', '--enqueue', branch, sha]):
            auto.main()
        with patch.object(auto, 'checkout', return_value=self.root), patch.object(auto, 'publish_records'), patch.object(auto, 'run', return_value=subprocess.CompletedProcess([], 0, 'sent')) as send:
            for _ in range(3):
                with patch('sys.argv', ['auto', '--drain']):
                    auto.main()
            self.assertEqual(send.call_count, 3)
            self.assertEqual({call.args[3] for call in send.call_args_list}, {'system_adamic_integration', 'system_adamic_compiler', 'system_adamic_typescript'})
            for call in send.call_args_list:
                self.assertIn(branch, call.args[4])
                self.assertIn(sha, call.args[4])
                self.assertEqual(call.args[-2:], ('--from', 'system_adamic_developer_tools'))
        queue = self.state / 'area-queue'
        self.assertEqual(list(queue.glob('*.done')), [])
        self.assertEqual(list(queue.glob('*.json')), [])
        with patch('sys.argv', ['auto', '--enqueue', branch, sha]):
            auto.main()
        self.assertEqual(list(queue.glob('*.json')), [])
        job = json.loads(next(queue.glob('*.held')).read_text())
        self.assertEqual(job['status'], 'held')
        self.assertTrue(job['reason'])
        self.assertIn(branch, (self.state / 'area-held.tsv').read_text())
        self.assertEqual(len(list((self.state / 'area-records-outbox').glob('*.json'))), 1)
        auto.clear_hold(branch)
        self.assertNotIn(branch, (self.state / 'area-held.tsv').read_text())

    def test_route_change_requeues_held_branch(self):
        (self.state / 'auto-area-merge').touch()
        branch, sha = 'codex/unknown', 'a' * 40
        with patch('sys.argv', ['auto', '--enqueue', branch, sha]):
            auto.main()
        with patch.object(auto, 'checkout', return_value=self.root), patch.object(auto, 'publish_records'), patch.object(auto, 'run', return_value=subprocess.CompletedProcess([], 0, 'sent')):
            with patch('sys.argv', ['auto', '--drain']):
                auto.main()
        with patch.object(auto, 'checkout', return_value=self.root), patch.object(auto, 'publish_records'), patch.object(auto, 'route', return_value=('compiler', 'new authoritative family')), patch.object(auto, 'attempt', return_value='merged') as attempt:
            with patch('sys.argv', ['auto', '--drain']):
                auto.main()
            self.assertEqual(attempt.call_count, 1)
        self.assertEqual(list((self.state / 'area-queue').glob('*.held')), [])
        self.assertEqual(len(list((self.state / 'area-queue').glob('*.done'))), 1)

    def test_failed_hold_report_retries_without_repeating_success(self):
        with patch.object(auto, 'run', side_effect=[subprocess.CompletedProcess([], 1, 'offline'), subprocess.CompletedProcess([], 0, 'sent')]) as send:
            auto.hold_job('codex/unknown', 'a' * 40, 'ask integration')
            auto.hold_job('codex/unknown', 'a' * 40, 'ask integration')
            auto.hold_job('codex/unknown', 'b' * 40, 'ask integration')
            self.assertEqual(send.call_count, 2)

    def test_refusals_and_unknown_exits_are_not_worker_reds(self):
        cases = [(1, 'refused: origin/branch moved; name the tip you mean\n', 'held'),
                 (1, 'refused: worktree has local changes\n', 'held'),
                 (1, 'refused: another merge into area/platforms is running (lock); wait for it\n', 'locked'),
                 (2, 'refused: pinned node missing\n', 'held'),
                 (1, 'fatal: origin failed\n', 'held'),
                 (2, 'unexpected output\n', 'held'),
                 (0, 'refused: local changes\n', 'held'),
                 (3, 'refused: origin moved\nconflict text\n', 'held')]
        for code, output, expected in cases:
            with patch.object(auto, 'remote_merge', return_value=subprocess.CompletedProcess([], code, output)), patch.object(auto, 'git', return_value='b' * 40 + '\tref'), patch.object(auto, 'notify') as notify, patch.object(auto, 'hold_job') as hold:
                self.assertEqual(auto.attempt(self.root, 'codex/host-new', 'a' * 40, False), expected)
                hold.assert_called_once()
                notify.assert_not_called()

    def test_refusal_is_held_without_repeating_merge_each_poll(self):
        (self.state / 'auto-area-merge').touch()
        branch, sha = 'codex/host-new', 'a' * 40
        with patch('sys.argv', ['auto', '--enqueue', branch, sha]):
            auto.main()
        with patch.object(auto, 'checkout', return_value=self.root), patch.object(auto, 'publish_records'), patch.object(auto, 'git', return_value='b' * 40 + '\tref'), patch.object(auto, 'remote_merge', return_value=subprocess.CompletedProcess([], 1, 'refused: worktree has local changes\n')) as remote, patch.object(auto, 'run', return_value=subprocess.CompletedProcess([], 0, 'sent')), patch.object(auto, 'notify') as notify:
            for _ in range(3):
                with patch('sys.argv', ['auto', '--drain']):
                    auto.main()
            self.assertEqual(remote.call_count, 1)
            notify.assert_not_called()
        self.assertEqual(len(list((self.state / 'area-queue').glob('*.held'))), 1)
        self.assertEqual(list((self.state / 'area-queue').glob('*.done')), [])

    def test_merged_sends_one_line_success_to_worker(self):
        path = self.root / 'ai.db'
        with sqlite3.connect(path) as db:
            db.execute('create table replies (session_id text, text text, fetched_at integer)')
            db.execute('insert into replies values (?, ?, ?)', ('worker', 'pushed codex/host-new', 1))
        with patch.dict('os.environ', {'ADAMIC_AI_DATABASE': str(path)}), patch.object(auto, 'run', return_value=subprocess.CompletedProcess([], 0, 'sent')) as send:
            auto.notify('codex/host-new', 'a' * 40, 'platforms', 'merged', 'merge output', self.root / 'merge.log')
            text = Path(send.call_args.args[-1]).read_text()
        self.assertEqual(len(text.splitlines()), 1)
        self.assertIn('codex/host-new', text)
        self.assertIn('area/platforms', text)
        self.assertNotIn('Resolve', text)

    def test_reply_must_name_full_branch_not_longer_branch(self):
        path = self.root / 'ai.db'
        with sqlite3.connect(path) as db:
            db.execute('create table replies (session_id text, text text, fetched_at integer)')
            db.execute('insert into replies values (?, ?, ?)', ('correct', 'pushed codex/host-promises', 1))
            db.execute('insert into replies values (?, ?, ?)', ('wrong', 'pushed codex/host-promises-typeof', 2))
            self.assertEqual(auto.worker_session(db, 'codex/host-promises'), 'correct')

    def test_reaping_enqueues_only_real_matching_green_worker_with_switch(self):
        watcher = Path(__file__).with_name('fast-gate-watch.sh').read_text()
        void_function = watcher[watcher.index('voidCause() {'):watcher.index('\ntips() {')]
        loop = watcher[watcher.index('  for file in "${state}"/running/*; do'):watcher.index('  # Shared with the dispatcher:')]
        sha = 'a' * 40
        cases = [('codex/worker', f'green: {sha} passed\n', True, True),
                 ('codex/worker', f'green: {sha} passed\n', False, False),
                 ('area/compiler', f'green: {sha} passed\n', True, False),
                 ('area-merge/compiler', 'no gate verdict\n', True, False),
                 ('codex/worker', 'Connection reset by peer\n', True, False),
                 ('codex/worker', f'red: {sha} failed\ncreating work dir\n', True, False),
                 ('codex/worker', f'red: {sha} failed\n', True, False),
                 ('codex/worker', 'green: ' + 'b' * 40 + ' passed\n', True, False)]
        for index, (branch, verdict, switch, expected) in enumerate(cases):
            directory = self.root / ('reap-' + str(index))
            (directory / 'running').mkdir(parents=True)
            (directory / 'logs').mkdir()
            (directory / 'cloud').mkdir()
            (directory / 'cloud/auto-area-merge.sh').write_text('echo "$*" >> "${state}/enqueued"\n')
            (directory / 'running/999999999').write_text(f'{branch} {sha} S\n')
            (directory / ('logs/' + sha[:12] + '.log')).write_text(verdict)
            (directory / 'gated').write_text(sha + '\n')
            if switch:
                (directory / 'auto-area-merge').touch()
            script = 'state=$1; here=$1; export state\nrecordWait() { :; }\n' + void_function + '\n' + loop
            result = subprocess.run(['bash', '-c', script, 'reap-test', str(directory)], capture_output=True, text=True)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual((directory / 'enqueued').exists(), expected, (branch, verdict, switch))
            if branch.startswith('area-merge/'):
                self.assertFalse((directory / 'queue').exists())
                self.assertFalse((directory / 'void-tries').exists())
            elif 'creating work dir' in verdict or not verdict.startswith(('green:', 'red:')):
                self.assertIn(sha, (directory / 'queue').read_text())


    def test_unexpected_outcome_never_becomes_done(self):
        (self.state / 'auto-area-merge').touch()
        with patch('sys.argv', ['auto', '--enqueue', 'codex/host-new', 'a' * 40]):
            auto.main()
        with patch.object(auto, 'checkout', return_value=self.root), patch.object(auto, 'publish_records'), patch.object(auto, 'attempt', return_value='unexpected'), patch.object(auto, 'run', return_value=subprocess.CompletedProcess([], 0, 'sent')):
            with patch('sys.argv', ['auto', '--drain']):
                auto.main()
        self.assertEqual(list((self.state / 'area-queue').glob('*.done')), [])
        entry = json.loads(next((self.state / 'area-queue').glob('*.held')).read_text())
        self.assertIn('unexpected', entry['reason'])

    def test_successful_attempt_notifies_worker(self):
        with patch.object(auto, 'remote_merge', return_value=subprocess.CompletedProcess([], 0, 'merged: area/platforms takes branch\n')), patch.object(auto, 'git', return_value='b' * 40 + '\tref'), patch.object(auto, 'notify') as notify:
            self.assertEqual(auto.attempt(self.root, 'codex/host-new', 'a' * 40, False), 'merged')
            self.assertEqual(notify.call_args.args[3], 'merged')

    def test_router_unavailable_is_held_and_import_leaves_no_bytecode(self):
        self.assertEqual(auto.route('codex/host-new', self.integration)[0], 'platforms')
        self.assertFalse((self.integration / '__pycache__').exists())
        (self.integration / 'area-route.py').unlink()
        area, why = auto.route('codex/host-new', self.integration)
        self.assertEqual(area, 'hold')
        self.assertIn('routing unavailable', why)


    def stub_transport(self):
        binaries = self.root / 'bin'
        binaries.mkdir()
        ssh = binaries / 'ssh'
        ssh.write_text('#!' + sys.executable + '\n' + """
import json, os, pathlib, shlex, sys
state = pathlib.Path(os.environ['AUTO_MERGE_TEST_STATE'])
arguments = shlex.split(sys.argv[-1])
entry = state / 'running' / str(os.getpid())
expected = f'area-merge/{arguments[0]} {arguments[2]} B threadripper B\\n'
assert entry.read_text() == expected, entry
assert not (state / 'slot-table.lock').exists()
script = sys.stdin.read()
(state / 'ssh-observed.json').write_text(json.dumps({'argv': sys.argv[1:], 'arguments': arguments, 'script': script, 'entry': str(entry)}))
print('new failures from merging ' + arguments[1] + ':')
print('  package TestFailure')
print('go vet failed: /tmp/area-log/vet.log')
print('logs in /tmp/area-log')
sys.exit(int(os.environ.get('AUTO_MERGE_TEST_EXIT', '1')))
""")
        ssh.chmod(0o755)
        scp = binaries / 'scp'
        scp.write_text('#!' + sys.executable + '\n' + """
import json, os, pathlib, sys
state = pathlib.Path(os.environ['AUTO_MERGE_TEST_STATE'])
assert not list((state / 'running').glob('*')), 'claim must be removed before artifact copy'
(state / 'scp-observed.json').write_text(json.dumps(sys.argv[1:]))
token = sys.argv[-2].rsplit('/', 1)[1]
bundle = pathlib.Path(sys.argv[-1]) / token
(bundle / 'logs/0').mkdir(parents=True)
(bundle / 'logs/0/vet.log').write_text('remote vet diagnostic')
(bundle / 'logs.json').write_text(json.dumps({'/tmp/area-log': 'logs/0'}))
(bundle / 'output.log').write_text('remote merge output')
""")
        scp.chmod(0o755)
        (self.state / 'slots').write_text('threadripper B\nthreadripper S\nworkshop B\n')
        (self.state / 'running').mkdir()
        return {'PATH': str(binaries) + os.pathsep + os.environ['PATH'], 'AUTO_MERGE_TEST_STATE': str(self.state)}

    def test_remote_ssh_claim_copy_and_cleanup_without_real_box(self):
        environment = self.stub_transport()
        with patch.dict('os.environ', environment), patch.object(auto, 'git', return_value='c' * 40):
            result = auto.remote_merge(self.root, 'platforms', 'codex/host-new', 'a' * 40, True)
        observed = json.loads((self.state / 'ssh-observed.json').read_text())
        self.assertEqual(observed['argv'][:4], ['-l', 'ahra', 'cloud', 'bash'])
        self.assertEqual(observed['arguments'][:4], ['platforms', 'codex/host-new', 'a' * 40, 'c' * 40])
        self.assertIn('--no-push', observed['arguments'])
        self.assertIn('source ~/adamic-tools/env.sh', observed['script'])
        self.assertIn('export PATH="${ADAMIC_TYPESCRIPT_SOURCE}/bin:${PATH}"', observed['script'])
        self.assertIn('taskset -c 0-23 bash ~/area-merge/integration/cloud/integration/area-merge.sh', observed['script'])
        self.assertEqual(result.returncode, 1)
        self.assertFalse(Path(observed['entry']).exists())
        self.assertFalse((self.state / 'slot-table.lock').exists())
        self.assertNotIn('/tmp/area-log', result.stdout)
        self.assertIn(str(self.state / 'dry-run'), result.stdout)
        self.assertEqual(len(list((self.state / 'dry-run').glob('*/logs/0/vet.log'))), 1)
        copied = json.loads((self.state / 'scp-observed.json').read_text())
        self.assertIn('User=ahra', copied)

    def test_remote_waits_for_cloud_big_slot_then_claims(self):
        environment = self.stub_transport()
        busy = self.state / 'running/999999998'
        busy.write_text('area/runtime oldsha B threadripper B\n')
        def release(_):
            self.assertFalse((self.state / 'ssh-observed.json').exists())
            busy.unlink()
        with patch.dict('os.environ', environment), patch.object(auto, 'git', return_value='c' * 40), patch.object(auto.time, 'sleep', side_effect=release) as sleep:
            auto.remote_merge(self.root, 'platforms', 'codex/host-new', 'a' * 40, True)
        sleep.assert_called_once()

    def test_cloud_slot_arithmetic_matches_box_and_assigned_class(self):
        self.stub_transport()
        running = self.state / 'running/999999998'
        for text, expected in [('gate sha B workshop B', True),
                               ('gate sha S threadripper B', True),
                               ('gate sha B threadripper B', False),
                               ('gate sha B', False)]:
            running.write_text(text)
            self.assertEqual(auto.cloud_area_free(), expected, text)

    def test_switch_off_while_waiting_does_not_start_ssh(self):
        self.stub_transport()
        with patch.object(auto, 'git', return_value='c' * 40), patch.object(auto.subprocess, 'Popen') as ssh:
            with self.assertRaises(auto.SlotWaitCancelled):
                auto.remote_merge(self.root, 'platforms', 'codex/host-new', 'a' * 40, False)
            ssh.assert_not_called()

    def test_ssh_failure_removes_slot_claim(self):
        environment = self.stub_transport()
        environment['AUTO_MERGE_TEST_EXIT'] = '255'
        with patch.dict('os.environ', environment), patch.object(auto, 'git', return_value='c' * 40):
            result = auto.remote_merge(self.root, 'platforms', 'codex/host-new', 'a' * 40, True)
        self.assertEqual(result.returncode, 255)
        self.assertEqual(list((self.state / 'running').glob('*')), [])

    def test_slot_guard_blocks_ssh_until_watcher_allocation_finishes(self):
        environment = self.stub_transport()
        guard = self.state / 'slot-table.lock'
        guard.mkdir()
        def release(_):
            self.assertFalse((self.state / 'ssh-observed.json').exists())
            guard.rmdir()
        with patch.dict('os.environ', environment), patch.object(auto, 'git', return_value='c' * 40), patch.object(auto.time, 'sleep', side_effect=release) as sleep:
            auto.remote_merge(self.root, 'platforms', 'codex/host-new', 'a' * 40, True)
        sleep.assert_called_once()

    def test_switch_removal_cancels_busy_slot_wait(self):
        self.stub_transport()
        switch = self.state / 'auto-area-merge'
        switch.touch()
        (self.state / 'running/999999998').write_text('gate sha B threadripper B\n')
        with patch.object(auto, 'git', return_value='c' * 40), patch.object(auto.time, 'sleep', side_effect=lambda _: switch.unlink()), patch.object(auto.subprocess, 'Popen') as ssh:
            with self.assertRaises(auto.SlotWaitCancelled):
                auto.remote_merge(self.root, 'platforms', 'codex/host-new', 'a' * 40, False)
            ssh.assert_not_called()
        self.assertFalse((self.state / 'slot-table.lock').exists())


    def test_remote_wrapper_environment_affinity_and_transaction_logs(self):
        home = self.root / 'remote-home'
        tools = home / 'adamic-tools'
        (tools / 'bin').mkdir(parents=True)
        ts = home / 'typescript'
        (ts / 'bin').mkdir(parents=True)
        temporary = home / 'temp'
        temporary.mkdir()
        (tools / 'env.sh').write_text(f'export ADAMIC_TYPESCRIPT_SOURCE={shlex.quote(str(ts))}\nexport TMPDIR={shlex.quote(str(temporary))}\n')
        for directory in ('integration', 'tree'):
            (home / 'area-merge' / directory / '.git').mkdir(parents=True)
        integration = home / 'area-merge/integration/cloud/integration'
        integration.mkdir(parents=True)
        private = home / 'private'
        private.mkdir()
        (private / 'must-not-copy').write_text('unrelated file')
        (integration / 'area-merge.sh').write_text('node -v\nlogs=$(mktemp -d)\necho "vet diagnostic" > "$logs/vet.log"\necho "not merged: logs in $logs"\necho "logs in ${HOME}/private"\nexit 1\n')
        commands = {'id': 'echo ahra\n', 'git': 'exit 0\n', 'node': 'echo v24.19.0\n',
                    'taskset': '[ "$1" = -c ] && [ "$2" = 0-23 ] || exit 9\nshift 2\nexec "$@"\n'}
        for name, body in commands.items():
            path = tools / 'bin' / name
            path.write_text('#!/usr/bin/env bash\n' + body)
            path.chmod(0o755)
        with patch.dict('os.environ', {'HOME': str(home), 'PATH': str(tools / 'bin') + os.pathsep + os.environ['PATH']}):
            result = subprocess.run(['bash', str(Path(__file__).with_name('auto-area-merge-remote.sh')), 'compiler', 'codex/compiler-fix', 'a' * 40, 'c' * 40, 'area-merge/out/test', '--no-push'], text=True, capture_output=True)
        self.assertEqual(result.returncode, 1, result.stderr)
        self.assertIn('v24.19.0', result.stdout)
        bundle = home / 'area-merge/out/test'
        manifest = json.loads((bundle / 'logs.json').read_text())
        self.assertEqual(len(manifest), 1)
        self.assertTrue(next(iter(manifest)).startswith(str(temporary)))
        self.assertNotIn(str(private), manifest)
        self.assertEqual((bundle / next(iter(manifest.values())) / 'vet.log').read_text().strip(), 'vet diagnostic')

    def test_watcher_cannot_schedule_into_claimed_cloud_area_slot(self):
        watcher = Path(__file__).with_name('fast-gate-watch.sh').read_text()
        start = watcher.index('  # Shared with the dispatcher:')
        stop = watcher.index('  # A durable queue survives restarts.', start)
        scheduling = watcher[start:stop]
        (self.state / 'slots').write_text('threadripper B\n')
        running = self.state / 'running'
        running.mkdir()
        (running / '999999999').write_text('area-merge/compiler ' + 'a' * 40 + ' B threadripper B\n')
        queued = 'B 1 codex/compiler-fix ' + 'b' * 40 + '\n'
        (self.state / 'queue').write_text(queued)
        script = 'state=$1; here=$1\n' + scheduling
        result = subprocess.run(['bash', '-c', script, 'schedule-test', str(self.state)], capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual((self.state / 'queue').read_text(), queued)
        self.assertEqual(len(list(running.glob('*'))), 1)
        self.assertFalse((self.state / 'slot-table.lock').exists())

    def test_hold_sender_is_passed_to_ahra_stub(self):
        binaries = self.root / 'bin'
        binaries.mkdir()
        stub = binaries / 'ahra'
        stub.write_text('#!' + sys.executable + '\n' + """
import json, os, pathlib, sys
assert sys.argv[1:4] == ['os', 'send', 'system_adamic_integration']
assert sys.argv[-2:] == ['--from', 'system_adamic_developer_tools']
pathlib.Path(os.environ['AUTO_MERGE_TEST_STATE'], 'hold-delivered.json').write_text(json.dumps(sys.argv[1:]))
""")
        stub.chmod(0o755)
        actual_run = auto.run
        def from_test_directory(*args, **kwargs):
            self.assertEqual(kwargs['cwd'], '/Users/kirkouimet/Projects/ahra')
            kwargs['cwd'] = self.root
            return actual_run(*args, **kwargs)
        with patch.dict('os.environ', {'PATH': str(binaries) + os.pathsep + os.environ['PATH'], 'AUTO_MERGE_TEST_STATE': str(self.state)}), patch.object(auto, 'run', side_effect=from_test_directory):
            auto.hold_job('codex/unknown', 'a' * 40, 'integration decides')
        delivered = json.loads((self.state / 'hold-delivered.json').read_text())
        self.assertIn('codex/unknown', delivered[3])


@unittest.skipUnless(INTEGRATION, 'pass --integration to run integration-owned routing coverage against origin')
class IntegrationRouting(unittest.TestCase):
    def test_integration_owned_routing_coverage(self):
        root = auto.checkout()
        result = auto.run(sys.executable, str(root / 'cloud/integration/area-route-test.py'), cwd=root, check=False)
        print(result.stdout, end='', flush=True)
        self.assertEqual(result.returncode, 0, result.stdout)


if __name__ == '__main__':
    unittest.main()
