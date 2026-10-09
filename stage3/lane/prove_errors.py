#!/usr/bin/env python3
"""Verify real oracle timeout/baseline evidence and kill a dropped-text mutant."""
import argparse
import json
from unittest.mock import patch
from pathlib import Path
import check

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('timeout', type=Path)
parser.add_argument('baseline', type=Path)
parser.add_argument('normal', type=Path)
args = parser.parse_args()
for directory, cause in [(args.timeout, 'timeout'), (args.baseline, 'baseline-content')]:
    oracle = json.loads((directory / 'report.json').read_text())
    lane = json.loads((directory / 'lane/report.json').read_text())
    assert oracle['counts']['failing'] == 1 and oracle['status'] == lane['status'] == 'fail'
    row = lane['failures'][0]
    assert row['source'] == 'mocha' and row['cause'] == cause and row['message']
    assert lane['cause_counts'][cause] == 1
    if cause == 'timeout':
        assert row['kind'] == 'hook' and row['timeout_ms'] == 1 and row['elapsed_ms'] >= 1
        assert 'Timeout of 1ms exceeded' in row['message']
    else:
        assert {b['path'] for b in row['baselines']} == {'Protected1.types', 'Protected1.symbols'}
        assert all('stage3ForcedMismatch' in b['preview'] for b in row['baselines'])
        assert all(len(b['preview'].splitlines()) <= 40 for b in row['baselines'])
    assert json.loads((directory / 'lane/verdict.json').read_text())['failures'] == lane['failures']
    print(cause, 'proved', flush=True)

normal = json.loads((args.normal / 'report.json').read_text())
assert normal['status'] == 'pass' and normal['counts'] == dict(passing=106366, failing=1, pending=0)
assert normal['cause_counts']['baseline-content'] == 1
assert normal['failures'][0]['source'] == 'mocha' and normal['failures'][0]['message']
print('normal full lane remains PASS 106366/1/0', flush=True)

original = check.read_failures
def drop(*args, **kwargs):
    rows = original(*args, **kwargs)
    for row in rows:
        row['message'] = ''
    return rows
with patch.object(check, 'read_failures', side_effect=drop):
    mutant = check.check_results(args.timeout / 'lane')
assert mutant['status'] == 'fail' and mutant['cause_counts']['timeout'] == 1
try:
    assert 'Timeout of 1ms exceeded' in mutant['failures'][0]['message'], 'dropped full timeout error text'
except AssertionError as caught:
    print('mutant caught only by error-text assertion:', caught, flush=True)
else:
    raise AssertionError('dropped-text mutant escaped')
