#!/usr/bin/env python3
"""Exercise the full lane entry point with isolated pipeline command doubles."""
import importlib.util
from pathlib import Path
import tempfile
import sys
import os
import subprocess
import unittest
from unittest.mock import patch
import json
import io
from contextlib import redirect_stdout

LANE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('lane_runner', LANE / 'run.py')
runner = importlib.util.module_from_spec(spec)
spec.loader.exec_module(runner)


class LaneTableTests(unittest.TestCase):
    def test_table_is_preserved_beside_report(self):
        with tempfile.TemporaryDirectory(prefix='lane-table-') as scratch:
            output = Path(scratch) / 'results'
            calls = []

            def command(arguments, **kwargs):
                calls.append(arguments)
                if arguments[1].endswith('apply.sh'):
                    tree = Path(arguments[2])
                    tree.mkdir()
                    (tree / 'patch-set.md').write_text('Table for this apply\n')
                else:
                    self.assertEqual((output / 'patch-set.md').read_text(), 'Table for this apply\n')
                return type('Result', (), {'returncode': 0})()

            def probe(arguments, **kwargs):
                return ('abc\n' if arguments[0] == 'git' else
                        json.dumps(dict(os='linux', arch='x64', node='v24.19.0')))

            with patch.object(sys, 'argv', ['run.py', str(output)]), \
                 patch.object(runner.subprocess, 'check_output', side_effect=probe), \
                 patch.object(runner.subprocess, 'run', side_effect=command), \
                 patch.object(runner, 'check_results', return_value=dict(status='pass', errors=[])):
                self.assertEqual(runner.main(), 0)
            self.assertEqual((output / 'patch-set.md').read_text(), 'Table for this apply\n')
            self.assertTrue((output / 'report.json').exists())
            self.assertEqual(len(calls), 2)
            self.assertNotIn('--write-table', calls[0])

    def test_no_arguments_creates_fresh_results(self):
        with tempfile.TemporaryDirectory(prefix='lane-default-test-') as scratch:
            calls = []
            stdout = io.StringIO()

            def command(arguments, **kwargs):
                calls.append(arguments)
                if arguments[1].endswith('apply.sh'):
                    tree = Path(arguments[2])
                    tree.mkdir()
                    (tree / 'patch-set.md').write_text('Default table\n')
                return type('Result', (), {'returncode': 0})()

            def probe(arguments, **kwargs):
                return ('abc\n' if arguments[0] == 'git' else
                        json.dumps(dict(os='linux', arch='x64', node='v24.19.0')))

            with patch.object(sys, 'argv', ['run.py']), \
                 patch.object(runner.tempfile, 'tempdir', scratch), \
                 patch.object(runner.subprocess, 'check_output', side_effect=probe), \
                 patch.object(runner.subprocess, 'run', side_effect=command), \
                 patch.object(runner, 'check_results', return_value=dict(status='pass', errors=[])), \
                 redirect_stdout(stdout):
                self.assertEqual(runner.main(), 0)
            lines = stdout.getvalue().splitlines()
            results = Path(lines[0])
            self.assertEqual(results.parent.resolve(), Path(scratch).resolve())
            self.assertTrue(results.is_dir())
            self.assertEqual(lines[-1], 'PASS stage3 landing lane')
            self.assertEqual((results / 'patch-set.md').read_text(), 'Default table\n')
            self.assertEqual(json.loads((results / 'report.json').read_text())['status'], 'pass')
            self.assertEqual(len(calls), 2)


    def test_shell_entry_point_with_no_arguments(self):
        with tempfile.TemporaryDirectory(prefix='lane-shell-test-') as scratch:
            node = Path(scratch) / 'node'
            node.write_text('#!/bin/sh\nprintf \'%s\\n\' \'{"os":"unknown","arch":"x64","node":"v24.19.0"}\'\n')
            node.chmod(0o755)
            env = dict(os.environ, TMPDIR=scratch,
                       PATH=scratch + os.pathsep + os.environ['PATH'],
                       PYTHONDONTWRITEBYTECODE='1')
            result = subprocess.run(['bash', str(LANE / 'run.sh')], env=env,
                                    stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
            self.assertEqual(result.returncode, 1)
            lines = result.stdout.splitlines()
            results = Path(lines[0])
            self.assertEqual(results.parent.resolve(), Path(scratch).resolve())
            self.assertTrue(results.is_dir())
            self.assertTrue(lines[-1].startswith('FAIL stage3 landing lane: unknown platform:'))
            self.assertEqual(json.loads((results / 'report.json').read_text())['status'], 'fail')


if __name__ == '__main__':
    unittest.main(verbosity=2)
