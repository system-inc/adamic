"""Re-select and validate the committed survey against a fresh diagnostic export.
Usage: python3 docs/tsc-strictness/audit.py DIAGNOSTICS_JSON TYPESCRIPT_CHECKOUT
The checkout must be the pinned v6.0.3 source. No repository writes.
"""
import collections
import json
import math
from pathlib import Path
import re
import sys

here = Path(__file__).resolve().parent
export = json.loads(Path(sys.argv[1]).read_text())
root = Path(sys.argv[2]).resolve()
rows = json.loads((here / 'sample.json').read_text())
meter = json.loads((here / 'meter.json').read_text())
expected = {2412: 617, 2375: 76, 2379: 31, 18048: 354, 2532: 224}
pools = collections.defaultdict(list)
for d in export['diagnostics']:
    m = re.fullmatch(r'(.+?):(\d+):(\d+): error TS(\d+): ([\s\S]*)', d)
    if not m or int(m[4]) not in expected:
        continue
    file, line, column, code, message = m.groups()
    pools[int(code)].append((str(Path(file).resolve().relative_to(root)), int(line), int(column), int(code), message))
assert export['roots'] == meter['files_examined'] == 77
assert {code: len(pools[code]) for code in expected} == expected
assert sum(r['diagnostics_removed'] for r in export['rewrites']) == 3718
for code, count in expected.items():
    assert next(r['count'] for r in meter['reasons'] if r['reason'] == f'TypeScript TS{code}') == count
assert len(rows) == 200
document = (here.parent / 'tsc-strictness.md').read_text()
for pattern, count in collections.Counter(r['pattern'] for r in rows).items():
    match = re.search(r'\| ' + re.escape(pattern) + r':[^|]+\| (\d+) \|', document)
    assert match and int(match[1]) == count, ('document pattern count', pattern, count)
for group, codes in [('E', [2412, 2375, 2379]), ('U', [18048, 2532])]:
    total = sum(expected[c] for c in codes)
    quotas = {c: math.floor(100 * expected[c] / total) for c in codes}
    for c in sorted(codes, key=lambda c: -(100 * expected[c] / total - quotas[c]))[:100 - sum(quotas.values())]:
        quotas[c] += 1
    selected = []
    for c in codes:
        pool = sorted(pools[c], key=lambda r: r[:3])
        for k in range(quotas[c]):
            rank = math.floor((k + .5) * len(pool) / quotas[c])
            selected.append((pool[rank], rank))
    selected.sort(key=lambda item: item[0][:4])
    actual = [r for r in rows if r['id'].startswith(group)]
    assert len(actual) == 100
    for n, (r, (original, rank)) in enumerate(zip(actual, selected), 1):
        assert r['id'] == f'{group}{n:03}'
        assert tuple(r[key] for key in ['file', 'line', 'column', 'code', 'message']) == original
        assert r['stratum_rank'] == rank and r['stratum_population'] == expected[r['code']]
        assert r['source'] == (root / r['file']).read_text().splitlines()[r['line'] - 1]
        assert r['pattern'] and r['rationale'] and r['judgment'] in ['M', 'C', 'R']
    print(group, 'population', total, 'sample', len(actual), 'quotas', quotas,
          'patterns', dict(sorted(collections.Counter(r['pattern'] for r in actual).items())),
          'judgments', dict(sorted(collections.Counter(r['judgment'] for r in actual).items())))
print('PASS: all 200 IDs, source locations, source lines, diagnostic chains and quantile ranks verified')
