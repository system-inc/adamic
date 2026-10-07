#!/usr/bin/env python3
"""Prove value and evaluation-order comparisons can fail; restore each mutation."""
from pathlib import Path
import os
import subprocess

repository = Path(__file__).resolve().parents[4]
source = repository / 'internal/lower/increment_values.go'
original = source.read_text()
logs = Path('/tmp/scanner-increments-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('postfix-return-new', 'if !postfix {', 'if true {'),
    ('field-target-twice', 'l.libraryArrayBuilder([]ir.Expression{receiver})',
     'l.libraryArrayBuilder([]ir.Expression{receiver, receiver})'),
    ('array-target-twice', 'l.libraryArrayBuilder([]ir.Expression{array, index})',
     'l.libraryArrayBuilder([]ir.Expression{array, index, array})'),
]
environment = dict(os.environ, ADAMIC_GATE_UNCACHED='1')
for name, before, after in mutants:
    assert original.count(before) == 1, name
    try:
        source.write_text(original.replace(before, after))
        with (logs / (name + '.log')).open('w') as output:
            result = subprocess.run([
                'go', 'test', './internal/oracle', '-count=1', '-timeout', '10m', '-v',
                '-run', 'TestNativeAgreesWithNode/internal/oracle/testdata/increment_values',
            ], cwd=repository, env=environment, stdout=output, stderr=subprocess.STDOUT)
        text = (logs / (name + '.log')).read_text()
        assert result.returncode != 0, name + ' survived'
        assert 'stdout differs' in text, name + ' did not fail the output comparison'
        assert 'JavaScript backend: stdout differs' in text, name + ' did not fail backend Node'
        assert 'error:' not in text and 'Lower:' not in text, name + ' was stopped before execution'
        print(name + ': caught by source Node versus native and backend stdout; exit ' + str(result.returncode), flush=True)
    finally:
        source.write_text(original)
