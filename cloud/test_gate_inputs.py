#!/usr/bin/env python3
"""Hold corpus/archive caches to their inputs and bytes; exercise real local rebuilds."""
import contextlib
import io
import shlex
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

sys.dont_write_bytecode=True
SOURCE=Path(__file__).resolve().parent
HELPER=Path(os.environ.get('ADAMIC_GATE_INPUTS_MODULE', SOURCE/'setup-gate-inputs.py'))
spec=importlib.util.spec_from_file_location('gate',HELPER);gate=importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)


class Inputs(unittest.TestCase):
    def test_shared_prettier_exports_and_typescript_pin(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            with patch.object(sys, 'argv', ['setup', 'env', temporary, temporary, 'node']), contextlib.redirect_stdout(io.StringIO()) as output:
                gate.main()
            exports = dict(shlex.split(line)[1].split('=', 1) for line in output.getvalue().splitlines())
            self.assertIn('ADAMIC_ESTREE_LIBRARY', exports)
            self.assertEqual(exports['ADAMIC_ESTREE_LIBRARY'], exports['ADAMIC_CSS_PRINTER_LIBRARY'])
            self.assertIn('ADAMIC_YAML_LIBRARY', exports)
            self.assertEqual(exports['ADAMIC_YAML_LIBRARY'], exports['ADAMIC_CSS_PRINTER_LIBRARY'])
            self.assertIn('ADAMIC_TS_PRETTIER', exports)
            self.assertEqual(exports['ADAMIC_TS_PRETTIER'], str(root / 'css-printer'))
            self.assertEqual(exports['ADAMIC_TS_PRETTIER'], exports['ADAMIC_CSS_PRINTER_LIBRARY'])
            self.assertEqual(exports['ADAMIC_TYPESCRIPT_SOURCE'], str(root / 'typescript'))
            self.assertEqual(gate.TS_COMMIT, '050880ce59e30b356b686bd3144efe24f875ebc8')

    def test_shared_formatter_pins_and_integrity(self):
        directory = SOURCE / 'gate-inputs/css-printer'
        manifest = json.loads((directory / 'package.json').read_text())
        lock = json.loads((directory / 'package-lock.json').read_text())
        for name, version in [('yaml', '2.9.0'), ('prettier', '3.9.6'), ('yaml-unist-parser', '3.2.0'), ('@typescript-eslint/typescript-estree', '8.65.0'), ('typescript', '6.0.3')]:
            self.assertEqual(manifest['dependencies'][name], version)
            package = lock['packages']['node_modules/' + name]
            self.assertEqual(package['version'], version)
            self.assertTrue(package['integrity'].startswith('sha512-'))

    def test_shared_prettier_installed_once(self):
        with tempfile.TemporaryDirectory() as temporary:
            with patch.object(sys, 'argv', ['setup', 'npm', temporary, temporary, 'node']), patch.object(
                gate.npm, 'prepare', return_value='verified'
            ) as install, contextlib.redirect_stdout(io.StringIO()):
                gate.main()
            names = [call.args[0].name for call in install.call_args_list]
            self.assertEqual(names.count('css-printer'), 1)
            self.assertEqual(len(names), 7)

    def test_every_key_component(self):
        inputs=dict(kind='checker',head='head',packages=[['input','build-id']],environment={'CC':'gcc'},version='go1',
                    cc_version='gcc1',flags=['-trimpath'],validation_flags=['-buildmode=exe'],helper='source',commit='pin',url='https://source',size=100,content='bytes')
        before=gate.cache_key(**inputs)
        for name in inputs:
            with self.subTest(name=name):
                changed=dict(inputs,**{name:'changed'})
                self.assertNotEqual(before,gate.cache_key(**changed),name)

    def test_actual_content_modes_and_names(self):
        with tempfile.TemporaryDirectory() as temporary:
            root=Path(temporary);file=root/'input';file.write_bytes(b'original');file.chmod(0o644)
            old=gate.artifact_digest(root);file.write_bytes(b'changed')
            changed=gate.artifact_digest(root);self.assertNotEqual(old,changed)
            file.chmod(0o600);mode=gate.artifact_digest(root);self.assertNotEqual(changed,mode)
            file.rename(root/'renamed');renamed=gate.artifact_digest(root);self.assertNotEqual(mode,renamed)
            root.chmod(0o700 if root.stat().st_mode & 0o777 != 0o700 else 0o755)
            self.assertNotEqual(renamed,gate.artifact_digest(root))

    def test_corruption_and_uncached(self):
        with tempfile.TemporaryDirectory() as temporary:
            root=Path(temporary)
            with patch.dict(os.environ,ADAMIC_GATE_UNCACHED='0'):
                self.assertIn('generated',gate.gitignore(root,size=100))
                self.assertIn('skipped',gate.gitignore(root,size=100))
                self.assertIn('generated',gate.gitignore(root,size=101))
                original=gate.artifact_digest(root/'gitignore')
                file=root/'gitignore/.gitignore';file.write_bytes(b'corruption')
                self.assertIn('generated',gate.gitignore(root,size=101))
                self.assertEqual(original,gate.artifact_digest(root/'gitignore'))
            with patch.dict(os.environ,ADAMIC_GATE_UNCACHED='1'):
                self.assertIn('generated',gate.gitignore(root,size=101))
                self.assertEqual(original,gate.artifact_digest(root/'gitignore'))

    def test_git_checkout_integrity_and_repair(self):
        with tempfile.TemporaryDirectory() as temporary:
            root=Path(temporary);source=root/'source';source.mkdir()
            log=(root/'git.log').open('wb')
            def run(args):subprocess.run(args,cwd=source,stdout=log,stderr=subprocess.STDOUT,check=True,timeout=30)
            run(['git','init']);file=source/'compiler';file.write_text('first\n');run(['git','add','.'])
            commit=['git','-c','user.name=Gate proof','-c','user.email=gate@example.invalid','commit','-m','Proof']
            run(commit);pin=gate.command(['git','rev-parse','HEAD'],source)
            file.write_text('second\n');run(['git','add','.']);run(commit)
            second=gate.command(['git','rev-parse','HEAD'],source);inputs=root/'inputs';inputs.mkdir()
            with patch.object(gate,'TS_URL',str(source)),patch.object(gate,'TS_COMMIT',pin),patch.dict(os.environ,ADAMIC_GATE_UNCACHED='0'):
                self.assertIn('installed',gate.typescript(inputs));self.assertIn('skipped',gate.typescript(inputs))
                (inputs/'typescript/compiler').write_text('tampered')
                self.assertIn('installed',gate.typescript(inputs));self.assertEqual((inputs/'typescript/compiler').read_text(),'first\n')
                with patch.object(gate,'TS_COMMIT',second):
                    self.assertIn('installed',gate.typescript(inputs));self.assertEqual((inputs/'typescript/compiler').read_text(),'second\n')
                self.assertIn('installed',gate.typescript(inputs))
                original=gate.artifact_digest(inputs/'typescript')
                with patch.dict(os.environ,ADAMIC_GATE_UNCACHED='1'):
                    self.assertIn('installed',gate.typescript(inputs));self.assertEqual(original,gate.artifact_digest(inputs/'typescript'))
            log.close()


@unittest.skipUnless(os.environ.get('ADAMIC_SETUP_INTEGRATION')=='1','opt-in actual Go C archive')
class Archive(unittest.TestCase):
    def test_dirty_source_and_uncached_equality(self):
        with tempfile.TemporaryDirectory(prefix='gate-archive-',dir='/tmp/adamic-gate') as temporary:
            root=Path(temporary);repo=root/'repository';package=repo/'bridge/tsgo/archive';package.mkdir(parents=True)
            (repo/'go.mod').write_text('module '+os.environ.get('ADAMIC_GATE_ARCHIVE_PROOF_MODULE','gate-proof')+'\n\ngo 1.27\n')
            file=package/'main.go';file.write_text('package main\nimport "C"\n//export answer\nfunc answer() C.int { return 1 }\nfunc main() {}\n')
            with (root/'git.log').open('wb') as log:
                for args in [['git','init'],['git','add','.'],['git','-c','user.name=Gate proof','-c','user.email=gate@example.invalid','commit','-m','Proof']]:
                    subprocess.run(args,cwd=repo,stdout=log,stderr=subprocess.STDOUT,check=True,timeout=30)
            inputs=root/'inputs';inputs.mkdir()
            with patch.dict(os.environ,ADAMIC_GATE_UNCACHED='0'):
                self.assertIn('built',gate.archive(repo,inputs));self.assertIn('skipped',gate.archive(repo,inputs))
                self.assertFalse((repo/'.h').exists(), 'Go list must not emit .h into the checkout')
                before=(inputs/'checker/tsgo.a').read_bytes();file.write_text(file.read_text().replace('return 1','return 2'))
                self.assertIn('built',gate.archive(repo,inputs));self.assertNotEqual(before,(inputs/'checker/tsgo.a').read_bytes())
                self.assertIn('skipped',gate.archive(repo,inputs));original=gate.artifact_digest(inputs/'checker')
            with patch.dict(os.environ,ADAMIC_GATE_UNCACHED='1'):
                self.assertIn('built',gate.archive(repo,inputs));self.assertEqual(original,gate.artifact_digest(inputs/'checker'))


if __name__=='__main__':unittest.main()
