#!/usr/bin/env python3
"""The automatic A/B's trigger: which reds write a job for Loom, what the job names, and the verdict coming back."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

script = Path(__file__).with_name('ab-trigger.py')
candidate, main = 'c' * 40, 'a' * 40
# Oct 8 23:51Z, 6b2c73f9's whole gate on the Threadripper: a stall guard's kill at load 81.7, read by hand as a 37x
# regression before a same-instance A/B showed 52 s on main and on the candidate alike.
stall = '''github.com/system-inc/adamic/stage1/cohere/tsprinter TestMutants/yield_loses_delegation
=== RUN   TestMutants/yield_loses_delegation
    shards_test.go:203: running /tmp/adamic-gate/x/002/port: stalled: no output for 2m0s after 2m27.1s, load 81.70 (/tmp/adamic-gate/x/002/port)
    mutants_test.go:116: shard 2: missing or extra case: 0 answers, want 21293
--- FAIL: TestMutants/yield_loses_delegation (278.01s)
'''
timeout = '''github.com/system-inc/adamic/internal/oracle
panic: test timed out after 3h0m0s
	running tests:
		TestLoopCountersAgreeWithNode (3h0m0s)
'''
assertion = '''github.com/system-inc/adamic/internal/fuzz TestOwnershipShapes
    shapes_test.go:40: got 3 shapes, want 4
--- FAIL: TestOwnershipShapes (0.40s)
'''


class AbTriggerTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        (self.root / 'bin').mkdir()
        (self.root / 'bin/ahra').write_text('#!/bin/bash\nprintf "%s|%s\\n" "$3" "$4" >> "$SENDS"\n')
        (self.root / 'bin/ahra').chmod(0o755)
        self.env = dict(os.environ, PATH=str(self.root / 'bin') + ':' + os.environ['PATH'], SENDS=str(self.root / 'sends'),
                        ADAMIC_AB_JOBS=str(self.root / 'jobs'), ADAMIC_FAST_GATE_AHRA_DIR=str(self.root))

    def write(self, text, notify='system_adamic_integration'):
        (self.root / 'first-failure.txt').write_text(text)
        self.output = subprocess.run(['python3', str(script), 'write', '--candidate', candidate, '--main', main, '--red', 'gate-logs/cccccccccccc/x/full-main',
                                      '--first-failure', str(self.root / 'first-failure.txt'), '--notify', notify], env=self.env, check=True, capture_output=True, text=True).stdout
        return sorted((self.root / 'jobs').glob('*.json')) if (self.root / 'jobs').exists() else []

    def test_a_stall_writes_one_job_naming_the_leaf_and_a_second_red_of_it_writes_none(self):
        written = self.write(stall)
        self.assertEqual(len(written), 1)
        job = json.loads(written[0].read_text())
        self.assertEqual((job['candidate'], job['main'], job['kind']), (candidate, main, 'stall'))
        self.assertEqual(job['package'], 'github.com/system-inc/adamic/stage1/cohere/tsprinter')
        # Leaf to leaf, never the parent: the -run pattern is the subtest itself.
        self.assertEqual(job['test'], '^TestMutants$/^yield_loses_delegation$')
        self.assertEqual(job['notify'], ['system_adamic_integration'])
        self.assertTrue(written[0].name.startswith(candidate[:12] + '-'))
        self.assertIn('wrote', self.output)
        self.assertEqual(len(self.write(stall)), 1)
        self.assertEqual(self.output, '', 'the same red and test write one job, ever')

    def test_a_timeout_names_the_test_the_panic_lists_as_running(self):
        job = json.loads(self.write(timeout)[0].read_text())
        self.assertEqual((job['kind'], job['package'], job['leaf']), ('timeout', 'github.com/system-inc/adamic/internal/oracle', 'TestLoopCountersAgreeWithNode'))

    def test_an_assertion_red_writes_no_job(self):
        self.assertEqual(self.write(assertion), [])

    def test_a_verdict_goes_once_to_whoever_heard_the_red(self):
        job = self.write(stall, notify='system_adamic_compiler,system_adamic_integration')[0]
        subprocess.run(['python3', str(script), 'collect'], env=self.env, check=True, capture_output=True)
        self.assertFalse((self.root / 'sends').exists(), 'no verdict yet, nothing sent')
        job.with_suffix('.verdict').write_text('regression 37.2x main 1.4 s candidate 52.1 s, cpu 2.9 s against 125 s\n')
        subprocess.run(['python3', str(script), 'collect'], env=self.env, check=True, capture_output=True)
        subprocess.run(['python3', str(script), 'collect'], env=self.env, check=True, capture_output=True)
        sends = (self.root / 'sends').read_text().splitlines()
        self.assertEqual([line.split('|')[0] for line in sends], ['system_adamic_compiler', 'system_adamic_integration'])
        self.assertIn('regression 37.2x main 1.4 s candidate 52.1 s', sends[0])
        self.assertIn('a regression in the candidate, not load', sends[0])
        self.assertIn('stage1/cohere/tsprinter TestMutants/yield_loses_delegation', sends[0])


if __name__ == '__main__':
    unittest.main()
