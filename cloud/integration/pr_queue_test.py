#!/usr/bin/env python3
"""pr_queue.py against real git: a bare origin standing in for GitHub and a clone the pr lane works in.

usage: python3 -B cloud/integration/pr_queue_test.py
"""
import importlib.util
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

specification = importlib.util.spec_from_file_location('pr_queue', Path(__file__).with_name('pr_queue.py'))
prQueue = importlib.util.module_from_spec(specification)
specification.loader.exec_module(prQueue)

identity = {'GIT_AUTHOR_NAME': 't', 'GIT_AUTHOR_EMAIL': 't@t', 'GIT_COMMITTER_NAME': 't', 'GIT_COMMITTER_EMAIL': 't@t'}
os.environ.update(identity)


def testOnly(path):
    return path.endswith('_test.go')


def git(where, *arguments):
    return subprocess.run(['git', '-C', where, *arguments], check=True, capture_output=True, text=True).stdout.strip()


class Merged(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.origin = os.path.join(self.tmp.name, 'origin.git')
        git(self.tmp.name, 'init', '-q', '--bare', self.origin)
        self.work = os.path.join(self.tmp.name, 'work')
        git(self.tmp.name, 'init', '-q', self.work)
        git(self.work, 'remote', 'add', 'origin', self.origin)
        self.base = self.commit({'main.go': 'package main\n'}, None)

    def commit(self, files, on):
        if on is not None:
            git(self.work, 'checkout', '-q', '--detach', on)
        for name, text in files.items():
            path = os.path.join(self.work, name)
            os.makedirs(os.path.dirname(path), exist_ok=True)
            with open(path, 'w') as handle:
                handle.write(text)
        git(self.work, 'add', '-A')
        git(self.work, 'commit', '-q', '-m', ' '.join(sorted(files)))
        return git(self.work, 'rev-parse', 'HEAD')

    def test_a_head_on_an_older_main_is_merged_with_mains_tip_as_its_first_parent(self):
        head = self.commit({'a_test.go': 'package main\n'}, self.base)
        main = self.commit({'main.go': 'package main\n\nfunc main() {}\n'}, self.base)
        sha, conflicts = prQueue.merged(self.work, main, head, 'Land pull request #7')
        self.assertEqual(conflicts, [])
        self.assertEqual(git(self.work, 'rev-list', '--parents', '-n', '1', sha).split()[1:], [main, head])
        # The tree is main's plus the head's change: diffed against main, it touches only the pull request's path.
        self.assertEqual(git(self.work, 'diff', '--name-only', main, sha), 'a_test.go')
        self.assertIn('func main', git(self.work, 'show', sha + ':main.go'))
        self.assertEqual(git(self.work, 'log', '-1', '--format=%s', sha), 'Land pull request #7')

    def test_a_head_that_already_descends_from_main_is_itself(self):
        head = self.commit({'a_test.go': 'package main\n'}, self.base)
        self.assertEqual(prQueue.merged(self.work, self.base, head, 'm'), (head, []))

    def test_a_head_main_already_holds_is_nothing_to_land(self):
        head = self.commit({'a_test.go': 'package main\n'}, self.base)
        main = self.commit({'b.go': 'package main\n'}, head)
        self.assertEqual(prQueue.merged(self.work, main, head, 'm'), (None, []))

    def test_a_conflict_writes_no_commit_and_names_its_paths(self):
        head = self.commit({'main.go': 'package main // head\n'}, self.base)
        main = self.commit({'main.go': 'package main // main\n'}, self.base)
        self.assertEqual(prQueue.merged(self.work, main, head, 'Land pull request #7'), (None, ['main.go']))
        self.assertNotIn('Land', git(self.work, 'log', '--all', '--format=%s'))
        self.assertEqual(git(self.work, 'rev-parse', 'HEAD'), main)

    def test_a_test_only_head_on_an_older_main_is_submitted_as_its_merge_with_only_its_own_paths(self):
        head = self.commit({'internal/x/a_test.go': 'package x\n'}, self.base)
        main = self.commit({'main.go': 'package main\n\nfunc main() {}\n'}, self.base)
        found = prQueue.candidate(self.work, main, 7, head, testOnly)
        self.assertEqual(found['action'], 'submit')
        self.assertEqual(found['body'], {'sha': found['sha'], 'base': main, 'paths': ['internal/x/a_test.go']})
        self.assertEqual(git(self.work, 'rev-list', '--parents', '-n', '1', found['sha']).split()[1:], [main, head])

    def test_a_head_that_changes_anything_but_tests_is_never_submitted(self):
        head = self.commit({'internal/x/a_test.go': 'package x\n', 'internal/x/a.go': 'package x\n'}, self.base)
        self.assertEqual(prQueue.candidate(self.work, self.base, 7, head, testOnly), {'action': 'outside', 'paths': ['internal/x/a.go']})

    def test_a_conflict_or_a_head_main_holds_is_never_submitted(self):
        head = self.commit({'main.go': 'package main // head\n'}, self.base)
        main = self.commit({'main.go': 'package main // main\n'}, self.base)
        self.assertEqual(prQueue.candidate(self.work, main, 7, head, testOnly), {'action': 'conflict', 'paths': ['main.go']})
        held = self.commit({'b_test.go': 'package main\n'}, main)
        self.assertEqual(prQueue.candidate(self.work, held, 7, main, testOnly), {'action': 'held'})

    def test_check_mode_decides_and_touches_nothing_on_origin_and_on_mode_publishes_then_submits_once(self):
        head = self.commit({'internal/x/a_test.go': 'package x\n'}, self.base)
        main = self.commit({'main.go': 'package main\n\nfunc main() {}\n'}, self.base)
        posts = []

        def post(body, owner, secret):
            posts.append((body, owner))
            return 201, {'change': 'chg_' + 'a' * 26, 'state': 'queued'}

        checked = prQueue.throughQueue('check', self.work, main, 7, head, testOnly, 'owner', 'secret', post)
        self.assertEqual((checked['action'], posts), ('submit', []))
        self.assertEqual(git(self.origin, 'for-each-ref', '--format=%(refname)'), '')
        landed = prQueue.throughQueue('on', self.work, main, 7, head, testOnly, 'owner', 'secret', post)
        self.assertEqual((landed['status'], landed['answer']['change']), (201, 'chg_' + 'a' * 26))
        self.assertEqual(posts, [({'sha': landed['sha'], 'base': main, 'paths': ['internal/x/a_test.go']}, 'owner')])
        # The queue's git facts need the sha on GitHub before it's submitted.
        self.assertEqual(git(self.origin, 'rev-parse', 'refs/heads/cloud/pr-queue-7-' + landed['sha'][:8]), landed['sha'])
        # Anything but a submit posts nothing, in either mode.
        outside = self.commit({'internal/x/a.go': 'package x\n'}, self.base)
        self.assertEqual(prQueue.throughQueue('on', self.work, main, 8, outside, testOnly, 'owner', 'secret', post)['action'], 'outside')
        self.assertEqual(len(posts), 1)

    def test_a_submit_token_is_the_wires_shape_signed_with_the_secret(self):
        claims, signature = prQueue.token('secret\n', 'system_adamic_loom_queue').split('.')
        decoded = __import__('json').loads(__import__('base64').urlsafe_b64decode(claims + '=' * (-len(claims) % 4)))
        self.assertEqual(sorted(decoded), ['expires', 'run', 'scope'])
        self.assertEqual((decoded['run'], decoded['scope']), ('system_adamic_loom_queue', 'submit'))
        expected = __import__('hmac').new(b'secret', claims.encode(), 'sha256').digest()
        self.assertEqual(__import__('base64').urlsafe_b64encode(expected).rstrip(b'=').decode(), signature)

    def test_publish_puts_each_sha_on_its_own_branch_and_never_rewrites_one(self):
        head = self.commit({'a_test.go': 'package main\n'}, self.base)
        main = self.commit({'main.go': 'package main\n\nfunc main() {}\n'}, self.base)
        sha, _ = prQueue.merged(self.work, main, head, 'm')
        branch = prQueue.publish(self.work, sha, 7)
        self.assertEqual(branch, 'cloud/pr-queue-7-' + sha[:8])
        self.assertEqual(git(self.origin, 'rev-parse', 'refs/heads/' + branch), sha)
        # Main moved: the new merge goes on a new branch, and the old branch keeps its sha.
        newer = self.commit({'c.go': 'package main\n'}, main)
        again, _ = prQueue.merged(self.work, newer, head, 'm')
        self.assertNotEqual(prQueue.publish(self.work, again, 7), branch)
        self.assertEqual(git(self.origin, 'rev-parse', 'refs/heads/' + branch), sha)


if __name__ == '__main__':
    unittest.main()
