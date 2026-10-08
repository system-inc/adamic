#!/usr/bin/env python3
"""Prove the void-of-async guard through an overlay, without editing production."""
import json
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parents[3]
work = Path(sys.argv[1]).resolve()
work.mkdir(parents=True, exist_ok=True)
source = root / 'internal/lower/expression.go'
text = source.read_text()
seam = 'if operand.Type() == ir.Promise {'
assert text.count(seam) == 1
changed = work / 'expression.go'
changed.write_text(text.replace(seam, 'if false && operand.Type() == ir.Promise {', 1))
overlay = work / 'overlay.json'
overlay.write_text(json.dumps({'Replace': {str(source): str(changed)}}))
for name, flags, expected in [('control', [], 0), ('mutant', ['-overlay', str(overlay)], 1)]:
    log = work / f'{name}.log'
    with log.open('w') as output:
        result = subprocess.run(
            ['go', 'run', *flags, './review/runtime-area-on-next/check',
             str(work / name), '--refusal'], cwd=root,
            stdout=output, stderr=subprocess.STDOUT, check=False)
    observed = log.read_text()
    assert result.returncode == expected, observed
    assert ('PASS void refusal:' if name == 'control' else 'FAIL void refusal: <nil>') in observed, observed
    print(f'{name}: exit {result.returncode}, {log}')
print('CAUGHT void guard removal: lowering accepted the refused task')
