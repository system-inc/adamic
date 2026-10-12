#!/usr/bin/env python3
"""stage3 runs the npm ADAMIC_NPM names, by the tools' node, and refuses to run without one."""
import os
from pathlib import Path
import sys
import tempfile
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parent))
from npm_cli import npm


class NpmTests(unittest.TestCase):
    def test_the_named_package_runs_under_node(self):
        with tempfile.TemporaryDirectory() as package:
            cli = Path(package) / 'package/bin/npm-cli.js'
            cli.parent.mkdir(parents=True)
            cli.write_text('')
            with patch.dict(os.environ, {'ADAMIC_NPM': package}):
                self.assertEqual(npm('run', 'build'), ['node', str(cli), 'run', 'build'])

    def test_no_package_is_refused(self):
        with patch.dict(os.environ, {}, clear=True):
            with self.assertRaisesRegex(RuntimeError, 'ADAMIC_NPM is unset'):
                npm('test')

    def test_a_package_without_npm_is_refused(self):
        with tempfile.TemporaryDirectory() as package, patch.dict(os.environ, {'ADAMIC_NPM': package}):
            with self.assertRaisesRegex(RuntimeError, 'names no npm package'):
                npm('test')


if __name__ == '__main__':
    unittest.main(verbosity=2)
