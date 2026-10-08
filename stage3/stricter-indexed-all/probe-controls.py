#!/usr/bin/env python3
"""Prove the scoped exclusion boundary and production report coverage fail closed."""
import argparse
import concurrent.futures
import copy
import json
import subprocess
from pathlib import Path

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--scratch', type=Path, required=True)
parser.add_argument('--tree', type=Path, required=True)
args = parser.parse_args()
root = Path(__file__).resolve().parent
scratch = args.scratch.resolve()
result = json.loads((scratch / 'result.json').read_text())
exclusions = json.loads((scratch / 'exclusions.json').read_text())
controls = scratch / 'controls'
controls.mkdir(exist_ok=True)

loader_cases = {}
loader_cases['missing-exclusion'] = (exclusions[:-1], 'want exactly 72 exclusions, got 71')
wrong = copy.deepcopy(exclusions)
wrong[0]['site']['line'] += 1
loader_cases['wrong-row-identity'] = (wrong, 'exclusions lack exact option audit evidence')
wrong = copy.deepcopy(exclusions)
joint = next(site for site in result['option_sites'] if len(site['options']) == 2)
# Pretend the joint indexed/optional row is an optional-only exclusion. The
# loader must reject its real joint attribution and keep D069 indexed-owned.
wrong[0]['id'] = 'D069'
wrong[0]['site'] = {**joint, 'options': ['exactOptionalPropertyTypes']}
loader_cases['exclude-joint-indexed-row'] = (wrong, 'exclusion attribution changed')

def run_loader(item):
    name, (rows, expected) = item
    manifest = controls / (name + '-exclusions.json')
    manifest.write_text(json.dumps(rows))
    run = subprocess.run([str(scratch / 'production-probe'), str(args.tree),
                          str(root / 'evidence/ledger-options.json'), str(manifest),
                          str(controls / name)], capture_output=True, text=True)
    (controls / (name + '.json')).write_text(run.stdout)
    (controls / (name + '.log')).write_text(run.stderr)
    observed = json.loads(run.stdout)
    assert run.returncode == 1 and observed['stage'] == 'checker'
    assert expected in observed['error'], observed
    return name + ': caught before lowering: ' + observed['error']

with concurrent.futures.ThreadPoolExecutor(max_workers=3) as executor:
    for outcome in executor.map(run_loader, loader_cases.items()):
        print(outcome)

for name in ['missing-indexed-audit', 'fabricated-built-entry', 'wrong-excluded-id']:
    mutant = copy.deepcopy(result)
    if name == 'missing-indexed-audit':
        index = next(i for i, site in enumerate(mutant['option_sites']) if 'noUncheckedIndexedAccess' in site['options'])
        del mutant['option_sites'][index]
    elif name == 'fabricated-built-entry':
        mutant['attempts'][0]['stage'] = 'built'
        mutant['attempts'][0]['binary'] = '/tmp/fabricated-binary'
    else:
        mutant['excluded'][0]['id'] = 'D069'
    path = controls / (name + '.json')
    path.write_text(json.dumps(mutant))
    run = subprocess.run(['python3', str(root / 'classify-production.py'),
                          '--result', str(path), '--tree', str(args.tree),
                          '--output', str(controls / name)], capture_output=True, text=True)
    assert run.returncode != 0, name
    print(name + ': caught by classifier: ' + run.stderr.strip())
