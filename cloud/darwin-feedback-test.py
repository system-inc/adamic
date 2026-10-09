#!/usr/bin/env python3
"""The Darwin feedback's diff of two go test -json runs."""
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest

spec = importlib.util.spec_from_file_location('feedback', Path(__file__).with_name('darwin-feedback.py'))
feedback = importlib.util.module_from_spec(spec)
spec.loader.exec_module(feedback)
P = 'github.com/system-inc/adamic/stage3/fixtures'


def run(*events):
    return '\n'.join(json.dumps(dict(event, Package=event.get('Package', P))) for event in events) + '\n'


class FeedbackTests(unittest.TestCase):
    def diff(self, base, candidate):
        with tempfile.TemporaryDirectory() as tmp:
            Path(tmp, 'b').write_text(base)
            Path(tmp, 'c').write_text(candidate)
            feedback.main.__globals__['sys'].argv = ['x', str(Path(tmp, 'b')), str(Path(tmp, 'c')), tmp]
            feedback.main()
            self.failing = Path(tmp, 'failing.txt').read_text().splitlines()
            return Path(tmp, 'summary.txt').read_text().strip(), Path(tmp, 'moved.txt').read_text()

    def test_moved_outcomes_and_failure_output(self):
        base = run({'Action': 'pass', 'Test': 'T/a'}, {'Action': 'pass', 'Test': 'T/b'},
                   {'Action': 'fail', 'Test': 'T/c'}, {'Action': 'fail', 'Test': 'T'}, {'Action': 'fail'})
        candidate = run({'Action': 'output', 'Test': 'T/a', 'Output': '=== RUN   T/a\n'},
                        {'Action': 'output', 'Test': 'T/a', 'Output': 'stdout differs: got 1\n'},
                        {'Action': 'fail', 'Test': 'T/a'}, {'Action': 'pass', 'Test': 'T/b'},
                        {'Action': 'pass', 'Test': 'T/c'}, {'Action': 'skip', 'Test': 'T/d'},
                        {'Action': 'fail', 'Test': 'T'}, {'Action': 'fail'})
        text, moved = self.diff(base, candidate)
        self.assertEqual(text, '3 outcomes moved (1 absent->skip, 1 fail->pass, 1 pass->fail)')
        self.assertIn('pass -> fail  %s T/a' % P, moved)
        self.assertIn('    stdout differs: got 1', moved)
        self.assertNotIn('=== RUN', moved)
        self.assertNotIn(' T\n', moved)
        self.assertEqual(self.failing, ['T/a'])

    def test_build_failure_reads_as_its_tests_gone(self):
        base = run({'Action': 'pass', 'Test': 'T/a'}, {'Action': 'pass', 'Test': 'T'}, {'Action': 'pass'})
        candidate = run({'Action': 'output', 'Output': 'node_fs_file.c:12: error: no member st_atimespec\n'},
                        {'Action': 'fail'})
        text, moved = self.diff(base, candidate)
        self.assertEqual(text, '3 outcomes moved (1 absent->fail, 2 pass->absent)')
        self.assertIn('st_atimespec', moved)
        self.assertEqual(self.failing, [P])

    def test_fixture_failing_before_its_children_is_named(self):
        base = run({'Action': 'pass', 'Test': 'T/f.a/node'}, {'Action': 'pass', 'Test': 'T/f.a/stage0'},
                   {'Action': 'pass', 'Test': 'T/f.a'}, {'Action': 'pass', 'Test': 'T'})
        candidate = run({'Action': 'pass', 'Test': 'T/f.a/node'}, {'Action': 'fail', 'Test': 'T/f.a'},
                        {'Action': 'fail', 'Test': 'T'})
        self.assertEqual(self.diff(base, candidate)[0], '3 outcomes moved (1 pass->absent, 2 pass->fail)')
        self.assertEqual(self.failing, ['T/f.a'])

    def test_vanished_and_what_a_landing_declares(self):
        base = run({'Action': 'pass', 'Test': 'T/a.a/native'}, {'Action': 'pass', 'Test': 'T/a.a'},
                   {'Action': 'pass', 'Test': 'T/b.a'}, {'Action': 'pass', 'Test': 'T/c.a'}, {'Action': 'fail', 'Test': 'T/d.a'})
        candidate = run({'Action': 'pass', 'Test': 'T/c.a'})
        with tempfile.TemporaryDirectory() as tmp:
            Path(tmp, 'b').write_text(base)
            Path(tmp, 'c').write_text(candidate)
            feedback.main.__globals__['sys'].argv = ['x', str(Path(tmp, 'b')), str(Path(tmp, 'c')), tmp]
            feedback.main()
            vanished = Path(tmp, 'vanished.txt').read_text()
        # A failing test going absent isn't a vanished pass.
        self.assertEqual(vanished, '%s T/a.a\n%s T/a.a/native\n%s T/b.a\n' % (P, P, P))
        self.assertEqual(feedback.undeclared(vanished, 'T/a.a -> T/moved/a.a\n# T/b.a -> removed\n'), ['%s T/b.a' % P])
        self.assertEqual(feedback.undeclared(vanished, 'T/a.a -> x\nT/b.a -> removed\n'), [])
        self.assertEqual(feedback.undeclared(vanished, 'T/a -> x\n'), vanished.splitlines())

    def test_nothing_moved(self):
        same = run({'Action': 'pass', 'Test': 'T/a'}, {'Action': 'pass'})
        self.assertEqual(self.diff(same, same)[0], '0 outcomes moved')


if __name__ == '__main__':
    unittest.main()
