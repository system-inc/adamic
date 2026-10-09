#!/usr/bin/env python3
"""Check the packaged measurement. Mutants alter evidence in memory only."""
import hashlib
import json
import pathlib
import sys

root = pathlib.Path(__file__).resolve().parent
mode = sys.argv[1] if len(sys.argv) > 1 else ''
summary = json.loads((root / 'summary.json').read_text())
merges = json.loads((root / 'merges.json').read_text())
assert merges[0]['skipped'] is False and merges[0]['after'] == summary['retained_scratch']
assert merges[1]['skipped'] is True and merges[1]['after'] == summary['retained_scratch']
assert merges[1]['conflicts'] == ['internal/oracle/counts.md', 'stage3/fixtures/nested-functions/09_checker_constituent_recursion.a']
builds = json.loads((root / 'builds/report.json').read_text())['attempts']
assert [row['split'] for row in builds] == [0, 1]
assert all(row['exit'] == 1 and row['binary'] is None for row in builds)
assert (root / 'builds/0.stderr').read_bytes() == (root / 'builds/1.stderr').read_bytes()
assert b"captureStackTrace" in (root / 'builds/0.stderr').read_bytes()
node = json.loads((root / 'node-report.json').read_text())['node']
if mode == '--wrong-hash':
    node['sha256'] = '0' * 64
assert node['bytes'] == 36429231 and node['sha256'] == '686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615', 'fixed Node reference'
stops = json.loads((root / 'ordered-stops.json').read_text())['rows']
if mode == '--drop-stop':
    stops.pop()
assert len(stops) == summary['ordered_stops'] == 9, 'ordered stop coverage'
assert [row['order'] for row in stops] == list(range(1, 10))
for row in stops:
    assert row['owner'] and row['minimal'] and row['status_vs_previous'] == 'still there'
    assert row['node']['exit'] == 0 and row['node']['stderr'] == ''
    assert (root.parents[1] / row['minimal']).exists()
    assert row['behind_discovery_placeholders'] == (row['order'] > 1)
probes = json.loads((root / 'probes/report.json').read_text())['probes']
assert len(probes) == 21 and sum(row['pass'] for row in probes) == 6
assert sum(row['status_vs_previous'] == 'gone' for row in probes) == 1
assert sum(row['status_vs_previous'] == 'new' for row in probes) == 3
for row in probes:
    assert row['node']['exit'] == 0 and row['node']['stderr'] == ''
    if row['pass']:
        assert row['node'] == row['native'] and row['byte_mutant_exit'] == 1
        assert (root / ('probes/' + row['file'] + '.mutant-comparison.log')).read_bytes()
print('PASS: merge skips, two checker-red builds, fixed Node hash, nine ordered stops, 21 Node-held probes, six native byte mutants')
