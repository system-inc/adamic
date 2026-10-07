"""Temporarily ignore field writes, require the method fixture's ASan failure, restore."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[2]
source = root / 'internal/native/borrow.go'
original = source.read_text()
before = 'if write, ok := statement.(ir.SetProperty); ok && names[write.Name] {'
after = 'if write, ok := statement.(ir.SetProperty); ok && names[write.Name] && false {'
assert original.count(before) == 1
log = Path('/tmp/borrow-coverage-mutant.log')
try:
    source.write_text(original.replace(before, after))
    with log.open('w') as output:
        result = subprocess.run([
            'go', 'test', './internal/oracle', '-run',
            'TestNativeAgreesWithNode/internal/oracle/testdata/borrow_chain_coverage_method.a$',
            '-count=1', '-timeout', '30m',
        ], cwd=root, env={**os.environ, 'ADAMIC_GATE_UNCACHED': '1'},
            stdout=output, stderr=subprocess.STDOUT)
    assert result.returncode != 0
    assert 'heap-use-after-free' in log.read_text()
    print('Mutant caught by ASan; compiler restored.')
finally:
    source.write_text(original)
