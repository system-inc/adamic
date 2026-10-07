#!/usr/bin/env python3
"""Refresh only stage0 records, with current Node and native observations checked."""
import json
from pathlib import Path
import re
import subprocess
import tempfile

root = Path(__file__).resolve().parents[1]
paths = sorted(root.glob('stage3/fixtures/*/status.json'))
before = {p: p.read_bytes() for p in paths}
log = Path('/tmp/adamic-reland-status-update.log')
source_path = root / 'stage3/fixtures/fixtures_test.go'
source = source_path.read_text()
old = 'if *update && recordedNodeAgrees && nativeAgrees {'
new = 'if *update && recordedNodeAgrees && (nativeAgrees || (actual.Outcome != "Compiles" && entry.Stage0.Outcome != "Compiles")) {'
if source.count(old) != 1:
    raise RuntimeError('review the stage3 updater before using this audit tool')
with tempfile.TemporaryDirectory(prefix='adamic-reland-status-') as directory:
    scratch = Path(directory)
    variant = scratch / 'fixtures_test.go'
    variant.write_text(source.replace(old, new))
    overlay = scratch / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': {str(source_path): str(variant)}}))
    with log.open('w') as output:
        result = subprocess.run(['go', 'test', '-overlay', str(overlay), '-count=1', '-timeout', '30m',
                                 './stage3/fixtures', '-run', '^TestFixtures$', '-args', '-update'],
                                cwd=root, stdout=output, stderr=subprocess.STDOUT)

def outside_stage0(raw):
    text = raw.decode()
    decoder = json.JSONDecoder()
    spans = []
    for match in re.finditer(r'"stage0"\s*:\s*', text):
        _, end = decoder.raw_decode(text, match.end())
        spans.append((match.end(), end))
    for start, end in reversed(spans):
        text = text[:start] + '<stage0>' + text[end:]
    return text

changes = []
regressions = []
for path, old in before.items():
    current = path.read_bytes()
    if outside_stage0(old) != outside_stage0(current):
        raise RuntimeError(f'{path}: bytes outside stage0 changed')
    old_rows, new_rows = json.loads(old), json.loads(current)
    for a, b in zip(old_rows, new_rows, strict=True):
        if a['node'] != b['node']:
            raise RuntimeError(f'{path}: Node observations changed')
        if a['stage0'] != b['stage0']:
            change = {'path': str(path.relative_to(root)), 'file': a['file'],
                      'before': a['stage0'], 'after': b['stage0']}
            changes.append(change)
            if a['stage0']['outcome'] == 'Compiles' and b['stage0']['outcome'] in ('NotYet', 'Refused'):
                regressions.append(change)
print('Compiles to Refused/NotYet: ' + str(len(regressions)), flush=True)
print(json.dumps(regressions, indent=2), flush=True)
print('Stage0 changes: ' + str(len(changes)) + '; every byte outside stage0 preserved', flush=True)
Path('/tmp/adamic-reland-status-changes.json').write_text(json.dumps(changes, indent=2) + '\n')
print('stage3 update exit=' + str(result.returncode), flush=True)
if regressions or result.returncode:
    raise SystemExit(1)
