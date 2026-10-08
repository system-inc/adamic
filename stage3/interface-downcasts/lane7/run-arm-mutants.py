#!/usr/bin/env python3
"""Each selected-arm mutant must finish valid code and break all refusal pins."""
import os
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parents[3]
logs = Path(sys.argv[1]).resolve()
logs.mkdir(parents=True, exist_ok=True)
for kind in ('skip', 'shape', 'nested', 'tag'):
    fixture = 'leading-access-unknown-tag' if kind == 'tag' else 'leading-access-wrong'
    path = logs / (kind + '.log')
    command = ['go', 'test', './internal/oracle', '-run',
               '^TestCheckedViewIntersectionSelectedArms/' + fixture + '$', '-count=1', '-v']
    with path.open('w') as log:
        result = subprocess.run(command, cwd=root,
            env=dict(os.environ, ADAMIC_INTERSECTION_ARM_MUTANT=kind),
            stdout=log, stderr=subprocess.STDOUT)
    output = path.read_text()
    if (result.returncode == 0 or output.count('got oracle.run') != 3
            or output.count('stderr:[]uint8{}, exitCode:0') != 3
            or output.count('stdout:[]uint8{0x74, 0x72, 0x75, 0x65, 0xa}') != 3):
        raise SystemExit('invalid mutant result: ' + str(path))
    print(kind + ': three refusal pins failed; valid execution printed true')
