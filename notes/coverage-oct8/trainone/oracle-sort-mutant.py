#!/usr/bin/env python3
"""Check whether the existing source oracle covers the two-index native gap."""
from pathlib import Path
import os
import subprocess
root = Path(__file__).resolve().parents[3]
log = Path(__file__).resolve().parent / 'evidence' / 'sort-existing-oracle.log'
path = root / 'internal/native/runtime/record.c'
original = path.read_text()
assert original.count('if (count > 1) {') == 1
try:
 path.write_text(original.replace('if (count > 1) {', 'if (count > 2) {', 1))
 with log.open('wb') as stream:
  result = subprocess.run(['go', 'test', './internal/oracle', '-run', 'TestNativeAgreesWithNode/internal/oracle/testdata/entries_record_alias', '-count=1', '-timeout', '30m', '-v'], cwd=root, stdout=stream, stderr=stream, env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), timeout=1800)
 print('existing source oracle sort mutant exit=' + str(result.returncode))
finally:
 path.write_text(original)
