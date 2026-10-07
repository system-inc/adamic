#!/usr/bin/env python3
"""Check manifest and cached artifact invalidation, including mtime-preserving edits."""
import importlib.util
import os
from pathlib import Path
import sys
import tempfile
import types
import unittest
from unittest.mock import patch

sys.dont_write_bytecode = True
spec=importlib.util.spec_from_file_location('modules',os.environ.get('ADAMIC_MODULES_HELPER',str(Path(__file__).with_name('setup-modules.py'))))
helper=importlib.util.module_from_spec(spec);spec.loader.exec_module(helper)


class Modules(unittest.TestCase):
    def test_key_components(self):
        inputs=dict(manifests=[['go.sum','bytes']],version='go1',environment={'GOMODCACHE':'cache'},helper='source')
        before=helper.key(**inputs)
        for name in inputs:
            self.assertNotEqual(before,helper.key(**dict(inputs,**{name:'changed'})),name)

    def test_all_manifests(self):
        with tempfile.TemporaryDirectory() as temporary:
            root=Path(temporary);nested=root/'cohere/nested';nested.mkdir(parents=True)
            for file in ['go.mod','go.sum','go.work','go.work.sum']:
                (root/file).write_text('initial')
                (nested/file).write_text('initial')
            modules,before=helper.manifests(root)
            self.assertEqual(modules,[root,nested])
            for path in sorted(root.rglob('go.*')):
                path.write_text(path.read_text()+'changed')
                _,after=helper.manifests(root);self.assertNotEqual(before,after,str(path));before=after

    def test_every_cache_field(self):
        fields=dict(st_mode=0o100644,st_size=4,st_mtime_ns=1,st_ctime_ns=1,st_ino=1)
        with tempfile.TemporaryDirectory() as temporary:
            file=Path(temporary)/'module.mod';file.write_text('data')
            with patch.object(Path,'lstat',return_value=types.SimpleNamespace(**fields)):
                before=helper.cache_state([str(file)])
            for name in fields:
                changed=dict(fields,**{name:fields[name]+1})
                with patch.object(Path,'lstat',return_value=types.SimpleNamespace(**changed)):
                    self.assertNotEqual(before,helper.cache_state([str(file)]),name)
            renamed=file.with_name('other.mod');file.rename(renamed)
            with patch.object(Path,'lstat',return_value=types.SimpleNamespace(**fields)):
                self.assertNotEqual(before,helper.cache_state([str(renamed)]))

    def test_changed_definition_during_download(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            (root / 'go.mod').write_text('module example\n')
            workspace = root / 'go.work'
            workspace.write_text('go 1.27\n')

            def download(arguments, **kwargs):
                if arguments[2] == 'verify':
                    workspace.write_text(workspace.read_text() + '\n')
                return helper.subprocess.CompletedProcess(arguments, 0, stdout=b'')

            with patch.object(helper.subprocess, 'check_output', side_effect=[b'{}', b'go1']), patch.object(
                helper.subprocess, 'run', side_effect=download
            ), patch.dict(os.environ, {'ADAMIC_GATE_UNCACHED': '1'}):
                with self.assertRaisesRegex(RuntimeError, 'definitions changed'):
                    helper.prepare(root, root / 'tools')
            self.assertFalse(list((root / 'tools').glob('modules-*')))

    def test_preserved_mtime_edit_and_removal(self):
        with tempfile.TemporaryDirectory() as temporary:
            file=Path(temporary)/'module.mod';file.write_text('before');state=file.stat()
            before=helper.cache_state([str(file)])
            file.write_text('after!');os.utime(file,ns=(state.st_atime_ns,state.st_mtime_ns))
            self.assertNotEqual(before,helper.cache_state([str(file)]))
            file.unlink()
            with self.assertRaises(FileNotFoundError):helper.cache_state([str(file)])


if __name__=='__main__':unittest.main()
