#!/usr/bin/env python3
"""Recount captured train observations; successful train builds remain pending."""
import argparse
import hashlib
import json
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument('evidence', type=Path)
parser.add_argument('--mutant', choices=['first-stop', 'exit', 'status', 'node-count'])
args = parser.parse_args()
root = args.evidence
summary = json.loads((root / 'summary.json').read_text())
if args.mutant == 'node-count':
    summary['node']['binder']['projects'] += 1
request = json.loads((root / 'projects.json').read_text())
assert len(request['projects']) == 301, 'request count'
for piece in ['binder', 'checker', 'emitter']:
    row = summary['node'][piece]
    assert row['projects'] == len(request['projects']), 'node population recount'
    comparison = json.loads((root / 'node' / piece / 'report.json').read_text())
    assert comparison['success'] and not comparison['timeout'], 'node differential'
    assert (root / 'node' / piece / 'actual.exit').read_text() == '0\n', 'node exit'
    assert (root / 'node' / piece / 'actual.stderr').read_bytes() == b'', 'node stderr'
    assert row['sha256'] == hashlib.sha256((root / 'node' / piece / 'golden.stdout').read_bytes()).hexdigest(), 'golden hash'
for label in ['train', 'main']:
    report = json.loads((root / 'native' / label / 'report.json').read_text())
    assert len(report['results']) == 6, 'native profile population'
    for index, row in enumerate(report['results']):
        if label == 'train' and index == 0:
            if args.mutant == 'first-stop': row['first_stop'] = 'mutant'
            if args.mutant == 'exit': row['exit'] = 0
            if args.mutant == 'status': row['status'] = 'pass'
        directory = root / 'native' / label / (row['piece'] + '-' + row['profile'])
        code = int((directory / 'exit').read_text())
        first = next((line for line in (directory / 'stderr').read_text().split('\n') if line.strip()), None)
        assert row['exit'] == code, 'native exit recount'
        assert row['first_stop'] == first, 'native first stop recount'
        assert row['status'] == ('pending' if code == 0 else 'blocked'), 'native pending status'
        assert not row['native_differential_run'], 'native execution not observed'
assert summary['native_pass'] is False, 'native parity not observed'
print('301 projects per Node differential; all native exits/stops/statuses match captured streams; no native pass claimed')
