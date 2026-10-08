#!/usr/bin/env python3
"""A stub ahra on PATH and local Git remotes only; no provider or network calls."""
import contextlib
import csv
import importlib.util
import io
import json
import os
from pathlib import Path
import plistlib
import re
import subprocess
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('floor', Path(__file__).with_name('fleet-floor.py'))
floor = importlib.util.module_from_spec(spec)
spec.loader.exec_module(floor)

STUB = r'''#!/usr/bin/env python3
import json, os, pathlib, sys
root = pathlib.Path(os.environ['STUB_ROOT'])
args = sys.argv[1:]
with (root / 'calls').open('a') as f:
    f.write(json.dumps(args) + '\n')
world = json.loads((root / 'world.json').read_text())
if args[:2] == ['ai', 'fleet']:
    assert args[3:] == ['--json', '--live'], args
    selector = args[2]
    if world.get('fleet_error'):
        sys.exit(1)
    values = [m for m in world['members'] if
              (m['fleet'].startswith(selector[:-1]) if selector.endswith('*') else m['fleet'] == selector)]
    print(json.dumps(values))
elif args[:2] == ['ai', 'usage']:
    assert args == ['ai', 'usage', '--provider', 'codex', '--json'], args
    print(json.dumps(world['usage']))
elif args[:2] == ['ai', 'send']:
    assert args[3] == '--message-file', args
    assert pathlib.Path(args[4]).is_file()
    for member in world['members']:
        if member['session']['sessionId'] == args[2]:
            assert member['session']['status'] in ['Idle', 'Completed']
            member['session']['status'] = 'Running'
            break
    else:
        raise AssertionError('unknown session')
elif args[:2] == ['ai', 'start']:
    assert args[2] == 'codex' and args[3] == '--directory' and args[5] == '--label'
    assert args[7] == '--fleet' and args[9] == '--prompt-file', args
    assert pathlib.Path(args[10]).is_file()
    if world.get('dispatch_error'):
        sys.exit(1)
    session = '01234567-0000-0000-0000-%012d' % len(world['members'])
    world['members'].append({'provider': 'Codex', 'fleet': args[8], 'label': args[6],
                              'session': {'sessionId': session, 'status': 'Running'}})
    print('Starting a Codex cloud session on ' + args[4] + '...')
    print(session + '  kirk@kirkouimet.com  in adamic')
    print('https://chatgpt.com/local/' + session)
    print('In fleet ' + args[8] + '.')
elif args[:2] == ['os', 'send']:
    import re
    assert args[4:] == ['--from', 'system_adamic_developer_tools'], args
    assert not re.search('[A-Z]{3,}', args[3]), args
else:
    raise AssertionError('unexpected command: ' + repr(args))
(root / 'world.json').write_text(json.dumps(world))
'''


def member(key, status, label='', fleet='devtools', provider='Codex'):
    return {'provider': provider, 'fleet': fleet, 'label': label,
            'session': {'sessionId': key, 'status': status}, 'statusStale': False,
            'pollError': None}


def usage(credits=62500, percent=100):
    return {'rules': {'creditThresholdPercent': 20, 'emptyWithinHours': 24},
            'accounts': [{'provider': 'Codex', 'account': 'kirk@kirkouimet.com', 'error': None,
                          'meters': [{'name': 'credits', 'unit': 'Credits', 'remaining': credits},
                                     {'name': '168h window', 'unit': 'Percent', 'remaining': percent}]}]}


class DispatchTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.state = self.root / 'state'
        self.queue = self.state / 'queues/developer_tools'
        self.queue.mkdir(parents=True)
        binary = self.root / 'bin'
        binary.mkdir()
        (binary / 'ahra').write_text(STUB)
        (binary / 'ahra').chmod(0o755)
        self.env = patch.dict(os.environ, {'PATH': str(binary) + os.pathsep + os.environ['PATH'],
                                          'STUB_ROOT': str(self.root)})
        self.env.start()
        self.addCleanup(self.env.stop)
        self.lane = {'lane': 'developer_tools', 'fleets': ['devtools', 'devtools-*'],
                     'floor': 4, 'owner': 'system_adamic_developer_tools'}
        self.world(members=[])

    def world(self, members=None, **kwargs):
        value = {'members': members or [], 'usage': usage()}
        value.update(kwargs)
        (self.root / 'world.json').write_text(json.dumps(value))

    def briefs(self, count, label=None):
        for i in range(count):
            (self.queue / ('%02d.md' % i)).write_text(('Label: ' + label + '\n' if label else '') + 'Do work.\n')

    def calls(self, prefix):
        path = self.root / 'calls'
        calls = [json.loads(line) for line in path.read_text().splitlines()] if path.exists() else []
        return [args for args in calls if args[:len(prefix)] == prefix]

    def rows(self):
        return [json.loads(p.read_text()) for p in (self.state / 'rows').glob('*.json')]

    def step(self, dry=False):
        with patch.object(floor.Records, 'flush') as flush, contextlib.redirect_stdout(io.StringIO()):
            floor.one_pass(self.state, [self.lane], Path('/checkout'), 5000, dry=dry)
        flush.assert_called_once()

    def test_idle_label_match_and_completed_reuse(self):
        self.world([member('idle', 'Completed', 'other'), member('match', 'Idle', 'target'),
                    member('running', 'Running')])
        self.briefs(3, 'target')
        self.step()
        self.assertEqual([c[2] for c in self.calls(['ai', 'send'])], ['match', 'idle'])
        self.assertEqual(len(self.calls(['ai', 'start'])), 1)
        self.assertEqual(len(list((self.state / 'used/developer_tools').iterdir())), 3)
        self.assertEqual(list(self.queue.iterdir()), [])
        self.assertTrue(all(p.read_text().startswith('Label: target') for p in
                            (self.state / 'used/developer_tools').iterdir()))
        self.step()
        self.assertEqual(len(self.calls(['ai', 'send'])), 2)
        self.assertEqual(len(self.calls(['ai', 'start'])), 1)

    def test_failed_cancelled_and_other_fleet_idle_are_not_retasked(self):
        self.world([member('failed', 'Failed'), member('cancelled', 'Cancelled'),
                    member('other', 'Idle', fleet='devtools-other')])
        self.briefs(1)
        self.step()
        self.assertFalse(self.calls(['ai', 'send']))
        call = self.calls(['ai', 'start'])[0]
        self.assertEqual(call[4:9], ['/checkout', '--label', '00.md', '--fleet', 'devtools'])
        row = next(r for r in self.rows() if r[-1] == 'started')
        self.assertTrue(row[6].startswith('01234567-'))

    def test_cap_live_prefix_and_provider_count(self):
        self.lane['floor'] = 10
        self.world([member('same', 'Running'), member('same', 'Running', fleet='devtools-sub'),
                    member('second', 'Running', fleet='devtools-extra'),
                    member('claude', 'Running', provider='Claude')])
        self.briefs(8)
        self.step()
        self.assertEqual(len(self.calls(['ai', 'start'])), 3)
        self.assertEqual(len(list(self.queue.iterdir())), 5)
        self.assertIn(2, [r[2] for r in self.rows() if r[-1] == 'started'])
        self.assertIn('pass_cap', [r[-1] for r in self.rows()])
        self.assertEqual([c[2] for c in self.calls(['ai', 'fleet'])], ['devtools', 'devtools-*'])
        self.assertFalse(self.calls(['ai', 'list']))

    def test_credit_and_percent_floors_and_unknown_usage_fail_closed(self):
        self.briefs(1)
        for data, outcome in [(usage(4999), 'credit_floor'), (usage(percent=19), 'percent_floor'),
                              (usage(credits=None), 'usage_unavailable')]:
            with self.subTest(outcome=outcome):
                self.world(usage=data)
                self.step()
                self.assertTrue(any(r[-1].startswith(outcome) for r in self.rows()))
        self.assertFalse(self.calls(['ai', 'start']))
        self.assertTrue((self.queue / '00.md').exists())
        self.world(usage=usage(5000, 20))
        self.step()
        self.assertEqual(len(self.calls(['ai', 'start'])), 1)

    def test_queue_notifications_and_refill_rearms(self):
        (self.state / 'off').touch()
        self.briefs(1)
        self.step()
        self.step()
        messages = self.calls(['os', 'send'])
        self.assertEqual(len(messages), 1)
        self.assertIn('1 briefs', messages[0][3])
        (self.queue / '00.md').unlink()
        self.step()
        self.step()
        self.assertEqual(len(self.calls(['os', 'send'])), 2)
        self.assertIn('empty (0 briefs)', self.calls(['os', 'send'])[1][3])
        self.briefs(3)  # strictly past half of floor, and past empty
        self.step()
        self.assertEqual(len(self.calls(['os', 'send'])), 2)
        for p in self.queue.iterdir():
            p.unlink()
        self.step()
        self.step()
        self.assertEqual(len(self.calls(['os', 'send'])), 4)
        self.assertTrue(all(c[2] == self.lane['owner'] for c in self.calls(['os', 'send'])))

    def test_floor_stops_before_cap_and_used_name_is_not_overwritten(self):
        self.world([member('one', 'Running'), member('two', 'Running'),
                    member('three', 'Running')])
        self.briefs(3)
        self.step()
        self.assertEqual(len(self.calls(['ai', 'start'])), 1)
        self.assertEqual(len(list(self.queue.iterdir())), 2)
        used = self.state / 'used/developer_tools/00.md'
        retained = used.read_bytes()
        self.world(members=[])
        (self.queue / '00.md').write_text('Replacement with an already used name.')
        self.step()
        self.assertEqual(used.read_bytes(), retained)
        self.assertTrue((self.queue / '00.md').exists())
        self.assertIn('already_used', [r[-1] for r in self.rows()])

    def test_empty_warning_requires_lane_under_floor(self):
        self.world([member(str(i), 'Running') for i in range(4)])
        self.step()
        self.assertEqual(len(self.calls(['os', 'send'])), 1)
        self.assertNotIn('empty', self.calls(['os', 'send'])[0][3])
        self.world(members=[])
        self.step()
        self.assertEqual(len(self.calls(['os', 'send'])), 2)
        self.assertIn('empty', self.calls(['os', 'send'])[1][3])

    def test_off_records_without_dispatch(self):
        self.briefs(4)
        (self.state / 'off').touch()
        self.step()
        self.assertFalse(self.calls(['ai', 'start']))
        self.assertIn('off', [r[-1] for r in self.rows()])
        self.assertEqual(len(list(self.queue.iterdir())), 4)

    def test_dry_run_sends_nothing_moves_nothing_changes_no_state(self):
        self.world([member('idle', 'Idle')])
        self.briefs(1)
        before = {str(p.relative_to(self.state)): p.read_bytes() for p in self.state.rglob('*') if p.is_file()}
        self.step(dry=True)
        self.assertFalse(self.calls(['ai', 'start']))
        self.assertFalse(self.calls(['ai', 'send']))
        self.assertFalse(self.calls(['os', 'send']))
        after = {str(p.relative_to(self.state)): p.read_bytes() for p in self.state.rglob('*') if p.is_file()}
        self.assertEqual(before, after)

    def test_uncertain_dispatch_is_retained_never_replayed(self):
        self.briefs(1)
        self.world(dispatch_error=True)
        self.step()
        self.step()
        self.assertEqual(len(self.calls(['ai', 'start'])), 1)
        self.assertTrue((self.state / 'used/developer_tools/00.md').exists())
        self.assertTrue(any(r[-1].startswith('dispatch_uncertain') for r in self.rows()))

    def test_crash_intent_recovers_and_does_not_send(self):
        self.briefs(1)
        brief = self.queue / '00.md'
        used = self.state / 'used/developer_tools/00.md'
        floor.save(self.state / 'inflight/one.json', {'lane': self.lane, 'running': 0, 'queued': 1,
                   'brief': str(brief), 'used': str(used), 'session': ''})
        self.step()
        self.assertTrue(used.exists())
        self.assertFalse(brief.exists())
        self.assertFalse(self.calls(['ai', 'start']))
        self.assertIn('dispatch_uncertain', [r[-1] for r in self.rows()])

    def test_stale_live_fleet_skips(self):
        stale = member('idle', 'Idle')
        stale['statusStale'] = True
        self.world([stale])
        self.briefs(1)
        self.step()
        self.assertFalse(self.calls(['ai', 'send']))
        self.assertTrue(any(r[-1].startswith('fleet_unavailable') for r in self.rows()))


