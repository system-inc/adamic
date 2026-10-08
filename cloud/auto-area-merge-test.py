#!/usr/bin/env python3
"""Offline wiring tests; never contact origin, workers, or real areas."""
import importlib.util
import json
from pathlib import Path
import subprocess
import sqlite3
import tempfile
import unittest
from unittest.mock import patch

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
        self.integration = self.root / 'cloud/integration'
        self.integration.mkdir(parents=True)
        (self.integration / 'areas.tsv').write_text('compiler\tcompiler-owner\tnone\nplatforms\tplatform-owner\tnone\nstage1-lint\tcohere-owner\tcohere\nstage1-format\tcohere-owner\tcohere\n')
        (self.integration / 'census-owners.tsv').write_text('codex/host-exception\tcompiler-owner\ncodex/ambiguous\tcohere-owner\ncodex/parser-recovery\tsystem_cohere_lint\n')
        (self.integration / 'branch-census.py').write_text("prefixOwners = [('codex/host-', 'platform-owner')]\n")

    def tearDown(self):
        self.state_patch.stop()
        self.temporary.cleanup()

    def test_mapping_override_prefix_unknown_and_ambiguity(self):
        for branch, expected in [('codex/host-exception', ['compiler']),
                                 ('codex/host-new', ['platforms']),
                                 ('codex/compiler-new', ['compiler']),
                                 ('codex/unknown', []),
                                 ('codex/ambiguous', ['stage1-format', 'stage1-lint']),
                                 ('codex/parser-recovery', ['stage1-lint'])]:
            self.assertEqual(auto.areas_for(branch, self.integration), expected)

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
        with patch.object(auto, 'checkout', return_value=self.root), patch.object(auto, 'publish_records'), patch.object(auto, 'attempt', return_value='locked') as attempt:
            with patch('sys.argv', ['auto', '--drain']):
                auto.main()
            self.assertEqual(len(list((self.state / 'area-queue').glob('*.json'))), 1)
            attempt.return_value = 'merged'
            with patch('sys.argv', ['auto', '--drain']):
                auto.main()
            self.assertEqual(attempt.call_count, 2)
        with patch('sys.argv', ['auto', '--enqueue', branch, sha]):
            auto.main()
        self.assertEqual(list((self.state / 'area-queue').glob('*.json')), [])

    def test_dry_run_passes_no_push_and_never_notifies(self):
        result = subprocess.CompletedProcess([], 3, 'conflict merging branch:\ninternal/foo.go\n')
        with patch.object(auto, 'run', return_value=result) as run, patch.object(auto, 'git', return_value='b' * 40 + '\tref'), patch.object(auto, 'notify') as notify:
            self.assertEqual(auto.attempt(self.root, 'codex/host-new', 'a' * 40, True), 'conflict')
            self.assertIn('--no-push', run.call_args.args)
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
        with patch.object(auto, 'run', return_value=result), patch.object(auto, 'git', return_value='b' * 40 + '\tref'), patch.object(auto, 'notify') as notify:
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


if __name__ == '__main__':
    unittest.main()
