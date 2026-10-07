#!/usr/bin/env python3
"""Test installation inputs, integrity and actual npm cache invalidation."""
import base64
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

sys.dont_write_bytecode = True
SOURCE = Path(__file__).resolve().parent
HELPER = Path(os.environ.get('ADAMIC_MARKDOWN_SETUP_MODULE', SOURCE / 'setup-markdown-width.py'))
spec = importlib.util.spec_from_file_location('markdown_setup', HELPER)
helper = importlib.util.module_from_spec(spec)
spec.loader.exec_module(helper)


class InstallationKey(unittest.TestCase):
    def test_each_component_invalidates(self):
        inputs = dict(lock='lock', manifest='manifest', bootstrap='bootstrap', helper='helper', node='node')
        original = helper.installation_key(**inputs)
        for component in inputs:
            with self.subTest(component=component):
                changed = dict(inputs, **{component: 'changed'})
                self.assertNotEqual(original, helper.installation_key(**changed), component)

    def test_tree_bytes_paths_and_modes(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            file = root / 'index.js'
            file.write_text('original')
            file.chmod(0o644)
            answer = helper.tree_digest(root)
            file.write_text('changed!')
            changed = helper.tree_digest(root)
            self.assertNotEqual(answer, changed)
            file.chmod(0o600)
            permissions = helper.tree_digest(root)
            self.assertNotEqual(changed, permissions)
            file.rename(root / 'other.js')
            renamed = helper.tree_digest(root)
            self.assertNotEqual(permissions, renamed)
            root.chmod(0o755 if root.stat().st_mode & 0o777 != 0o755 else 0o700)
            self.assertNotEqual(renamed, helper.tree_digest(root))

    def test_unexpected_symlink_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'index.js').symlink_to('/outside/install')
            with self.assertRaisesRegex(ValueError, 'unexpected installed symlink'):
                helper.tree_digest(root)

    def test_integrity_rejects_corruption(self):
        data = b'archive bytes'
        integrity = 'sha512-' + base64.b64encode(hashlib.sha512(data).digest()).decode()
        helper.verified_archive(data, integrity)
        with self.assertRaisesRegex(ValueError, 'integrity mismatch'):
            helper.verified_archive(data + b'corrupt', integrity)


@unittest.skipUnless(os.environ.get('ADAMIC_SETUP_INTEGRATION') == '1', 'opt-in network')
class InstallationIntegration(unittest.TestCase):
    def test_real_invalidation_integrity_and_uncached_bytes(self):
        scratch = Path(tempfile.mkdtemp(prefix='markdown-proof-', dir='/tmp/adamic-gate'))
        print('integration logs:', scratch, flush=True)
        source = scratch / 'source'
        shutil.copytree(SOURCE / 'markdown-width', source)
        destination = scratch / 'dependencies'
        node = os.environ.get('ADAMIC_SETUP_NODE', '/workspace/adamic-tools/bin/node')
        number = 0

        def run(uncached=False, failure=False):
            nonlocal number
            number += 1
            log = scratch / f'{number:02d}.log'
            environment = dict(os.environ)
            environment.pop('ADAMIC_GATE_UNCACHED', None)
            if uncached:
                environment['ADAMIC_GATE_UNCACHED'] = '1'
            with log.open('wb') as output:
                result = subprocess.run([sys.executable, str(HELPER), str(source), str(destination), node],
                                        env=environment, stdout=output, stderr=subprocess.STDOUT, timeout=120)
            text = log.read_text()
            if failure:
                self.assertNotEqual(result.returncode, 0, text)
            else:
                self.assertEqual(result.returncode, 0, text)
            return text

        self.assertIn('installed (npm ci', run())
        self.assertIn('skipped (validated lock', run())
        for name in ['package-lock.json', 'package.json', 'npm-bootstrap.json']:
            file = source / name
            file.write_text(file.read_text() + '\n')
            self.assertIn('installed (npm ci', run())
            self.assertIn('skipped (validated lock', run())
        installed = destination / 'node_modules/emoji-regex/index.js'
        installed.write_text(installed.read_text() + '\n// corruption\n')
        self.assertIn('installed (npm ci', run())
        cached = helper.tree_digest(destination)
        self.assertIn('installed (npm ci', run(uncached=True))
        self.assertEqual(cached, helper.tree_digest(destination), 'uncached installed bytes must match')
        # Corrupt the expected package integrity. npm ci must reject it, and leave the
        # previously published installation intact rather than stamping a failed attempt.
        lock = source / 'package-lock.json'
        content = json.loads(lock.read_text())
        content['packages']['node_modules/emoji-regex']['integrity'] = 'sha512-' + base64.b64encode(bytes(64)).decode()
        lock.write_text(json.dumps(content))
        self.assertIn('EINTEGRITY', run(failure=True))
        self.assertEqual(cached, helper.tree_digest(destination))

    def test_install_uses_the_keyed_input_snapshot(self):
        scratch = Path(tempfile.mkdtemp(prefix='markdown-snapshot-', dir='/tmp/adamic-gate'))
        print('snapshot integration:', scratch, flush=True)
        source = scratch / 'source'
        shutil.copytree(SOURCE / 'markdown-width', source)
        lock = source / 'package-lock.json'
        original = lock.read_bytes()
        destination = scratch / 'dependencies'
        run = helper.subprocess.run

        def concurrent_edit(arguments, **kwargs):
            if arguments[0] == 'curl':
                lock.write_bytes(original + b'\n')
            return run(arguments, **kwargs)

        with (scratch / 'install.log').open('w') as output, patch.object(helper.subprocess, 'run', concurrent_edit):
            # Keep the real download/npm processes and mutate only the checkout input.
            # Redirect file descriptors because npm inherits the process's stdout/stderr.
            stdout, stderr = os.dup(1), os.dup(2)
            try:
                os.dup2(output.fileno(), 1)
                os.dup2(output.fileno(), 2)
                helper.prepare(source, destination, os.environ.get('ADAMIC_SETUP_NODE', '/workspace/adamic-tools/bin/node'))
            finally:
                os.dup2(stdout, 1)
                os.dup2(stderr, 2)
                os.close(stdout)
                os.close(stderr)
        self.assertEqual(original, (destination / 'package-lock.json').read_bytes())


if __name__ == '__main__':
    unittest.main()
