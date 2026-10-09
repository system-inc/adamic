#!/usr/bin/env python3
"""Run fixtures, mutants, the 301 CLI projects and upstream compiler tests."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

here = Path(__file__).resolve().parent
repo = here.parents[2]
parser = argparse.ArgumentParser()
parser.add_argument('tree', type=Path)
parser.add_argument('output', type=Path)
args = parser.parse_args()
tree, output = args.tree.resolve(), args.output.resolve()
output.mkdir(parents=True, exist_ok=False)
lib = tree / 'built/local/typescript.js'
base = dict(os.environ)
base['NODE_OPTIONS'] = '--require=' + str(here / 'probe.cjs')
results = []

def execute(name, argv, cwd=repo, expected=0, extra=None):
    env = dict(base, **(extra or {}))
    counts = output / (name + '-counts')
    counts.mkdir()
    env['TWO_WRITES_RESULTS'] = str(counts)
    with (output / (name + '.log')).open('w') as log:
        done = subprocess.run(argv, cwd=cwd, env=env, stdout=log, stderr=subprocess.STDOUT)
    rows, examples = [], []
    for file in sorted(counts.glob('*.json')):
        data = json.loads(file.read_text())
        rows.extend(data['rows'])
        examples.extend(data['examples'])
    totals = {}
    for site in ['flags', 'parent']:
        totals[site] = {key: sum(row[key] for row in rows if row['site'] == site)
                        for key in ['assignments', 'violations', 'reads', 'outOfTypeReads']}
    record = {'name': name, 'argv': argv, 'cwd': str(cwd), 'exit': done.returncode,
              'expected_exit': expected, 'process_records': len(list(counts.glob('*.json'))),
              'totals': totals, 'rows': rows, 'examples': examples}
    (output / (name + '.json')).write_text(json.dumps(record, indent=2) + '\n')
    results.append(record)
    print(name, 'exit=' + str(done.returncode), json.dumps(totals), flush=True)
    assert done.returncode == expected, f'{name}: see {output / (name + ".log")}'
    if name.startswith('mutant-'):
        assert 'AssertionError' in (output / (name + '.log')).read_text()

for fixture in ['fixtures', 'flags', 'parent', 'typed']:
    execute(fixture, ['node', str(here / (fixture + '.cjs')), str(lib)],
            extra={'TWO_WRITES_INPUT': fixture + '-witness'})
for site in ['flags', 'parent']:
    execute('mutant-' + site, ['node', str(here / 'fixtures.cjs'), str(lib)], expected=1,
            extra={'TWO_WRITES_MUTANT': site, 'TWO_WRITES_INPUT': 'compatible-control'})
for site in ['flags', 'parent']:
    execute('mutant-reads-' + site, ['node', str(here / (site + '.cjs')), str(lib)], expected=1,
            extra={'TWO_WRITES_MUTANT': 'reads-' + site, 'TWO_WRITES_INPUT': site + '-witness'})
execute('static', ['node', str(here / 'static.cjs'), str(lib), str(tree)])
execute('acceptance', ['bash', str(repo / 'stage3/drivers/tsc/run.sh'), 'node', str(tree / 'built/local/tsc.js')],
        extra={'TSC_RESULTS': str(output / 'acceptance-projects')})
execute('compiler', ['npm', 'test', '--', '--runners=compiler', '--workers=4', '--lint=false'], cwd=tree)
summary = {'upstream_commit': subprocess.check_output(['git', '-C', str(tree), 'rev-parse', 'HEAD'], text=True).strip(),
           'node': subprocess.check_output(['node', '--version'], text=True).strip(),
           'probe_sha256': hashlib.sha256((here / 'probe.cjs').read_bytes()).hexdigest(),
           'hashes': {str(p.relative_to(tree)): hashlib.sha256(p.read_bytes()).hexdigest()
                      for p in [tree / 'src/compiler/utilities.ts', lib, tree / 'built/local/_tsc.js']},
           'runs': [{k: v for k, v in r.items() if k not in ['rows', 'examples']} for r in results]}
(output / 'summary.json').write_text(json.dumps(summary, indent=2) + '\n')
