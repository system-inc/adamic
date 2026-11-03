#!/usr/bin/env python3
"""Restore a second measurement and require the invocation-count check to fail."""
import os
from pathlib import Path
import subprocess
import sys
import tempfile

root = Path(__file__).resolve().parents[2]
log_path = Path(sys.argv[1]).resolve()
source = (root / 'stage3/meter/twice-daily.sh').read_text()
changes = [
    ('for label in main; do', 'for label in main area; do'),
    ('cp -a "$run/main" "$run/area"', ':'),
    ('"$scratch/main-adapted" "$scratch/main-adapted"',
     '"$scratch/main-adapted" "$scratch/area-adapted"'),
]
for old, new in changes:
    assert source.count(old) == 1
    source = source.replace(old, new)
with tempfile.TemporaryDirectory(prefix='meter-measure-once-mutant-') as temporary:
    script = Path(temporary) / 'twice-daily.sh'
    script.write_text(source)
    with log_path.open('w') as log:
        result = subprocess.run(
            [sys.executable, '-m', 'unittest', 'discover', '-s', 'stage3/meter',
             '-p', 'compiler_test.py', '-k', 'test_per_ref_builds_and_runs_each_pinned_compiler'],
            cwd=root, env=dict(os.environ, METER_SCRIPT_UNDER_TEST=str(script)),
            stdout=log, stderr=log, timeout=300)
    text = log_path.read_text()
    if result.returncode == 0 or '8 != 4 : main must be measured exactly once' not in text:
        raise AssertionError('second measurement survived or failed for another reason: ' + text)
print('Caught second measurement: 8 != 4, main must be measured exactly once')
