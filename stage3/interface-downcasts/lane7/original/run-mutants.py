#!/usr/bin/env python3
"""Counterfactuals for complete original intersection contracts, in three modes."""
import os
from pathlib import Path
import subprocess
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: run-mutants.py <declarations> <logs>')
root = Path(__file__).resolve().parents[4]
declarations, logs = map(lambda value: Path(value).resolve(), sys.argv[1:])
logs.mkdir(parents=True, exist_ok=True)
for kind, fixture in [('skip', 'tracker-wrong'), ('shape', 'tracker-wrong'),
                      ('nested', 'tracker-wrong'), ('presence', 'tracker-missing'),
                      ('outer', 'tracker-root-wrong'), ('absence', 'tracker-absent')]:
    path = logs / (kind + '.log')
    env = dict(os.environ, ADAMIC_INTERSECTION_ORIGINAL_DECLS=str(declarations),
               ADAMIC_INTERSECTION_ORIGINAL_MUTANT=kind)
    with path.open('w') as log:
        result = subprocess.run(['go', 'test', './internal/oracle', '-run',
            '^TestCheckedViewIntersectionOriginalPairs/' + fixture + '$', '-count=1', '-v'],
            cwd=root, env=env, stdout=log, stderr=subprocess.STDOUT)
    output = path.read_text()
    if result.returncode == 0 or output.count('got oracle.run') != 3:
        raise SystemExit('mutant escaped refusal pins: ' + str(path))
    expected = 'exitCode:70' if kind == 'absence' else 'stderr:[]uint8{}, exitCode:0'
    if output.count(expected) != 3:
        raise SystemExit('mutant did not execute valid expected counterfactual: ' + str(path))
    print(kind + ': three independent pins caught the counterfactual')
