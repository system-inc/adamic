#!/usr/bin/env python3
"""Run independent semantic overlays without changing the checked-in compiler."""
import json
from pathlib import Path
import subprocess
import sys
import tempfile

repository = Path(__file__).resolve().parents[3]
source = repository / 'stage3/scouts/step12/compare/main.go'
text = source.read_text()
out = Path(sys.argv[1]).resolve()
out.mkdir()  # Refuse to replace prior evidence.
for name, condition in [('ignored-byte', 'false'),
                        ('prefix-only', 'l != r && l != -1 && r != -1')]:
    with tempfile.TemporaryDirectory(prefix='step12-mutant-') as directory:
        scratch = Path(directory)
        mutant = scratch / 'main.go'
        if text.count('l != r') != 1:
            raise SystemExit('mutation anchor changed')
        mutant.write_text(text.replace('l != r', condition, 1))
        overlay = scratch / 'overlay.json'
        overlay.write_text(json.dumps({'Replace': {str(source): str(mutant)}}))
        log = out / (name + '.log')
        with log.open('w') as output:
            result = subprocess.run(
                ['go', 'test', '-overlay', str(overlay),
                 './stage3/scouts/step12/compare', '-run', 'TestCompare', '-count=1'],
                cwd=repository, stdout=output, stderr=subprocess.STDOUT)
        observation = log.read_text()
        if result.returncode != 1 or 'equal:' not in observation or '[build failed]' in observation:
            raise SystemExit(f'{name} was not caught by the equality assertion')
        print(f'{name}: caught by TestCompare equality assertion, exit 1')
