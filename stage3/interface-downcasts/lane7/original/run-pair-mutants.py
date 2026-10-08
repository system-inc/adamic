#!/usr/bin/env python3
"""Counterfactuals for the original intersection pairs, each caught in three modes."""
import os
from pathlib import Path
import subprocess
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: run-pair-mutants.py <declarations> <logs>')
root = Path(__file__).resolve().parents[4]
declarations, logs = map(lambda value: Path(value).resolve(), sys.argv[1:])
logs.mkdir(parents=True, exist_ok=True)
# Each fixture's wrong value is reachable only by the read the mutant weakens.
bindable = 'TestCheckedViewIntersectionOriginalBindable'
emit_class = 'TestCheckedViewIntersectionOriginalNodes'
true = 'stdout:[]uint8{0x74, 0x72, 0x75, 0x65, 0xa}'
zero = 'stdout:[]uint8{0x30, 0xa}'
runs = [(bindable, 'left-skip', 'bindable-static-left-wrong', true), (bindable, 'left-shape', 'bindable-static-left-kind', true),
        (bindable, 'left-nested', 'bindable-static-left-wrong', true),
        (bindable, 'expression-skip', 'bindable-static-expression-wrong', true), (bindable, 'expression-shape', 'bindable-static-expression-this', true),
        (bindable, 'expression-nested', 'bindable-static-expression-wrong', true),
        (bindable, 'access-skip', 'bindable-access-left-symbol', true), (bindable, 'access-shape', 'bindable-access-left-kind', true),
        (bindable, 'access-nested', 'bindable-access-left-symbol', true),
        (bindable, 'expression-skip', 'bindable-access-expression-wrong', true), (bindable, 'expression-shape', 'bindable-access-expression-this', true),
        (bindable, 'expression-nested', 'bindable-access-expression-wrong', true),
        (emit_class, 'emit-skip', 'emit-original-range-wrong', zero), (emit_class, 'emit-shape', 'emit-original-wrong', zero),
        (emit_class, 'emit-nested', 'emit-original-range-wrong', zero), (emit_class, 'emit-presence', 'emit-original-missing', zero),
        (emit_class, 'class-skip', 'class-augments-symbol', true), (emit_class, 'class-shape', 'class-augments-kind', true),
        (emit_class, 'class-nested', 'class-augments-symbol', true),
        (emit_class, 'class-skip', 'class-implements-symbol', true), (emit_class, 'class-shape', 'class-implements-kind', true),
        (emit_class, 'class-nested', 'class-implements-symbol', true),
        (emit_class, 'jsdoc-skip', 'jsdoc-parent-parameters', true), (emit_class, 'jsdoc-shape', 'jsdoc-parent-kind', true),
        (emit_class, 'jsdoc-nested', 'jsdoc-parent-parameters', true),
        (emit_class, 'root-skip', 'jsdoc-root-symbol', true), (emit_class, 'root-shape', 'jsdoc-root-kind', true),
        (emit_class, 'root-nested', 'jsdoc-root-symbol', true)]
selected = os.environ.get('ADAMIC_MUTANT_TEST')
for test, kind, fixture, stdout in runs:
    if selected and selected != test:
        continue
    path = logs / (kind + '-' + fixture + '.log')
    env = dict(os.environ, ADAMIC_INTERSECTION_ORIGINAL_DECLS=str(declarations),
               ADAMIC_INTERSECTION_ORIGINAL_MUTANT=kind)
    with path.open('w') as log:
        result = subprocess.run(['go', 'test', './internal/oracle', '-run',
            '^' + test + '/' + fixture + '$', '-count=1', '-v'],
            cwd=root, env=env, stdout=log, stderr=subprocess.STDOUT)
    output = path.read_text()
    if result.returncode == 0 or output.count('got oracle.run') != 3:
        raise SystemExit('mutant escaped refusal pins: ' + str(path))
    if output.count(stdout + ', stderr:[]uint8{}, exitCode:0') != 3:
        raise SystemExit('mutant did not execute the valid counterfactual: ' + str(path))
    print(kind + ' on ' + fixture + ': three independent pins caught the counterfactual')