class RecordTests(unittest.TestCase):
    def test_failed_push_retry_and_post_push_crash_do_not_lose_or_duplicate_rows(self):
        with tempfile.TemporaryDirectory() as tmp, contextlib.redirect_stdout(io.StringIO()):
            root = Path(tmp)
            remote, repo, state = root / 'origin.git', root / 'repo', root / 'state'
            floor.run('git', 'init', '--bare', remote)
            floor.run('git', 'init', repo)
            floor.run('git', '-C', repo, 'remote', 'add', 'origin', remote)
            records = floor.Records(state, repo=repo)
            lane = {'lane': 'test', 'floor': 4}
            row = records.row(lane, 0, 0, brief='comma,quote".md', outcome='off')
            hook = remote / 'hooks/pre-receive'
            hook.write_text('#!/bin/sh\nexit 1\n')
            hook.chmod(0o755)
            records.flush()
            self.assertEqual(len(list(records.pending.glob('*.json'))), 1)
            self.assertTrue(records.transaction.exists())
            hook.unlink()
            # Simulate push success and a crash before local acknowledgement.
            transaction = floor.load(records.transaction, {})
            records.wg('push', 'origin', transaction['commit'] + ':refs/heads/' + floor.BRANCH)
            records.flush()
            self.assertEqual(list(records.pending.glob('*.json')), [])
            contents = floor.run('git', '--git-dir', remote, 'show', floor.BRANCH + ':' + floor.CSV)
            table = list(csv.reader(io.StringIO(contents)))
            self.assertEqual(table, [floor.FIELDS, list(map(str, row))])
            records.row(lane, 1, 1, outcome='sent')
            records.flush()
            contents = floor.run('git', '--git-dir', remote, 'show', floor.BRANCH + ':' + floor.CSV)
            self.assertEqual(len(list(csv.reader(io.StringIO(contents)))), 3)
            message = floor.run('git', '--git-dir', remote, 'log', '-1', '--format=%an <%ae>%n%B', floor.BRANCH)
            self.assertIn('kirkouimet <kirk@kirkouimet.com>', message)
            self.assertIn('Co-Authored-By: Ahra <ahra@ahra.ai>', message)

    def test_initial_rejected_push_retries_with_new_rows(self):
        with tempfile.TemporaryDirectory() as tmp, contextlib.redirect_stdout(io.StringIO()):
            root = Path(tmp)
            remote, repo = root / 'origin.git', root / 'repo'
            floor.run('git', 'init', '--bare', remote)
            floor.run('git', 'init', repo)
            floor.run('git', '-C', repo, 'remote', 'add', 'origin', remote)
            records = floor.Records(root / 'state', repo=repo)
            lane = {'lane': 'test', 'floor': 4}
            records.row(lane, 0, 0, outcome='first')
            hook = remote / 'hooks/pre-receive'
            hook.write_text('#!/bin/sh\\nexit 1\\n')
            hook.chmod(0o755)
            records.flush()
            self.assertTrue(records.transaction.exists())
            hook.unlink()
            records.row(lane, 1, 1, outcome='second')
            records.flush()
            self.assertEqual(list(records.pending.glob('*.json')), [])
            contents = floor.run('git', '--git-dir', remote, 'show', floor.BRANCH + ':' + floor.CSV)
            self.assertEqual([r[-1] for r in list(csv.reader(io.StringIO(contents)))[1:]], ['first', 'second'])

    def test_retry_after_rejection_preserves_remote_concurrent_rows(self):
        with tempfile.TemporaryDirectory() as tmp, contextlib.redirect_stdout(io.StringIO()):
            root = Path(tmp)
            remote, repo = root / 'origin.git', root / 'repo'
            floor.run('git', 'init', '--bare', remote)
            floor.run('git', 'init', repo)
            floor.run('git', '-C', repo, 'remote', 'add', 'origin', remote)
            one = floor.Records(root / 'one', repo=repo)
            two = floor.Records(root / 'two', repo=repo)
            lane = {'lane': 'test', 'floor': 4}
            one.row(lane, 0, 2, outcome='first')
            one.flush()
            hook = remote / 'hooks/pre-receive'
            hook.write_text('#!/bin/sh\nexit 1\n')
            hook.chmod(0o755)
            one.row(lane, 1, 1, outcome='retry')
            one.flush()
            hook.unlink()
            two.row(lane, 2, 0, outcome='concurrent')
            two.flush()
            one.flush()
            contents = floor.run('git', '--git-dir', remote, 'show', floor.BRANCH + ':' + floor.CSV)
            self.assertEqual([r[-1] for r in list(csv.reader(io.StringIO(contents)))[1:]],
                             ['first', 'concurrent', 'retry'])
            self.assertEqual(list(one.pending.glob('*.json')), [])


class ConfigurationTests(unittest.TestCase):
    def test_lanes_and_launchd(self):
        configured = floor.lanes(Path(__file__).with_name('fleet-floor') / 'lanes.tsv')
        self.assertEqual([l['floor'] for l in configured], [30, 30, 10, 8, 8, 4])
        with (Path(__file__).with_name('fleet-floor') / 'com.adamic.fleet-floor.plist').open('rb') as file:
            plist = plistlib.load(file)
        for flag in ['KeepAlive', 'RunAtLoad', 'AbandonProcessGroup']:
            self.assertIs(plist[flag], True)
        self.assertEqual(plist['ThrottleInterval'], 60)
        self.assertTrue(plist['EnvironmentVariables']['PATH'].startswith('/Users/kirkouimet/Projects/ahra/node_modules/.bin:'))


if __name__ == '__main__':
    unittest.main()
