#!/usr/bin/env python3
"""Exercise ledger cache repair, complete keys, output isolation and exports."""
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
HELPER = Path(os.environ.get('ADAMIC_GATE_INPUTS_MODULE', SOURCE / 'setup-gate-inputs.py'))
spec = importlib.util.spec_from_file_location('ledger_gate', HELPER)
gate = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gate)


class Ledger(unittest.TestCase):
    def fixture(self, root):
        upstream = root / 'upstream'
        (upstream / 'scripts').mkdir(parents=True)
        (upstream / 'src/compiler').mkdir(parents=True)
        (upstream / gate.LEDGER_SCRIPT).write_text('upstream-generator-v1')
        (upstream / gate.LEDGER_DIAGNOSTICS).write_text('{"input":1}')
        with (root / 'git.log').open('wb') as log:
            for args in [['git', 'init'], ['git', 'add', '.'],
                         ['git', '-c', 'user.name=Ledger proof', '-c', 'user.email=proof@example.invalid',
                          'commit', '-m', 'Pinned fixture']]:
                subprocess.run(args, cwd=upstream, stdout=log, stderr=subprocess.STDOUT, check=True)
        pin = gate.command(['git', 'rev-parse', 'HEAD'], upstream)
        node = root / 'node'
        version = json.loads((SOURCE / 'node-pin.json').read_text())['version']
        node.write_text('#!' + sys.executable + '\n' +
                        'import sys\nfrom pathlib import Path\n' +
                        'if sys.argv[1] == "--version": print(' + repr(version) + ')\n' +
                        'else:\n    data = Path(sys.argv[1]).read_bytes() + Path(sys.argv[2]).read_bytes()\n' +
                        '    for name in ' + repr(gate.LEDGER_GENERATED) + ': Path(name).write_bytes(data)\n')
        node.chmod(0o755)
        inputs = root / 'inputs'
        inputs.mkdir()
        return upstream, pin, str(node), inputs

    def test_every_ledger_key_component(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            upstream, pin, node, inputs = self.fixture(root)
            before = gate.ledger_key(upstream, 'v24.19.0')
            for name in ['TS_COMMIT', 'TS_URL', 'LEDGER_SPARSE']:
                with self.subTest(name=name), patch.object(gate, name, 'changed'):
                    self.assertNotEqual(before, gate.ledger_key(upstream, 'v24.19.0'), name)
            with patch.object(gate, 'source_hash', return_value='changed'):
                self.assertNotEqual(before, gate.ledger_key(upstream, 'v24.19.0'), 'helper')
            self.assertNotEqual(before, gate.ledger_key(upstream, 'v24.21.0'), 'node')
            for name in [gate.LEDGER_SCRIPT, gate.LEDGER_DIAGNOSTICS]:
                path = upstream / name
                original = path.read_bytes()
                path.write_bytes(original + b'changed')
                self.assertNotEqual(before, gate.ledger_key(upstream, 'v24.19.0'), name)
                path.write_bytes(original)

    def test_install_skip_repair_output_and_uncached_bytes(self):
        with tempfile.TemporaryDirectory() as temporary:
            upstream, pin, node, inputs = self.fixture(Path(temporary))
            with patch.object(gate, 'TS_URL', str(upstream)), patch.object(gate, 'TS_COMMIT', pin), patch.dict(os.environ, ADAMIC_GATE_UNCACHED='0'):
                self.assertIn('installed', gate.cycle_ledger(inputs, node))
                checkout = inputs / 'cycle-ledger'
                original = gate.artifact_digest(checkout)
                self.assertIn('skipped', gate.cycle_ledger(inputs, node))
                output = inputs / gate.VARIABLES['ADAMIC_CYCLE_LEDGER_OUTPUT']
                output.write_text('test produced a ledger')
                self.assertIn('skipped', gate.cycle_ledger(inputs, node))
                self.assertEqual(output.read_text(), 'test produced a ledger')
                for name in [gate.LEDGER_SCRIPT, gate.LEDGER_DIAGNOSTICS, *gate.LEDGER_GENERATED]:
                    (checkout / name).write_text('corrupt')
                    self.assertIn('installed', gate.cycle_ledger(inputs, node), name)
                    self.assertEqual(original, gate.artifact_digest(checkout))
                (checkout / gate.LEDGER_GENERATED[1]).unlink()
                self.assertIn('installed', gate.cycle_ledger(inputs, node))
                with patch.dict(os.environ, ADAMIC_GATE_UNCACHED='1'):
                    self.assertIn('installed', gate.cycle_ledger(inputs, node))
                    self.assertEqual(original, gate.artifact_digest(checkout))

    def test_missing_generated_file_and_wrong_head_refused(self):
        with tempfile.TemporaryDirectory() as temporary:
            upstream, pin, node, inputs = self.fixture(Path(temporary))
            with patch.object(gate, 'TS_URL', str(upstream)), patch.object(gate, 'TS_COMMIT', pin):
                gate.cycle_ledger(inputs, node)
                checkout = inputs / 'cycle-ledger'
                (checkout / gate.LEDGER_GENERATED[0]).unlink()
                with self.assertRaisesRegex(ValueError, 'missing generated diagnostics'):
                    gate.validate_ledger(checkout)
                with patch.object(gate, 'TS_COMMIT', 'wrong'):
                    with self.assertRaisesRegex(ValueError, 'source pin'):
                        gate.validate_ledger(checkout)
            with patch.object(gate, 'command', return_value='v24.21.0'):
                with self.assertRaisesRegex(ValueError, 'expected v24.19.0, got v24.21.0'):
                    gate.cycle_ledger(inputs, node)

    def test_flag_exports_and_ordinary_unset(self):
        for phase, exported in [('env', True), ('env-no-archive', True), ('env-archive', False)]:
            with self.subTest(phase=phase), tempfile.TemporaryDirectory() as temporary:
                with patch.object(sys, 'argv', ['setup', phase, '/different/main', temporary, 'node']), contextlib.redirect_stdout(io.StringIO()) as output:
                    gate.main()
                lines = output.getvalue().splitlines()
                for name in ['ADAMIC_CYCLE_LEDGER_ROOT', 'ADAMIC_CYCLE_LEDGER_OUTPUT']:
                    if exported:
                        self.assertIn('export ' + name + '=' + shlex.quote(str(Path(temporary) / gate.VARIABLES[name])), lines)
                    else:
                        self.assertIn('unset ' + name, lines)
                    self.assertIn(name, (SOURCE / 'setup.sh').read_text().split('echo "unset ', 1)[1].split('"', 1)[0].split())
                self.assertNotEqual(gate.VARIABLES['ADAMIC_CYCLE_LEDGER_ROOT'], gate.VARIABLES['ADAMIC_TYPESCRIPT_SOURCE'])


if __name__ == '__main__':
    unittest.main()
