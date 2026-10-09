#!/usr/bin/env python3
"""Run the unchanged fast gate's aCheck method with an already-built topic compiler."""
import argparse
import json
from pathlib import Path
import runpy
import shutil
import subprocess
import tempfile
from types import SimpleNamespace

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('gate_file', type=Path)
parser.add_argument('compiler', type=Path)
parser.add_argument('logs', type=Path)
args = parser.parse_args()
Gate = runpy.run_path(str(args.gate_file.resolve()))['Gate']
root = Path(__file__).resolve().parent
logs = args.logs.resolve()
logs.mkdir(parents=True, exist_ok=True)
entries = json.loads((root / 'status.json').read_text())


class FocusedGate:
    def __init__(self, tree):
        self.arguments = SimpleNamespace(tree=str(tree))
        self.steps, self.exits, self.result, self.failures = {}, {}, {}, []

    def step(self, name, command):
        # Only the build step is reused. aCheck's classification, header parsing,
        # expectation match, reporting, and failure path execute unchanged.
        assert name == 'a-check-build'
        return True

    def spawn(self, command, stdout, stderr):
        return subprocess.Popen(command, cwd=self.arguments.tree, stdout=stdout, stderr=stderr, text=True)

    def fail(self, name, message):
        self.failures.append({'check': name, 'message': message})


with tempfile.TemporaryDirectory(prefix='adamic-real-headers-') as scratch:
    tree = Path(scratch) / 'tree'
    paths = []
    for entry in entries:
        relative = Path('stage3/fixtures/real/bytes-revealed') / entry['file']
        path = tree / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(root / entry['file'], path)
        paths.append(str(relative))
    binaries = Path(str(tree) + '-binaries')
    binaries.mkdir()
    shutil.copyfile(args.compiler.resolve(), binaries / 'adamic-acheck')
    (binaries / 'adamic-acheck').chmod(0o755)
    gate = FocusedGate(tree)
    Gate.aCheck(gate, paths)
    assert gate.exits['a-check'] == 0, gate.failures
    (logs / 'headers.json').write_text(json.dumps(gate.result, indent=2) + '\n')
    selected = next(path for path in paths if '/computed-field/' in path)
    path = tree / selected
    original = path.read_text()
    path.write_text('// a-check: refused nonexistent-mutant\n' + original.split('\n', 1)[1])
    mutant = FocusedGate(tree)
    Gate.aCheck(mutant, [selected])
    assert mutant.exits['a-check'] == 1 and len(mutant.failures) == 1
    assert 'expected refused nonexistent-mutant' in mutant.failures[0]['message']
    (logs / 'header-mutant.json').write_text(json.dumps({'exit': 1, 'caught_by': mutant.failures, 'observations': mutant.result}, indent=2) + '\n')
print('PASS: 10 topic a-check headers; contradictory-header mutant caught by unchanged Gate.aCheck')
