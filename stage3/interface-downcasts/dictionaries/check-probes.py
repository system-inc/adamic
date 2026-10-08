"""Observe source Node and the current compiler frontier; no support claims."""
import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
HERE = Path(__file__).resolve().parent
cases = {'options-valid.a': 'true\nundefined\n', 'options-wrong.a': '42\n',
         'nested-wrong.a': '42\n'}
observations = []
compiler = Path(sys.argv[1]) if len(sys.argv) == 2 else None
for name, expected in cases.items():
    fixture = HERE / 'fixtures' / name
    node = subprocess.run(['node', '--experimental-strip-types',
                           '--input-type=module-typescript'],
                          input=fixture.read_text(), text=True, capture_output=True)
    (HERE / 'logs' / (name + '.node.log')).write_text(node.stdout + node.stderr)
    assert (node.returncode, node.stdout, node.stderr) == (0, expected, ''), name
    row = {'fixture': name, 'source_node_exit': 0, 'source_node_stdout': expected,
           'source_support': 'unclaimed'}
    if compiler is not None:
        for backend in ['c', 'js']:
            result = subprocess.run([str(compiler), backend, str(fixture)],
                                    cwd=ROOT, text=True, capture_output=True)
            (HERE / 'logs' / (name + '.' + backend + '.log')).write_text(
                result.stdout + result.stderr)
            row[backend] = {'exit': result.returncode, 'stderr': result.stderr}
    observations.append(row)
(HERE / 'probe-observations.json').write_text(json.dumps(observations, indent=2) + '\n')
print('Source Node controls: 3 passed; compiler support remains unclaimed')
