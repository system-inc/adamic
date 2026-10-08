#!/usr/bin/env python3
"""Group every remaining pinned class-d relation with its dependency or rule."""
import collections
import json
import pathlib

root = pathlib.Path(__file__).resolve().parent
census = json.loads((root / 'census.json').read_text())
rows = [row for row in census['records'] if row['class'] == 'd']
assert len(rows) == 515
specifications = [
    ('Boxed-union fields and elements', [165,166,478,479,480,481,483,484,485,486,492,493,494,537], 'contract gap', 'views V2 and V5: tagged union slot contracts and safe representation conversion'),
    ('Tuple slots', [207,208,313,373], 'contract gap', 'views V3: positional slot and callable element contracts'),
    ('Recursive structural references', [151,152], 'contract gap', 'views V5/V6: recursive contracts preserving allocation identity'),
    ('Set element', [89], 'contract gap', 'views V2 and V5: nullable Set element contracts at insertion'),
    ('Constrained generic array', [53], 'contract gap', 'views V3: concrete member and element contracts for the constrained generic array union'),
    ('Method parameter variance', [429,430], 'independent rule', 'producer signatures accepting undefined; method-signature-style refuses inputs the implementation cannot take'),
    ('Erased any-to-generic callback relation', [413,613], 'independent rule', 'typed callback/table relations replacing value-level any; no-unsafe-assignment stays enforced'),
]
remaining = {row['id']: row for row in rows if row['status'] == 'stopped'}
groups = []
claimed = set()
for name, ids, category, dependency in specifications:
    records = [remaining[identifier] for identifier in ids]
    assert not claimed.intersection(ids)
    claimed.update(ids)
    groups.append({'reason': name, 'count': len(ids), 'category': category, 'waits_on': dependency, 'records': records})
assert claimed == set(remaining)
generic_ids = [177,178,386,387,413,428,429,430,431,470,471,490,496,504,605,613]
by_id = {row['id']: row for row in rows}
counts = collections.Counter(row['status'] for row in rows)
report = {
    'source_pin': '9b8ebd77', 'baseline_pin': '674b68b5', 'merged_main': '54cbc125',
    'measurement': census['measurement'],
    'coverage': {'checked': counts['checked relation'], 'proven': counts['proven relation'], 'refused': counts['stopped'], 'unmatched': counts['unmatched'], 'total': len(rows)},
    'slices': {
        'short_circuit': [by_id[identifier] for identifier in [112,233]],
        'generic_relations': [by_id[identifier] for identifier in generic_ids],
        'optional_boolean': [by_id[identifier] for identifier in [461,463,464]],
        'diagnostic_branches': [by_id[identifier] for identifier in [179,411,548]],
        'bottom_array_branches': [by_id[identifier] for identifier in [110,218]],
        'overloaded_callback_result': [by_id[482]],
    },
    'remaining_groups': groups,
}
(root / 'next-views-results.json').write_text(json.dumps(report, indent=2) + '\n')
print(json.dumps({'coverage': report['coverage'], 'remaining_groups': [{key: value for key, value in group.items() if key != 'records'} for group in groups]}, indent=2))
