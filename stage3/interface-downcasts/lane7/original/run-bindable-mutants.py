#!/usr/bin/env python3
"""Counterfactuals for the original Bindable pairs, each caught in three modes."""
import os
from pathlib import Path
import subprocess
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: run-bindable-mutants.py <declarations> <logs>')
root = Path(__file__).resolve().parents[4]
declarations, logs = map(lambda value: Path(value).resolve(), sys.argv[1:])
logs.mkdir(parents=True, exist_ok=True)
# Each fixture's wrong value is reachable only by the read the mutant weakens.
for kind, fixture in [('left-skip', 'bindable-static-left-wrong'), ('left-shape', 'bindable-static-left-kind'),
                      ('left-nested', 'bindable-static-left-wrong'),
                      ('expression-skip', 'bindable-static-expression-wrong'), ('expression-shape', 'bindable-static-expression-this'),
                      ('expression-nested', 'bindable-static-expression-wrong'),
                      ('access-skip', 'bindable-access-left-symbol'), ('access-shape', 'bindable-access-left-kind'),
                      ('access-nested', 'bindable-access-left-symbol'),
                      ('expression-skip', 'bindable-access-expression-wrong'), ('expression-shape', 'bindable-access-expression-this'),
                      ('expression-nested', 'bindable-access-expression-wrong')]:
    path = logs / (kind + '-' + fixture + '.log')
    env = dict(os.environ, ADAMIC_INTERSECTION_ORIGINAL_DECLS=str(declarations),
               ADAMIC_INTERSECTION_ORIGINAL_MUTANT=kind)
    with path.open('w') as log:
        result = subprocess.run(['go', 'test', './internal/oracle', '-run',
            '^TestCheckedViewIntersectionOriginalBindable/' + fixture + '$', '-count=1', '-v'],
            cwd=root, env=env, stdout=log, stderr=subprocess.STDOUT)
    output = path.read_text()
    if result.returncode == 0 or output.count('got oracle.run') != 3:
        raise SystemExit('mutant escaped refusal pins: ' + str(path))
    if output.count('stdout:[]uint8{0x74, 0x72, 0x75, 0x65, 0xa}, stderr:[]uint8{}, exitCode:0') != 3:
        raise SystemExit('mutant did not execute the valid counterfactual: ' + str(path))
    print(kind + ' on ' + fixture + ': three independent pins caught the counterfactual')
