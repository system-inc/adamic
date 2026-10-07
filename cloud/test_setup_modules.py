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
                return helper.subprocess.CompletedProcess(arguments, 0, stdout=b'', stderr=b'')

            with patch.object(helper.subprocess, 'check_output', side_effect=[b'{}', b'go1']), patch.object(
                helper.subprocess, 'run', side_effect=download
            ), patch.dict(os.environ, {'ADAMIC_GATE_UNCACHED': '1'}):
                with self.assertRaisesRegex(RuntimeError, 'definitions changed'):
                    helper.prepare(root, root / 'tools', all_modules=True)
            self.assertFalse(list((root / 'tools').glob('modules-*')))

    def test_preserved_mtime_edit_and_removal(self):
        with tempfile.TemporaryDirectory() as temporary:
            file=Path(temporary)/'module.mod';file.write_text('before');state=file.stat()
            before=helper.cache_state([str(file)])
            file.write_text('after!');os.utime(file,ns=(state.st_atime_ns,state.st_mtime_ns))
            self.assertNotEqual(before,helper.cache_state([str(file)]))
            file.unlink()
            with self.assertRaises(FileNotFoundError):helper.cache_state([str(file)])




class ClosureChecks(unittest.TestCase):
    def fixture(self, root):
        import zipfile
        directory=root/'module';directory.mkdir()
        contents={'go.mod':b'module example.org/module\n\ngo 1.27\n','main.go':b'package module\n'}
        prefix='example.org/module@v1.0.0/'
        archive=root/'module.zip'
        with zipfile.ZipFile(archive,'w') as zipped:
            for name,data in contents.items():
                zipped.writestr(prefix+name,data);(directory/name).write_bytes(data)
        manifest=root/'module.mod';manifest.write_bytes(contents['go.mod'])
        return dict(Path='example.org/module',Version='v1.0.0',Zip=str(archive),Dir=str(directory),
                    GoMod=str(manifest),Sum=helper.content_sum([(prefix+n,d) for n,d in contents.items()]),
                    GoModSum=helper.content_sum([('go.mod',contents['go.mod'])]))

    def test_archive_integrity(self):
        import zipfile
        with tempfile.TemporaryDirectory() as temporary:
            record=self.fixture(Path(temporary));helper.verify_selected(record)
            with zipfile.ZipFile(record['Zip'],'w') as archive:archive.writestr('example.org/module@v1.0.0/main.go',b'corruption')
            with self.assertRaisesRegex(ValueError,'archive/extracted checksum mismatch'):helper.verify_selected(record)

    def test_extracted_integrity(self):
        with tempfile.TemporaryDirectory() as temporary:
            record=self.fixture(Path(temporary));(Path(record['Dir'])/'main.go').write_bytes(b'corruption')
            with self.assertRaisesRegex(ValueError,'archive/extracted checksum mismatch'):helper.verify_selected(record)

    def test_go_mod_integrity(self):
        with tempfile.TemporaryDirectory() as temporary:
            record=self.fixture(Path(temporary));Path(record['GoMod']).write_bytes(b'corruption')
            with self.assertRaisesRegex(ValueError,'go.mod checksum mismatch'):helper.verify_selected(record)

    def test_closure_definition_changed_during_download(self):
        with tempfile.TemporaryDirectory() as temporary:
            root=Path(temporary);manifest=root/'go.mod';manifest.write_text('module example\n')
            def version(arguments,directory):
                manifest.write_text(manifest.read_text()+'\n');return 'go1'
            with patch.object(helper,'gate_closure',return_value=({}, {manifest}, [])), patch.object(helper,'go',side_effect=version), patch.dict(os.environ,ADAMIC_GATE_UNCACHED='1'):
                with self.assertRaisesRegex(RuntimeError,'definitions changed'):helper.prepare(root,root/'tools')

    def test_proxy_fallback_preserves_checksum_policy(self):
        failed=helper.subprocess.CompletedProcess([],1,stdout=b'',stderr=b'403 Forbidden')
        passed=helper.subprocess.CompletedProcess([],0,stdout=b'answer',stderr=b'')
        with patch.object(helper.subprocess,'run',side_effect=[failed,failed,passed]) as run, patch.dict(os.environ,GOPROXY='https://proxy.golang.org',GOSUMDB='sum.golang.org'):
            self.assertEqual(helper.go(['version'],Path('.')),'answer')
            self.assertEqual(run.call_args.kwargs['env']['GOPROXY'],'direct')
            self.assertEqual(run.call_args.kwargs['env']['GOSUMDB'],'sum.golang.org')
        failed.stderr=b'checksum mismatch'
        with patch.object(helper.subprocess,'run',return_value=failed) as run:
            with self.assertRaises(RuntimeError):helper.go(['version'],Path('.'))
            self.assertEqual(run.call_count,1)


@unittest.skipUnless(os.environ.get('ADAMIC_SETUP_INTEGRATION')=='1','opt-in actual Go lists')
class ClosureIntegration(unittest.TestCase):
    def test_unreachable_require_cannot_fail_mandatory_setup(self):
        with tempfile.TemporaryDirectory(prefix='closure-unreachable-', dir='/tmp/adamic-gate') as temporary:
            root=Path(temporary);repository=root/'repository';repository.mkdir()
            (repository/'go.mod').write_text('module closure-proof\n\ngo 1.27\n\nrequire example.invalid/unreachable v0.0.0\n')
            (repository/'main.go').write_text('package main\nfunc main() {}\n')
            environment=dict(GOWORK='off',GOPROXY='off',GOMODCACHE=str(root/'cache'),ADAMIC_GATE_UNCACHED='0')
            with patch.dict(os.environ,environment):
                self.assertIn('(0 downloaded modules)',helper.prepare(repository,root/'tools'))
                self.assertIn('skipped',helper.prepare(repository,root/'tools'))
                with patch.dict(os.environ,ADAMIC_GATE_UNCACHED='1'):
                    self.assertIn('(0 downloaded modules)',helper.prepare(repository,root/'tools'))
                self.assertFalse(list((root/'cache').rglob('*.zip')))
                with self.assertRaises(RuntimeError):helper.prepare(repository,root/'tools',all_modules=True)

if __name__=='__main__':unittest.main()
