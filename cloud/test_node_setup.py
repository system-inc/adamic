#!/usr/bin/env python3
"""Hold the published Node pin, installer stamp and integrity refusal to real inputs."""
import contextlib
import hashlib
import importlib.util
import io
import json
import os
from pathlib import Path
import tarfile
import tempfile
import unittest
from unittest.mock import patch

SOURCE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('node_setup', os.environ.get('ADAMIC_NODE_SETUP_MODULE', SOURCE / 'setup-node.py'))
node = importlib.util.module_from_spec(spec)
spec.loader.exec_module(node)


class NodeSetup(unittest.TestCase):
    def test_published_pins_and_platforms(self):
        pin = json.loads((SOURCE / 'node-pin.json').read_text())
        self.assertEqual(pin['version'], 'v24.19.0')
        for system in ['Linux', 'Darwin']:
            for machine in ['x86_64', 'arm64', 'aarch64']:
                platform, arch = node.asset(system, machine)
                self.assertRegex(pin['sha256'][f'node-v24.19.0-{platform}-{arch}.tar.gz'], r'^[0-9a-f]{64}$')
        with self.assertRaisesRegex(ValueError, 'unsupported'):
            node.asset('Windows', 'x86_64')

    def test_every_key_component(self):
        inputs = dict(version='v24.19.0', checksum='sum', system='linux', architecture='x64', pin='pin', helper='helper')
        before = node.installation_key(**inputs)
        for component in inputs:
            with self.subTest(component=component):
                self.assertNotEqual(before, node.installation_key(**dict(inputs, **{component: 'changed'})), component)

    def fixture(self, root):
        version = 'v24.19.0'
        name = f'node-{version}-linux-x64.tar.gz'
        binary = b'#!/bin/sh\nprintf "v24.19.0\\n"\n'
        archive = root / name
        with tarfile.open(archive, 'w:gz') as release:
            member = tarfile.TarInfo(name[:-7] + '/bin/node')
            member.size = len(binary)
            member.mode = 0o755
            release.addfile(member, io.BytesIO(binary))
        checksum = hashlib.sha256(archive.read_bytes()).hexdigest()
        pin = root / 'pin.json'
        pin.write_text(json.dumps(dict(version=version, sha256={name: checksum})))
        def download(url, destination):
            if url.endswith('SHASUMS256.txt'):
                destination.write_text(checksum + '  ' + name + '\n')
            else:
                destination.write_bytes(archive.read_bytes())
        return pin, download

    def test_warm_corruption_modes_uncached_and_pin_bytes(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            pin, download = self.fixture(root)
            tools = root / 'tools'
            with patch.object(node, 'PIN', pin), patch.object(node, 'asset', return_value=('linux', 'x64')), patch.object(node, 'download', side_effect=download) as fetch, patch.dict(os.environ, ADAMIC_GATE_UNCACHED='0'):
                self.assertIn('installed (published SHA-256 verified)', node.prepare(tools))
                before = (tools / 'bin/node').read_bytes(), (tools / 'node.stamp').read_bytes()
                fetch.reset_mock()
                self.assertIn('skipped', node.prepare(tools))
                fetch.assert_not_called()
                (tools / 'bin/node').write_text('#!/bin/sh\nprintf "v24.21.0\\n"\n')
                self.assertIn('installed (published SHA-256 verified)', node.prepare(tools))
                (tools / 'bin/node').chmod(0o700)
                self.assertIn('installed (published SHA-256 verified)', node.prepare(tools))
                pin.write_text(pin.read_text() + '\n')
                self.assertIn('installed (published SHA-256 verified)', node.prepare(tools))
                pin.write_text(pin.read_text().rstrip() + '\n')
                node.prepare(tools)
                before = (tools / 'bin/node').read_bytes(), (tools / 'node.stamp').read_bytes()
                with patch.dict(os.environ, ADAMIC_GATE_UNCACHED='1'):
                    self.assertIn('installed (published SHA-256 verified)', node.prepare(tools))
                    self.assertEqual(before, ((tools / 'bin/node').read_bytes(), (tools / 'node.stamp').read_bytes()))

    def test_corrupted_archive_refused_before_install(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            pin, download = self.fixture(root)
            def corrupt(url, destination):
                download(url, destination)
                if not url.endswith('SHASUMS256.txt'):
                    destination.write_bytes(destination.read_bytes() + b'corruption')
            with patch.object(node, 'PIN', pin), patch.object(node, 'asset', return_value=('linux', 'x64')), patch.object(node, 'download', side_effect=corrupt):
                with self.assertRaisesRegex(ValueError, 'SHA-256 mismatch'):
                    node.prepare(root / 'tools')
                self.assertFalse((root / 'tools/bin/node').exists())

    def test_published_checksum_must_match_pin(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            pin, download = self.fixture(root)
            def changed(url, destination):
                download(url, destination)
                if url.endswith('SHASUMS256.txt'):
                    destination.write_text(destination.read_text().replace('  ', '0  '))
            with patch.object(node, 'PIN', pin), patch.object(node, 'asset', return_value=('linux', 'x64')), patch.object(node, 'download', side_effect=changed):
                with self.assertRaisesRegex(ValueError, 'published SHA-256 differs'):
                    node.prepare(root / 'tools')

    def test_portable_darwin_flow(self):
        spec = importlib.util.spec_from_file_location('darwin_setup', SOURCE / 'setup-darwin.py')
        darwin = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(darwin)
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            def command_output(command, **kwargs):
                if command[:2] == ['go', 'version']: return 'go version go1.27.1 darwin/arm64'
                if command[:2] == ['clang', '--version']: return 'Apple clang'
                if command[-1] == '--version': return 'v24.19.0'
                if 'setup-key.py' in ' '.join(command): return 'warm-key'
                if 'rev-parse' in command: return 'commit'
                raise AssertionError(command)
            with patch.dict(os.environ, ADAMIC_TOOLS=str(root), ADAMIC_GATE_UNCACHED='0'), patch.object(
                darwin.sys, 'argv', ['setup-darwin.py']), patch.object(darwin.node, 'prepare', return_value='node v24.19.0 ready'), patch.object(
                darwin.shutil, 'which', return_value='/go'), patch.object(darwin, 'output', side_effect=command_output), patch.object(
                darwin.subprocess, 'run'), contextlib.redirect_stdout(io.StringIO()) as captured:
                darwin.main()
            self.assertEqual(captured.getvalue().splitlines()[-1], 'setup: node v24.19.0')
            self.assertIn('node=v24.19.0', captured.getvalue())
            self.assertIn('unset ADAMIC_TYPESCRIPT_SOURCE', (root / 'env.sh').read_text())

    def test_root_location_refused(self):
        with self.assertRaisesRegex(ValueError, 'outside /root'):
            node.prepare('/root/node-proof')


if __name__ == '__main__':
    unittest.main()
