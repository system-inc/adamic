#!/usr/bin/env python3
"""Hold CSS fixture pins, corpus coverage and exact oracle resolution paths."""
import contextlib
import importlib.util
import io
import json
import os
from pathlib import Path
import shlex
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

sys.dont_write_bytecode = True
SOURCE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('gate', os.environ.get('ADAMIC_GATE_INPUTS_MODULE', SOURCE / 'setup-gate-inputs.py'))
gate = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gate)


class CSSInputs(unittest.TestCase):
    def test_exports_and_ordinary_unset(self):
        with patch.object(sys, 'argv', ['setup', 'env', '/tmp/repo', '/tmp/inputs', 'node']), contextlib.redirect_stdout(io.StringIO()) as output:
            gate.main()
        exports = dict(shlex.split(line)[1].split('=', 1) for line in output.getvalue().splitlines())
        expected = {'ADAMIC_CSS_FIXTURES': '/tmp/inputs/css-fixtures',
                    'ADAMIC_CSSNUMBERS_LIBRARY': '/tmp/inputs/css-printer',
                    'ADAMIC_CSSSTRINGS_LIBRARY': '/tmp/inputs/css-printer',
                    'ADAMIC_MARKDOWNINLINE_LIBRARY': '/tmp/inputs/css-printer/node_modules/prettier'}
        unset = next(line for line in (SOURCE / 'setup.sh').read_text().splitlines() if 'echo "unset ' in line)
        for name, path in expected.items():
            self.assertEqual(exports.get(name), path)
            self.assertIn(name, unset)
        self.assertEqual(gate.CSS_COMMIT, 'cb4b33fba24a8428d00e54be85fc886288a374ea')
        self.assertEqual(gate.CSS_COUNTS, {'.css': 157, '.scss': 90, '.less': 43})

    def test_every_fixture_key_component(self):
        before = gate.css_fixture_key()
        for name, value in [('CSS_COMMIT', 'other'), ('CSS_URL', 'other'),
                            ('CSS_SPARSE', ['other']), ('CSS_COUNTS', {'.css': 1})]:
            with self.subTest(name=name), patch.object(gate, name, value):
                self.assertNotEqual(before, gate.css_fixture_key(), name)
        with patch.object(gate, 'source_hash', return_value='changed helper'):
            self.assertNotEqual(before, gate.css_fixture_key())

    def test_counts_reject_degraded_corpus(self):
        with tempfile.TemporaryDirectory() as temporary:
            with self.assertRaisesRegex(ValueError, 'CSS fixture counts differ'):
                gate.css_counts(Path(temporary))

    def test_checkout_commit_and_sparse_checks(self):
        with patch.object(gate, 'command', return_value='wrong commit'):
            with self.assertRaisesRegex(ValueError, 'commit integrity mismatch'):
                gate.validate_css_checkout(Path('/unused'))
        with patch.object(gate, 'command', side_effect=[gate.CSS_COMMIT, 'wrong sparse']):
            with self.assertRaisesRegex(ValueError, 'sparse checkout differs'):
                gate.validate_css_checkout(Path('/unused'))

    def test_shared_prettier_version_is_verified(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary);package = root / 'css-printer/node_modules/prettier';package.mkdir(parents=True)
            (root / 'css-printer/package.json').write_text('{}')
            (package / 'package.json').write_text(json.dumps({'version': '3.9.6', 'main': 'index.js'}))
            (package / 'index.js').write_text("exports.version = require('./package.json').version;\n")
            (package / 'plugins').mkdir()
            for name in ['postcss', 'markdown']:
                (package / 'plugins' / (name + '.js')).write_text('module.exports = {};\n')
            self.assertIn('paths verified', gate.shared_prettier(root, 'node'))
            # Keep the direct export correct so only prefix/package.json validation catches this.
            (package / 'package.json').write_text(json.dumps({'version': '0.0.0', 'main': 'index.js'}))
            (package / 'index.js').write_text("exports.version = '3.9.6';\n")
            with self.assertRaises(subprocess.CalledProcessError):
                gate.shared_prettier(root, 'node')
            # Then preserve prefix metadata so only direct package validation catches this.
            (package / 'package.json').write_text(json.dumps({'version': '3.9.6', 'main': 'index.js'}))
            (package / 'index.js').write_text("exports.version = '0.0.0';\n")
            with self.assertRaises(subprocess.CalledProcessError):
                gate.shared_prettier(root, 'node')

    def test_local_sparse_cache_repair_and_uncached_equality(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary);source = root / 'source';source.mkdir();inputs = root / 'inputs';inputs.mkdir()
            log = (root / 'git.log').open('wb')
            def git(*args):
                subprocess.run(['git', *args], cwd=source, check=True, stdout=log, stderr=subprocess.STDOUT)
            git('init')
            for directory, extension in [('tests/format/css', '.css'), ('tests/format/scss', '.scss'), ('tests/format/less', '.less')]:
                path = source / directory / ('fixture' + extension);path.parent.mkdir(parents=True);path.write_text('first\n')
            outside = source / 'outside/ignored.css';outside.parent.mkdir();outside.write_text('must not enter corpus\n')
            git('add', '.');git('-c', 'user.name=Proof', '-c', 'user.email=proof@example.invalid', 'commit', '-m', 'First')
            pin = gate.command(['git', 'rev-parse', 'HEAD'], source)
            (source / 'tests/format/css/fixture.css').write_text('second\n')
            git('add', '.');git('-c', 'user.name=Proof', '-c', 'user.email=proof@example.invalid', 'commit', '-m', 'Second')
            second = gate.command(['git', 'rev-parse', 'HEAD'], source)
            sparse = ['tests/format/css', 'tests/format/scss', 'tests/format/less']
            with patch.object(gate, 'CSS_URL', str(source)), patch.object(gate, 'CSS_COMMIT', pin), patch.object(gate, 'CSS_SPARSE', sparse), patch.object(gate, 'CSS_COUNTS', {'.css': 1, '.scss': 1, '.less': 1}), patch.dict(os.environ, ADAMIC_GATE_UNCACHED='0'):
                self.assertIn('installed', gate.css_fixtures(inputs))
                self.assertIn('skipped', gate.css_fixtures(inputs))
                self.assertFalse((inputs / 'css-fixtures/outside').exists())
                file = inputs / 'css-fixtures/tests/format/css/fixture.css';file.write_text('corruption\n')
                self.assertIn('installed', gate.css_fixtures(inputs));self.assertEqual(file.read_text(), 'first\n')
                with patch.object(gate, 'CSS_COMMIT', second):
                    self.assertIn('installed', gate.css_fixtures(inputs));self.assertEqual(file.read_text(), 'second\n')
                self.assertIn('installed', gate.css_fixtures(inputs))
                original = gate.artifact_digest(inputs / 'css-fixtures')
                with patch.dict(os.environ, ADAMIC_GATE_UNCACHED='1'):
                    self.assertIn('installed', gate.css_fixtures(inputs))
                    self.assertEqual(original, gate.artifact_digest(inputs / 'css-fixtures'))
            log.close()


if __name__ == '__main__':
    unittest.main()
