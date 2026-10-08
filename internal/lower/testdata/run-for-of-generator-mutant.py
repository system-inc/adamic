#!/usr/bin/env python3
"""Pin the generator refusal separately from the latent iterator-body stop."""
from pathlib import Path
import os
import subprocess

repository = Path(__file__).resolve().parents[3]
source = repository / 'internal/lower/refusals.go'
original = source.read_text()
assert original.count('if generator {') == 1
try:
    source.write_text(original.replace('if generator {', 'if generator && false {'))
    with Path('/tmp/for-of-generator-mutant.log').open('w') as output:
        result = subprocess.run(['go', 'test', './internal/oracle', '-run', 'TestForOfGeneratorRootsRemainRefused', '-count=1', '-timeout', '10m'], cwd=repository, env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=output, stderr=subprocess.STDOUT)
    observed = Path('/tmp/for-of-generator-mutant.log').read_text()
    assert result.returncode != 0 and observed.count('want generator ownership refusal and no IR') == 3, observed
    assert '[build failed]' not in observed and '[-Werror' not in observed, observed
    print('omit-generator-refusal: caught by all three generator-root fixtures, exit 1', flush=True)
finally:
    source.write_text(original)
