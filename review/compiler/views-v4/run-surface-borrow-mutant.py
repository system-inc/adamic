"""Prove that a checking adapter cannot be borrowed as its source field."""
import difflib
import os
from pathlib import Path
import re
import subprocess

root = Path(__file__).resolve().parents[3]
source = root / 'internal/native/borrow.go'
evidence = root / 'review/compiler/views-v4'
original = source.read_bytes()
before = b' || expression.ViewEscape && !expression.ViewEscapeAdamic'
if original.count(before) != 1:
    raise SystemExit('borrow mutation site moved')
mutated = original.replace(before, b'', 1)
(evidence / 'surface-borrow-mutant.patch').write_text(''.join(difflib.unified_diff(original.decode().splitlines(True), mutated.decode().splitlines(True), fromfile='a/internal/native/borrow.go', tofile='b/internal/native/borrow.go')))
try:
    source.write_bytes(mutated)
    with (evidence / 'surface-borrow-mutant.log').open('w') as log:
        result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^TestV4EscapeAdapterName$', '-count=1', '-timeout', '90s'], cwd=root, env={**os.environ, 'ADAMIC_GATE_UNCACHED': '1'}, stdout=log, stderr=subprocess.STDOUT, timeout=90)
    output = (evidence / 'surface-borrow-mutant.log').read_text()
    decoded = '\n'.join(bytes(int(value,16) for value in re.findall(r'0x([0-9a-f]+)', match.group(1))).decode() for match in re.finditer(r'stderr:\[\]uint8\{([^}]*)\}', output))
    (evidence / 'surface-borrow-mutant-asan.log').write_text(decoded)
    if result.returncode != 1 or 'heap-use-after-free' not in decoded or 'TestV4EscapeAdapterName' not in output:
        raise SystemExit('borrow mutant was not caught by ASan')
    print('borrow-field mutant caught by ASan heap-use-after-free in TestV4EscapeAdapterName')
finally:
    source.write_bytes(original)
    if source.read_bytes() != original:
        raise SystemExit('failed to restore borrow source')
