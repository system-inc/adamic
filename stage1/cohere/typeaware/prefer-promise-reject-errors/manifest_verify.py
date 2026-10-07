#!/usr/bin/env python3
"""Compare manifest kinds against an independent Go registration oracle."""
import argparse
import json
from pathlib import Path
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('--oracle', type=Path, required=True)
parser.add_argument('--artifacts', type=Path, required=True)
args = parser.parse_args()
base = Path(__file__).resolve().parent.parent
root = base.parents[2]
args.artifacts.mkdir(parents=True, exist_ok=True)
with (args.artifacts / 'oracle.stdout').open('wb') as stdout, (args.artifacts / 'oracle.stderr').open('wb') as stderr:
    subprocess.run([str(args.oracle), str(root / 'cohere')], stdout=stdout, stderr=stderr, check=True)
assert not (args.artifacts / 'oracle.stderr').read_bytes()
truth = (args.artifacts / 'oracle.stdout').read_bytes()
rows = [('floating', 'no-floating-promises'), ('eval', 'no-implied-eval'), ('void', 'no-meaningless-void-operator'), ('output', 'no-process-exit-after-output'), ('timer', 'no-uncleared-race-timeout'), ('blocking', 'require-blocking-standard-streams'), ('promise', 'prefer-promise-reject-errors'), ('regex', 'prefer-regex-literals'), ('rest', 'prefer-rest-params'), ('effect', 'react-hooks-set-state-in-effect'), ('render', 'react-hooks-set-state-in-render'), ('static', 'react-hooks-static-components')]
records = []
for label, directory in rows:
    obj = json.loads((base / directory / 'rule.json').read_text())
    assert set(obj) == {'kinds'}
    kinds = obj['kinds']
    assert kinds and all(type(kind) is int and kind >= 0 for kind in kinds)
    assert len(set(kinds)) == len(kinds)
    records.append((label, kinds))
def serialize(values):
    return ''.join(label + ' ' + ','.join(map(str, kinds)) + '\n' for label, kinds in values).encode()
assert serialize(records) == truth, 'manifest differs from production Go registrations'
mutants = []
for index, (label, kinds) in enumerate(records):
    changed = [(name, values[:]) for name, values in records]
    changed[index][1][0] += 1
    assert serialize(changed) != truth, label + ': manifest mutation escaped oracle'
    mutants.append({'rule': label, 'mutation': 'first kind plus one', 'caught': True})
(args.artifacts / 'mutants.json').write_text(json.dumps(mutants, indent=2) + '\n')
print('PASS 12 manifests, 136 production registration bytes, 12 metadata mutants')
