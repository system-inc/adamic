"""Recount committed evidence independently and kill one mutant per report check."""
import collections
import copy
import gzip
import hashlib
import json
import pathlib
import re
import sys

root = pathlib.Path(__file__).resolve().parent
adapted = pathlib.Path(sys.argv[1]).resolve()
report = json.loads((root / 'REPORT.json').read_text())

def check(value):
    sources = value['source']['files']
    expected = {str(p.relative_to(adapted)): hashlib.sha256(p.read_bytes()).hexdigest() for p in (adapted / 'src/compiler').rglob('*.ts')}
    assert {row['file']: row['sha256'] for row in sources} == expected, 'source manifest'
    for run in value['runs']:
        assert set(run['per_file']) == set(expected), 'file coverage'
        assert run['lowering_counts'] is None and run['feature_lowering_delta'] is None, 'unknown lowering result'
        if run['status'] != 'complete': continue
        with gzip.open(root / 'data' / (run['name'] + '.jsonl.gz'), 'rt') as stream:
            raw = [json.loads(line) for line in stream]
        assert raw[0]['status'] == 'checker', 'checker prerequisite'
        diagnostics = raw[0]['diagnostics']
        counts = collections.Counter(re.search(r'error (TS\d+):', d)[1] for d in diagnostics)
        assert run['checker_total'] == len(diagnostics) and run['per_reason'] == counts, 'reason recount'
        assert len(raw[1:]) == len(expected), 'raw file coverage'
        for row in raw[1:]:
            own = [d for d in diagnostics if d.startswith(row['file'] + ':')]
            assert row['checker_diagnostics'] == own, 'raw location attribution'
            assert run['per_file'][row['file']] == {'NotYet': None, 'Refused': None, 'status': 'blocked', 'checker': len(own)}, 'per-file recount'
    baseline = value['runs'][0]
    for run in value['runs'][1:]:
        if 'checker_delta' in run:
            assert run['checker_delta'] == run['checker_total'] - baseline['checker_total'], 'delta recount'

check(report)
mutants = [
    ('reason recount', lambda r: r['runs'][0].__setitem__('checker_total', r['runs'][0]['checker_total'] + 1)),
    ('source manifest', lambda r: r['source']['files'][0].__setitem__('sha256', '0' * 64)),
    ('file coverage', lambda r: r['runs'][0]['per_file'].pop(next(iter(r['runs'][0]['per_file'])))),
    ('unknown lowering result', lambda r: r['runs'][0].__setitem__('feature_lowering_delta', 0)),
    ('per-file recount', lambda r: r['runs'][0]['per_file'][next(iter(r['runs'][0]['per_file']))].__setitem__('checker', -1)),
    ('delta recount', lambda r: r['runs'][1].__setitem__('checker_delta', 1)),
]
for name, mutate in mutants:
    value = copy.deepcopy(report)
    mutate(value)
    try:
        check(value)
    except AssertionError as failure:
        assert str(failure) == name, (name, failure)
        print(name + ' mutant caught')
    else:
        raise AssertionError(name + ' mutant survived')
print('report evidence audit passed')
