import os
import sys
import unittest
from unittest import mock

sys.path.insert(0, os.path.dirname(__file__))
import sort_reds as sort


def unit(name='TestAnswer', hash='h', action='fail', **extra):
    return dict(package='example.com/p', test=name, input_hash=hash,
                input_paths=['p/p.go'], action=action, **extra)


def record(*units, tools='tools', **extra):
    return dict(dict(sha='a' * 40, tools_fingerprint=tools, finished=True, units=list(units)), **extra)


class RedSort(unittest.TestCase):
    def test_same_unit_same_inputs(self):
        result = sort.sort_reds(record(unit()), record(unit()))
        self.assertEqual(len(result['mains']), 1)
        self.assertEqual(result['status'], 'candidate reds: 0')

    def test_wrong_answer_with_changed_inputs_is_candidate(self):
        result = sort.sort_reds(record(unit(hash='new', detail='assertion diff: got 2 want 1')),
                               record(unit()))
        self.assertEqual(result['candidate_reds'], 1)

    def test_different_tools_and_unfinished_main_do_not_excuse(self):
        for main in (record(unit(), tools='old'), record(unit(), finished=False), None):
            result = sort.sort_reds(record(unit()), main)
            self.assertEqual(result['candidate_reds'], 1)
            self.assertEqual(result['status'], 'no main record on these tools')

    def test_missing_hashes_never_match(self):
        result = sort.sort_reds(record(unit(hash=None)), record(unit(hash=None)))
        self.assertEqual(result['candidate_reds'], 1)

    def test_infra_signatures(self):
        for detail in ('killed at 90 s', 'go exited 2', 'no space left on device',
                       'clone failed', 'fetch failure', 'Traceback (most recent call last)',
                       'ssh timed out', 'void selection', 'never reported'):
            with self.subTest(detail=detail):
                result = sort.sort_reds(record(unit(detail=detail)), record())
                self.assertEqual(len(result['infra']), 1)
                self.assertEqual(result['candidate_reds'], 0)

    def test_unreported_unit_is_infra(self):
        result = sort.sort_reds(record(unit(action=None, status='not run')), record())
        self.assertEqual(len(result['infra']), 1)

    def test_untouched_cold_timing_is_mains(self):
        for detail in ('budget', 'deadline', 'killed', 'setup exceeded'):
            result = sort.sort_reds(record(unit(detail=detail, cold_miss=True)), record(),
                                   {'candidate_changed': ['other/x.go']})
            self.assertEqual(len(result['mains']), 1, result)

    def test_touched_or_unknown_cold_timing_needs_rerun_or_candidate(self):
        for context in ({}, {'candidate_changed': ['p/p.go']}):
            result = sort.sort_reds(record(unit(detail='budget', cold_miss=True)), record(), context)
            self.assertEqual(result['candidate_reds'], 1)

    def test_wrong_answers_never_use_timing_exemption(self):
        result = sort.sort_reds(record(unit(detail='assertion diff after deadline on cold miss')),
                               record(), {'candidate_changed': []})
        self.assertEqual(result['candidate_reds'], 1)

    def test_stale_main_fix(self):
        result = sort.sort_reds(record(unit()), record(unit(hash='fixed', action='pass')),
                               {'candidate_changed': ['other/x.go'], 'main_changed': ['p/p.go']})
        self.assertEqual(len(result['stale']), 1)
        self.assertIn("recut, don't land", result['status'])

    def test_candidate_edit_does_not_masquerade_as_stale(self):
        result = sort.sort_reds(record(unit()), record(unit(hash='fixed', action='pass')),
                               {'candidate_changed': ['p/p.go'], 'main_changed': ['p/p.go']})
        self.assertEqual(result['candidate_reds'], 1)

    def test_parent_outcomes_do_not_duplicate_leaves(self):
        result = sort.sort_reds(record(unit('TestAnswer/sub'),
                                      test_outcomes=[unit(status='failed')],
                                      failed_tests=['example.com/p TestAnswer']), record())
        self.assertEqual(result['candidate_reds'], 1)

    def test_unledgered_failure_and_void_get_labels(self):
        result = sort.sort_reds(record(failure={'step': 'setup', 'detail': 'clone failed'}), record())
        self.assertEqual(result['infra'][0]['unit'], 'setup')
        result = sort.sort_reds(record(void=True), record())
        self.assertEqual(result['infra'][0]['unit'], 'gate void')

    def test_discovery_newest_finished_matching_tools(self):
        sha = 'a' * 40
        refs = ['refs/heads/gate-logs/' + sha[:12] + '/' + stamp + '/full-main'
                for stamp in ('20261009T010000Z', '20261009T020000Z', '20261009T030000Z')]
        reader = sort.MainRecords('/unused')
        current = [None]
        import json
        def git(*args):
            if args[:3] == ('ls-remote', 'origin', 'refs/heads/main'):
                return sha + '\trefs/heads/main'
            if args[0] == 'ls-remote':
                return '\n'.join('b' * 40 + '\t' + ref for ref in refs)
            if args[0] == 'fetch':
                current[0] = args[-1]
                return ''
            index = refs.index(current[0])
            return json.dumps(record(unit(), tools='other' if index == 2 else 'tools',
                                     finished=index != 1))
        with mock.patch.object(reader, 'git', side_effect=git):
            found, info = reader.newest('tools')
        self.assertEqual(info['main_ref'], refs[0].removeprefix('refs/heads/'))
        self.assertEqual(found['tools_fingerprint'], 'tools')

    def test_lookup_failure_is_explicit_and_closed(self):
        with mock.patch.object(sort.MainRecords, 'newest', side_effect=TimeoutError('limit')):
            result = sort.sort_record(record(unit()), '/unused')
        self.assertTrue(result['no_main_record'])
        self.assertEqual(result['candidate_reds'], 1)
        self.assertEqual(result['lookup_error'], 'limit')
