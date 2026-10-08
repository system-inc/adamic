#!/usr/bin/env python3
"""Require valid mutant execution to break each independent three-mode refusal pin."""
import os
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parents[3]
logs = Path(sys.argv[1]).resolve()
logs.mkdir(parents=True, exist_ok=True)
for kind, fixture in [
    ('skip', 'recursive-read'), ('shape', 'recursive-read'),
    ('nested', 'recursive-read'), ('canonical', 'recursive-read'),
    ('optional', 'recursive-optional-root'),
    ('literal-mode', 'recursive-literal-wrong'),
    ('literal-code', 'recursive-number-literal'),
    ('literal-enabled', 'recursive-boolean-literal'),
]:
    path = logs / (kind + '.log')
    command = ['go', 'test', './internal/oracle', '-run',
               '^TestCheckedViewIntersectionRecursiveDemand/' + fixture + '$', '-count=1', '-v']
    env = dict(os.environ, ADAMIC_INTERSECTION_RECURSIVE_MUTANT=kind)
    with path.open('w') as log:
        result = subprocess.run(command, cwd=root, env=env, stdout=log, stderr=subprocess.STDOUT)
    output = path.read_text()
    if (result.returncode == 0 or output.count('got oracle.run') != 3
            or output.count('stderr:[]uint8{}, exitCode:0') != 3
            or output.count('stdout:[]uint8{0x74, 0x72, 0x75, 0x65, 0xa}') != 3):
        raise SystemExit('invalid mutant result: ' + str(path))
    print(kind + ': three refusal pins failed; valid execution printed true')
