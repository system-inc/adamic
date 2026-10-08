#!/usr/bin/env python3
"""Rerun the eleven unchanged native stop probes with strict byte comparisons."""
import argparse
import json
from pathlib import Path
import subprocess

p = argparse.ArgumentParser()
p.add_argument('compiler', type=Path)
p.add_argument('scratch', type=Path)
p.add_argument('output', type=Path)
a = p.parse_args()
a.compiler, a.scratch, a.output = a.compiler.resolve(), a.scratch.resolve(), a.output.resolve()
a.output.mkdir(exist_ok=False)
here = Path(__file__).resolve().parent
previous = json.loads((here / 'evidence/front10/build/report.json').read_text())
rows = []

def run(command, stem):
    with (a.output / (stem + '.stdout')).open('wb') as stdout, (a.output / (stem + '.stderr')).open('wb') as stderr:
        code = subprocess.run(command, cwd=a.scratch, stdout=stdout, stderr=stderr).returncode
    return {'exit': code, 'stdout': (a.output / (stem + '.stdout')).read_text(),
            'stderr': (a.output / (stem + '.stderr')).read_text()}

for previous_probe in previous['minimal_programs']:
    file = previous_probe['file']
    binary = a.output / (file + '.bin')
    row = {'file': file}
    row['build'] = run([str(a.compiler), 'build', str(here / file), '-o', str(binary)], file + '.build')
    row['node'] = run(['node', '--disable-warning=ExperimentalWarning', str(a.scratch / 'oracle/node.mjs'),
                       str(here / file)], file + '.node')
    row['pass'] = False
    if row['build']['exit'] == 0:
        row['native'] = run([str(binary)], file + '.native')
        comparisons = {stream: run(['cmp', str(a.output / (file + '.node.' + stream)),
                                    str(a.output / (file + '.native.' + stream))],
                                   file + '.' + stream + '-cmp')['exit'] for stream in ['stdout', 'stderr']}
        row['comparison'] = comparisons
        row['pass'] = row['native']['exit'] == row['node']['exit'] == 0 and all(c == 0 for c in comparisons.values())
        if row['pass']:
            data = (a.output / (file + '.native.stdout')).read_bytes()
            assert data, 'output-byte mutant requires nonempty output'
            mutant = a.output / (file + '.mutant.stdout')
            mutant.write_bytes(bytes([data[0] ^ 1]) + data[1:])
            row['native_byte_mutant_cmp'] = run(['cmp', str(a.output / (file + '.node.stdout')), str(mutant)],
                                                file + '.mutant-cmp')['exit']
            assert row['native_byte_mutant_cmp'] == 1
    rows.append(row)
report = {'probes': rows, 'passing': sum(r['pass'] for r in rows), 'total': len(rows),
          'definition': 'unchanged source builds; native and Node exits zero, stdout/stderr byte-identical',
          'node_runner': 'scratch oracle/node.mjs, Node transform mode for namespace syntax'}
(a.output / 'report.json').write_text(json.dumps(report, indent=2) + '\n')
print(json.dumps(report, indent=2))
