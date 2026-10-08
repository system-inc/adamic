#!/usr/bin/env python3
"""Observe source Node and both compiler frontiers, without claiming admission."""
import json
import subprocess
import sys
from pathlib import Path

lane = Path(__file__).resolve().parent
root = lane.parents[2]
expected = {'diagnostic-good': 'plain\nnested\n', 'diagnostic-wrong': '42\n',
            'node-indicator-good': 'true\nIdentifier\nabsent\n', 'node-indicator-wrong': '42\n',
            'jsdoc-good': 'plain\nnested\nabsent\n', 'jsdoc-wrong': '42\n',
            'option-type-good': 'string\n42\n', 'option-type-wrong': '42\n'}
observations = []
for fixture in sorted((lane / 'fixtures').glob('*.a')):
    runs = {}
    commands = {'source_node': ['node', '--input-type=module-typescript', '-e', fixture.read_text()],
                'native_compiler': [sys.argv[1], 'c', str(fixture)],
                'javascript_compiler': [sys.argv[1], 'js', str(fixture)]}
    for backend, command in commands.items():
        result = subprocess.run(command, cwd=root, capture_output=True, text=True)
        runs[backend] = {'exit': result.returncode, 'stdout': result.stdout, 'stderr': result.stderr}
        (lane / 'logs' / f'{fixture.stem}-{backend}.log').write_text(
            f'exit={result.returncode}\nstdout:\n{result.stdout}\nstderr:\n{result.stderr}')
    if fixture.stem == 'option-type-wrong':
        assert runs['source_node']['exit'] == 1, runs['source_node']
        assert 'TypeError: member.get is not a function' in runs['source_node']['stderr'], runs['source_node']
    else:
        assert runs['source_node']['exit'] == 0, runs['source_node']
        assert runs['source_node']['stdout'] == expected[fixture.stem], runs['source_node']
    observations.append({'fixture': fixture.name, 'runs': runs})
    print(f"{fixture.stem}: source Node {runs['source_node']['exit']}; C compiler {runs['native_compiler']['exit']}; JS compiler {runs['javascript_compiler']['exit']}")
(lane / 'frontier-observations.json').write_text(json.dumps(observations, indent=2) + '\n')
print('Compiler success alone is not runtime parity or completed demand.')
