#!/usr/bin/env python3
"""Refresh existing NotYet diagnostics after recorded Node agrees, using a test overlay.

The normal updater already checks new native successes. This overlay additionally
records unsupported forms, and cannot turn a compiling fixture into a gap.
"""
from pathlib import Path
import json
import subprocess
import tempfile

root = Path(__file__).resolve().parents[2]
source = root / 'stage3/fixtures/fixtures_test.go'
original = source.read_text()
before = 'if *update && recordedNodeAgrees && nativeAgrees {'
after = 'if *update && recordedNodeAgrees && (nativeAgrees || (actual.Outcome == "NotYet" && entry.Stage0.Outcome != "Compiles" && entry.Stage0.Outcome != "CheckedStop")) {'
assert original.count(before) == 1
with tempfile.TemporaryDirectory(prefix='records-status-') as temporary:
    temporary = Path(temporary)
    updated = temporary / 'fixtures_test.go'
    updated.write_text(original.replace(before, after))
    overlay = temporary / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': {str(source): str(updated)}}))
    result = subprocess.run(['go', 'test', '-overlay', str(overlay), './stage3/fixtures', '-run', '^TestFixtures/(records|objects|taste)$', '-count=1', '-timeout', '30m', '-v', '-args', '-update'], cwd=root)
    raise SystemExit(result.returncode)
