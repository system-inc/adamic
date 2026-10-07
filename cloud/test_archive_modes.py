#!/usr/bin/env python3
"""Assert archive selection cannot leak through the other gate inputs."""
import contextlib
import importlib.util
import io
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
gate = importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)

class ArchiveModes(unittest.TestCase):
    def test_environment_modes(self):
        for phase in ['env', 'env-no-archive', 'env-archive']:
            with self.subTest(phase=phase), patch.object(sys, 'argv', ['setup', phase, '/tmp/repo', '/tmp/inputs', 'node']), contextlib.redirect_stdout(io.StringIO()) as output:
                gate.main()
            lines = [shlex.split(line) for line in output.getvalue().splitlines()]
            exports = {line[1].split('=')[0] for line in lines if line[0]=='export'}
            cleared = {line[1] for line in lines if line[0]=='unset'}
            archive = {'ADAMIC_CLANG_TSGO_ARCHIVE'}
            expected = set(gate.VARIABLES) if phase=='env' else archive if phase=='env-archive' else set(gate.VARIABLES)-archive
            self.assertEqual(exports, expected)
            self.assertEqual(cleared, set(gate.VARIABLES)-expected)

    def test_darwin_selection_calls_and_exports(self):
        spec = importlib.util.spec_from_file_location('darwin', os.environ.get('ADAMIC_ARCHIVE_DARWIN_MODULE', SOURCE / 'setup-darwin.py'))
        darwin = importlib.util.module_from_spec(spec);spec.loader.exec_module(darwin)
        for arguments, phases, names in [([], [], set()),
                (['--gate-inputs'], ['npm', 'corpora', 'archive'], set(gate.VARIABLES)),
                (['--gate-inputs-no-archive'], ['npm', 'corpora'], set(gate.VARIABLES)-{'ADAMIC_CLANG_TSGO_ARCHIVE'}),
                (['--gate-archive'], ['archive'], {'ADAMIC_CLANG_TSGO_ARCHIVE'})]:
            with self.subTest(arguments=arguments), tempfile.TemporaryDirectory() as temporary:
                calls = []
                def run(command, **kwargs):
                    if 'setup-gate-inputs.py' in ' '.join(command):calls.append(command[2])
                def output(command, **kwargs):
                    if command[:2]==['go','version']:return 'go version go1.27.1'
                    if command[:2]==['clang','--version']:return 'clang'
                    if command[-1]=='--version':return 'v24.19.0'
                    if 'rev-parse' in command:return 'commit'
                    if 'setup-key.py' in ' '.join(command):return 'key'
                    if 'setup-gate-inputs.py' in ' '.join(command):
                        with patch.object(gate.sys, 'argv', command[1:]), contextlib.redirect_stdout(io.StringIO()) as captured:gate.main()
                        return captured.getvalue().strip()
                    raise AssertionError(command)
                with patch.dict(os.environ, ADAMIC_TOOLS=temporary, ADAMIC_GATE_UNCACHED='0'), patch.object(darwin.sys,'argv',['setup', *arguments]), patch.object(darwin.node,'prepare',return_value='ready'), patch.object(darwin.shutil,'which',return_value='go'), patch.object(darwin,'output',side_effect=output), patch.object(darwin.subprocess,'run',side_effect=run), contextlib.redirect_stdout(io.StringIO()):
                    darwin.main()
                self.assertEqual(calls, phases)
                environment=(Path(temporary)/'env.sh').read_text()
                exports={line.split('=',1)[0].split()[1] for line in environment.splitlines() if line.startswith('export ADAMIC_')}-{'ADAMIC_MARKDOWNWIDTH_DEPS'}
                self.assertEqual(exports,names)
                for name in set(gate.VARIABLES)-names:self.assertIn(name, environment)

    def test_conflicting_flags_fail_before_preparation(self):
        for extra in ['--gate-inputs', '--gate-archive']:
            for arguments in [['--gate-inputs-no-archive', extra], [extra, '--gate-inputs-no-archive']]:
                result = subprocess.run(['bash', str(SOURCE/'setup.sh'), *arguments], capture_output=True, text=True)
                self.assertEqual(result.returncode, 2, result.stdout+result.stderr)
                self.assertIn('conflicts', result.stderr)

if __name__=='__main__':unittest.main()
