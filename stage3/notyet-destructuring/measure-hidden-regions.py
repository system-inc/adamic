#!/usr/bin/env python3
"""Measure only the two supplied hidden regions from guarded census replays."""
import hashlib
import json
from pathlib import Path
import sys

source = Path(sys.argv[1])
evidence = Path(sys.argv[2])
baseline = json.loads((evidence / 'BASELINE.json').read_text())
results = []
for name, short in [('visitorPublic.ts', 'visitor'), ('parser.ts', 'parser')]:
    region = next(x for x in baseline['largest_regions'] if x['file'] == name)
    data = (source / name).read_bytes()
    metadata = baseline['files'][name]
    assert len(data) == metadata['bytes']
    assert hashlib.sha256(data).hexdigest() == metadata['sha256']
    assert len(region['causes']) == 1 and region['causes'][0][0] == 'Boundary'
    before = json.loads((evidence / (short + '-before.json')).read_text())
    after = json.loads((evidence / (short + '-after.json')).read_text())
    assert len(before['units']) == len(after['units']) == 1
    assert before['units'][0]['where'] == after['units'][0]['where']
    assert after['units'][0]['status'] == 'attempted'
    assert not after['units'][0]['checker_diagnostics']
    original = [f for f in before['findings'] if f['kind'] == 'Boundary']
    assert len(original) == 1
    assert (original[0]['start'], original[0]['end']) == (region['start'], region['end'])
    assert any(f['kind'] == 'NotYet' and f['reason'] == 'a computed field name' for f in before['findings'])
    assert not any(f['kind'] == 'NotYet' and f['reason'] == 'a computed field name' for f in after['findings'])
    # These variable-statement attempts have no skipped checker/dependency bodies.
    # Reject any new record category instead of silently omitting its spans.
    assert all(f['kind'] in ('NotYet', 'Boundary') for f in after['findings'])
    spans = sorted((f['start'], f['end']) for f in after['findings'] if f['kind'] == 'Boundary')
    merged = []
    mask = bytearray(region['bytes'])
    for left, right in spans:
        assert region['start'] <= left < right <= region['end']
        mask[left-region['start']:right-region['start']] = bytes([1]) * (right-left)
        if merged and left <= merged[-1][1]:
            merged[-1][1] = max(merged[-1][1], right)
        else:
            merged.append([left, right])
    remaining = sum(right-left for left, right in merged)
    assert remaining == sum(mask)  # Independent per-byte union oracle.
    stops = [f for f in after['findings'] if f['kind'] == 'NotYet']
    results.append(dict(file=name, start=region['start'], end=region['end'],
        source_sha256=metadata['sha256'], before_hidden_bytes=region['bytes'],
        after_hidden_bytes=remaining, revealed_bytes=region['bytes']-remaining,
        remaining_boundaries=len(spans), next_stop=stops[0] if stops else None,
        remaining_ranges=merged))
print(json.dumps(dict(measurement='regional census-replay ledger on checker-rejected entry program; not a fresh global census',
    baseline='codex/stage3-hidden-source 388096e6',
    head_reproduction='controlled replay with const-enum field-name support disabled',
    regions=results), indent=2))
