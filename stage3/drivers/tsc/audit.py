#!/usr/bin/env python3
"""Audit provenance and expectation/golden consistency without executing tsc."""
import hashlib
import json
import re
from pathlib import Path
import sys
from corpus import ROOT, baseline_output, materialize


def audit(root=ROOT):
    selection = json.loads((root / 'selection.json').read_text())
    cases = selection['cases']
    assert len(cases) == 300 and len({c['id'] for c in cases}) == 300, 'case population'
    assert sum(not c['codes'] for c in cases) == 60, 'clean population'
    for row in cases:
        folder = root / 'corpus' / row['id']
        raw = (folder / 'input.a').read_bytes()
        assert hashlib.sha256(raw).hexdigest() == row['source_sha256'], 'source hash: ' + row['id']
        assert materialize(raw)[1] == row['options'], 'header options: ' + row['id']
        if row['baseline']:
            raw_baseline = (folder / 'reference.errors.txt').read_bytes()
            assert hashlib.sha256(raw_baseline).hexdigest() == row['baseline_sha256'], 'baseline hash: ' + row['id']
            expected = baseline_output(raw_baseline)
        else:
            assert not (folder / 'reference.errors.txt').exists(), 'unexpected clean baseline'
            expected = b''
        assert row['codes'] == sorted(set(map(int, re.findall(rb'error TS(\d+):', expected)))), 'baseline codes: ' + row['id']
        assert (folder / 'expected.stdout').read_bytes() == expected, 'baseline summary: ' + row['id']
        assert (folder / 'expected.stderr').read_bytes() == b'', 'expected stderr'
        assert (folder / 'expected.exit').read_text() == ('2\n' if row['codes'] else '0\n'), 'expected exit'
        for suffix in ('stdout', 'stderr', 'exit'):
            assert (folder / ('golden.' + suffix)).read_bytes() == (folder / ('expected.' + suffix)).read_bytes(), 'golden: ' + row['id'] + '/' + suffix
    for suffix in ('stdout', 'stderr', 'exit'):
        assert (root / 'tiny' / ('golden.' + suffix)).read_bytes() == (root / 'tiny' / ('expected.' + suffix)).read_bytes(), 'tiny golden: ' + suffix
    print('PASS provenance, 300 cases, 60 clean, baseline summaries, headers, golden bytes and exits')


if __name__ == '__main__':
    audit(Path(sys.argv[1]) if len(sys.argv) > 1 else ROOT)
