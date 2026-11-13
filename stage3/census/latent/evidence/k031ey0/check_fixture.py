import json
from pathlib import Path
import sys
sys.path.insert(0, '/workspace/adamic/stage3/census/latent')
from pass_ledger import body_lowers
from stream_records import census_records
root = Path(__file__).resolve().parent
for mode in ('full', 'first-error', 'speculative'):
    for side in ('base', 'head'):
        rows = census_records([json.loads(line) for line in (root / f'{side}-generic-{mode}.jsonl').read_text().splitlines()])
        assert rows[0]['diagnostics'] == [], rows[0]
        findings = [f for row in rows[1:] for f in row['findings'] if f['kind'] == 'NotYet' and f['reason'].startswith('a view of a type that reaches ')]
        assert bool(findings) == (side == 'head'), (side, mode, findings)
        if side == 'head':
            assert all(f['phase'] == 'prepass' for f in findings)
            for row in rows[1:]:
                for unit in row['units']:
                    if unit['status'] not in ('attempted', 'speculative'): continue
                    assert not body_lowers(unit, row['findings']), unit
                    assert not any(p['status'].startswith('executed') for p in unit.get('pass_ledger', [])), unit
        print(side, mode, 'view findings', len(findings), 'checker diagnostics 0; eligibility verified')
