#!/usr/bin/env python3
"""Check manifest and cached artifact invalidation, including mtime-preserving edits."""
import importlib.util
import os
from pathlib import Path
import sys
import tempfile
import types
import unittest
import json
import zipfile
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

            with patch.object(helper.subprocess, 'check_output', side_effect=[b'{}', b'go1', b'{"Module":{"Path":"example"}}', b'{"Module":{"Path":"example"}}']), patch.object(
                helper.subprocess, 'run', side_effect=download
            ), patch.dict(os.environ, {'ADAMIC_GATE_UNCACHED': '1'}):
                with self.assertRaisesRegex(RuntimeError, 'definitions changed'):
                    helper.prepare(root, root / 'tools')
            self.assertFalse(list((root / 'tools').glob('modules-*')))

    def test_each_manifest_downloads_all_without_workspace(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            (root / 'go.mod').write_text('module example\nrequire unused.example/input v1.0.0\n')
            seen = []
            def run(arguments, **kwargs):
                seen.append(arguments)
                self.assertEqual(kwargs['env']['GOWORK'], 'off')
                if arguments[2] == 'download':
                    self.assertEqual(arguments[-1], 'all')
                    self.assertIn('-json', arguments)
                    modfile = Path(next(arg.split('=', 1)[1] for arg in arguments if arg.startswith('-modfile=')))
                    self.assertIn('unused.example/input v1.0.0', modfile.read_text())
                return helper.subprocess.CompletedProcess(arguments, 0, stdout=b'')
            outputs = [b'{}', b'go1', b'{"Module":{"Path":"example"}}', b'{"Module":{"Path":"example"}}']
            with patch.object(helper.subprocess, 'check_output', side_effect=outputs), patch.object(helper.subprocess, 'run', side_effect=run), patch.dict(os.environ, ADAMIC_GATE_UNCACHED='1'):
                helper.prepare(root, root / 'tools')
            self.assertIn('download', [args[2] for args in seen])
            self.assertEqual((root / 'go.mod').read_text(), 'module example\nrequire unused.example/input v1.0.0\n')
            self.assertFalse((root / 'go.sum').exists())

    def test_preserved_mtime_edit_and_removal(self):
        with tempfile.TemporaryDirectory() as temporary:
            file=Path(temporary)/'module.mod';file.write_text('before');state=file.stat()
            before=helper.cache_state([str(file)])
            file.write_text('after!');os.utime(file,ns=(state.st_atime_ns,state.st_mtime_ns))
            self.assertNotEqual(before,helper.cache_state([str(file)]))
            file.unlink()
            with self.assertRaises(FileNotFoundError):helper.cache_state([str(file)])


@unittest.skipUnless(os.environ.get('ADAMIC_SETUP_INTEGRATION') == '1', 'opt-in actual Go module graph')
class FullGraph(unittest.TestCase):
    def test_unused_transitive_requirement_is_downloaded(self):
        with tempfile.TemporaryDirectory(prefix='module-graph-', dir='/tmp/adamic-gate') as temporary:
            root = Path(temporary)
            proxy = root / 'proxy'
            for name, requirement in [('unused.example/input', 'require unused.example/nested v1.0.0\n'),
                                      ('unused.example/nested', '')]:
                directory = proxy / name / '@v'
                directory.mkdir(parents=True)
                mod = 'module ' + name + '\n\ngo 1.27\n' + requirement
                (directory / 'v1.0.0.mod').write_text(mod)
                (directory / 'v1.0.0.info').write_text(json.dumps({'Version':'v1.0.0','Time':'2026-01-01T00:00:00Z'}))
                (directory / 'list').write_text('v1.0.0\n')
                with zipfile.ZipFile(directory / 'v1.0.0.zip', 'w') as archive:
                    archive.writestr(name + '@v1.0.0/go.mod', mod)
                    archive.writestr(name + '@v1.0.0/input.go', 'package input\n')
            repository = root / 'repository'
            repository.mkdir()
            original = 'module consumer.example/main\n\ngo 1.27\nrequire unused.example/input v1.0.0\n'
            (repository / 'go.mod').write_text(original)
            cache = root / 'cache'
            with patch.dict(os.environ, GOMODCACHE=str(cache), GOPROXY=proxy.as_uri(), GOSUMDB='off',
                            GONOPROXY='none', GOPRIVATE='', GOTOOLCHAIN='local', GOWORK='off',
                            GOFLAGS='', ADAMIC_GATE_UNCACHED='0'):
                self.assertIn('downloaded', helper.prepare(repository, root / 'tools'))
                for name in ['input', 'nested']:
                    self.assertTrue((cache / ('unused.example/' + name + '@v1.0.0/input.go')).is_file(), name)
                self.assertIn('skipped', helper.prepare(repository, root / 'tools'))
                self.assertEqual((repository / 'go.mod').read_text(), original)
                self.assertFalse((repository / 'go.sum').exists())
                with patch.dict(os.environ, ADAMIC_GATE_UNCACHED='1'):
                    self.assertIn('downloaded', helper.prepare(repository, root / 'tools'))


if __name__=='__main__':unittest.main()
