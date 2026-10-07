#!/usr/bin/env python3
"""Exercise the full lane entry point with isolated pipeline command doubles."""
import importlib.util
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import json

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

            with patch.object(runner.sys if hasattr(runner, 'sys') else __import__('sys'),
                              'argv', ['run.py', str(output)]), \
                 patch.object(runner.subprocess, 'check_output', side_effect=probe), \
                 patch.object(runner.subprocess, 'run', side_effect=command), \
                 patch.object(runner, 'check_results', return_value=dict(status='pass', errors=[])):
                self.assertEqual(runner.main(), 0)
            self.assertEqual((output / 'patch-set.md').read_text(), 'Table for this apply\n')
            self.assertTrue((output / 'report.json').exists())
            self.assertEqual(len(calls), 2)
            self.assertNotIn('--write-table', calls[0])


if __name__ == '__main__':
    unittest.main(verbosity=2)
