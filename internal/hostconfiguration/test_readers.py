"""Exercise real reader entry points with scratch status, before expensive work."""
import contextlib
import ast
import types
import copy
import io
import json
import os
from pathlib import Path
import runpy
import shutil
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(Path(__file__).parent))
sys.dont_write_bytecode = True
from configuration import REGISTRY
SWITCH = REGISTRY['node-without-inspector']['environment']
CONFIG = dict(name='node-without-inspector', node='v24.19.0', configure='--without-inspector', ruling='#cn4yapb')
STATUS = ROOT / 'stage3/fixtures/host/status.json'
ORIGINAL = json.loads(STATUS.read_text())


class AtNode(Exception):
    pass


class ReaderTests(unittest.TestCase):
    def run_reader(self, reader, filename, mode, metadata=False, added_configuration=False):
        row = copy.deepcopy(next(row for row in ORIGINAL if row['file'] == filename))
        row['configuration'] = dict(CONFIG)
        if mode == 'unknown':
            row['configuration']['name'] = 'unknown-build'
        data = json.dumps([row])
        before = copy.deepcopy(row)
        if added_configuration:
            before.pop("configuration")
        before_data = json.dumps([before])
        read_text, read_bytes, glob = Path.read_text, Path.read_bytes, Path.glob
        write_text, write_bytes = Path.write_text, Path.write_bytes
        source_overlay = {}
        commands = []
        real_run = subprocess.run
        output = io.StringIO()
        with tempfile.TemporaryDirectory() as temp:
            out = Path(temp)
            def text(path, *args, **kwargs):
                return data if path == STATUS else source_overlay[path].decode() if path in source_overlay else read_text(path, *args, **kwargs)
            def raw(path, *args, **kwargs):
                return data.encode() if path == STATUS else source_overlay[path] if path in source_overlay else read_bytes(path, *args, **kwargs)
            def put_text(path, value, *args, **kwargs):
                if path == ROOT/'stage3/fixtures/host/25_readDirectory.a':
                    source_overlay[path]=value.encode();return len(value)
                return write_text(path,value,*args,**kwargs)
            def put_bytes(path, value, *args, **kwargs):
                if path == ROOT/'stage3/fixtures/host/25_readDirectory.a':
                    source_overlay[path]=value;return len(value)
                return write_bytes(path,value,*args,**kwargs)
            def paths(path, pattern, *args, **kwargs):
                if path == ROOT / 'stage3/fixtures' and pattern == '*/status.json':
                    return iter([STATUS])
                return glob(path, pattern, *args, **kwargs)
            def execute(command, *args, **kwargs):
                if len(command)>1 and command[1]=='-p':
                    return real_run(command,*args,**kwargs)
                commands.append(command)
                if mode=='unset' and not metadata:
                    if reader.endswith('check.py'):
                        self.assertTrue(command[0]=='go' or command[0]=='/unused',command)
                    else:
                        self.assertIn('source-audit.cjs',str(command))
                    raise AtNode()
                if not metadata or reader == 'stage3/records-maplike-rebuild/audit-status.py':
                    self.assertEqual(command[0], selected)
                    result = real_run(command, *args, **kwargs)
                    self.assertEqual(result.returncode, row['node']['exit'], result.stderr)
                    self.assertEqual(result.stdout.decode(), row['node']['stdout'])
                    self.assertEqual(result.stderr.decode(), row['node']['stderr'])
                    if reader == 'stage3/records-maplike-rebuild/audit-status.py':
                        return result
                raise AtNode()
            argv = {
                'stage3/fixtures/host/check.py': [reader, '--logs', str(out/'logs')],
                'internal/oracle/node_fs_file_host_check.py': [reader, '--compiler', '/unused', '--logs', str(out/'logs'), '--report', str(out/'report.json')],
                'stage3/fixtures/host/error-host-proof.py': [reader, str(ROOT), str(out/'adapted'), str(out/'logs')],
                'stage3/fixtures/audit-shards.py': [reader, '/unused-fixtures-test', str(out/'audit'), '--count', '1'],
                'stage3/records-maplike-rebuild/audit-status.py': [reader],
            }[reader]
            selected = str(out / 'selected-node')
            Path(selected).write_text("#!/bin/sh\nif [ \"$1\" = -p ]; then echo 'v24.19.0 false'; exit 0; fi\nexec '" + shutil.which('node') + "' \"$@\"\n")
            Path(selected).chmod(0o700)
            with patch.dict(os.environ, {SWITCH: selected if mode == 'selected' else ''}), patch.object(sys, 'argv', argv), patch.object(Path, 'read_text', text), patch.object(Path, 'read_bytes', raw), patch.object(Path, 'glob', paths), patch.object(Path,'write_text',put_text), patch.object(Path,'write_bytes',put_bytes), patch('subprocess.run', execute), patch('subprocess.check_output', return_value=before_data.encode()), contextlib.redirect_stdout(output):
                error = None
                try:
                    runpy.run_path(str(ROOT/reader), run_name='__main__')
                except AtNode:
                    pass
                except SystemExit as exc:
                    if exc.code not in (0, None):
                        raise
                except ValueError as exc:
                    error = exc
                if mode == 'unknown':
                    self.assertIsNotNone(error, output.getvalue())
                    self.assertIn('unknown host configuration', str(error))
                    return
                self.assertIsNone(error)
                if mode == 'unset':
                    self.assertIn(f'SKIP: row 1 needs node-without-inspector; set {SWITCH}', output.getvalue())
                    if not metadata:
                        self.assertTrue(commands, 'unset reference build skipped non-Node checks')
                        self.assertFalse(any(command[0]==selected for command in commands))
                else:
                    self.assertNotIn('SKIP:', output.getvalue())
                    if not metadata:
                        self.assertEqual(commands[0][0], selected, 'reader ignored selected Node')
                if reader == 'stage3/fixtures/audit-shards.py':
                    self.assertIn('host/'+filename, json.loads((out/'audit/cases.json').read_text()))

    def test_python_execution_readers(self):
        for reader, filename in [
            ('stage3/fixtures/host/check.py', '01_readFile_utf8.a'),
            ('internal/oracle/node_fs_file_host_check.py', '01_readFile_utf8.a'),
            ('stage3/fixtures/host/error-host-proof.py', '25_readDirectory.a'),
        ]:
            for mode in ('unset', 'selected', 'unknown'):
                with self.subTest(reader=reader, mode=mode):
                    self.run_reader(reader, filename, mode)

    def test_python_glob_readers_keep_rows(self):
        for reader in ('stage3/fixtures/audit-shards.py', 'stage3/records-maplike-rebuild/audit-status.py'):
            for mode in ('unset', 'selected', 'unknown'):
                with self.subTest(reader=reader, mode=mode):
                    self.run_reader(reader, '01_readFile_utf8.a', mode, metadata=True)


    def test_configuration_addition_is_not_a_stock_comparison(self):
        for mode in ('unset', 'selected'):
            with self.subTest(mode=mode):
                self.run_reader('stage3/records-maplike-rebuild/audit-status.py',
                                '01_readFile_utf8.a', mode, metadata=True, added_configuration=True)

    def test_fs_javascript_uses_configured_node(self):
        source = ast.parse((ROOT/'internal/oracle/node_fs_file_host_check.py').read_text())
        statement = next(node for node in source.body if isinstance(node, ast.With))
        row = copy.deepcopy(ORIGINAL[0])
        row['configuration'] = CONFIG
        for binary in (None, '/verified-node'):
            with self.subTest(binary=binary), tempfile.TemporaryDirectory() as temp:
                bucket = Path(temp)
                (bucket/'status.json').write_text(json.dumps([row]))
                commands = []
                def run(command, label):
                    return row['node'] if label.endswith('.native') else dict(stdout='compiled JavaScript', stderr='', exit=0)
                def node(path, label, selected):
                    commands.append((label, selected))
                    self.assertEqual(selected, binary, 'generated JavaScript ignored configured Node')
                    return row['node']
                scope = dict(tempfile=tempfile, Path=Path, json=json, bucket=bucket, report=[],
                             owned={'01'}, node_for=lambda *args: binary, node=node, run=run,
                             stage=lambda result: 'Compiles',
                             args=types.SimpleNamespace(compiler='/compiler', mutants=False))
                with contextlib.redirect_stdout(io.StringIO()):
                    exec(compile(ast.Module(body=[statement], type_ignores=[]), 'node_fs_file_host_check.py', 'exec'), scope)
                self.assertEqual(len(commands), 0 if binary is None else 2)
                self.assertTrue(scope['report'][0]['native']['agrees'])

    def test_python_reporting_does_not_claim_skipped_rows(self):
        source = ast.parse((ROOT/'stage3/fixtures/host/error-host-proof.py').read_text())
        start = next(index for index, node in enumerate(source.body)
                     if isinstance(node, ast.Assign) and any(isinstance(target, ast.Name) and target.id == 'proof' for target in node.targets))
        tail = ast.Module(body=source.body[start:], type_ignores=[])
        for missing in (False, True):
            with self.subTest(missing=missing), tempfile.TemporaryDirectory() as temp:
                observed = {'one.a': {'stdout': 'one'}, 'two.a': {'skipped': True} if missing else {'stdout': 'two'}}
                scope = dict(observed=observed, rows=[{}, {}], mutants=[], out=Path(temp), json=json,
                             p=types.SimpleNamespace(returncode=0), fixture_node=None if missing else 'node',
                             subprocess=types.SimpleNamespace(check_output=lambda args: b'v24.19.0\n'))
                output = io.StringIO()
                with contextlib.redirect_stdout(output):
                    exec(compile(tail, 'error-host-proof.py', 'exec'), scope)
                report = json.loads((Path(temp)/'proof.json').read_text())
                self.assertEqual(report['unchangedBeforeAndGoldens'], not missing)
                if missing:
                    self.assertEqual(report['skipped'], ['two.a'])
                    self.assertEqual(report['compared'], 1)
                    self.assertIn('two.a', output.getvalue())
                else:
                    self.assertNotIn('skipped', report)


if __name__ == '__main__':
    unittest.main()
