#!/usr/bin/env python3
"""Compare only the baseline's 289 optional-alias stops, with the full pinned context."""
import collections, gzip, hashlib, json, sys
from pathlib import Path
baseline = json.load(gzip.open(sys.argv[1]))
current = json.load(open(sys.argv[2]))
root = Path(sys.argv[3]).resolve()
for name, wanted in baseline['sourceHashes'].items():
    assert hashlib.sha256((root / name).read_bytes()).hexdigest() == wanted, name
assert current['roots'] == baseline['roots'] == 79
assert current['checkerDiagnostics'] == baseline['checkerDiagnostics']
assert current['matchedCalls'] == 0, 'body-only measurement must not claim call results'
rows = {}
for row in current['predicates']:
    location = row['location'].removeprefix(str(root) + '/')
    assert location not in rows
    assert 'ProbeError' not in row['adamicDiagnostic'] + row['proofDiagnostic']
    assert row['admission'] != 'Proven' or (row['bodyProof'] and not row['adamicDiagnostic'])
    rows[location] = row
assert len(rows) == 580
assert set(rows) == {row['location'] for row in baseline['predicates']}
selected = [row for row in baseline['predicates'] if 'a checked field alias requiring an optional, accessor, or representation conversion' in row['adamicDiagnostic']]
assert len(selected) == 289
result = []
for old in selected:
    new = rows[old['location']]
    assert new['bodyProof'] == old['bodyProof'], 'conversion must not invent a body proof'
    result.append(dict(location=old['location'], group=old['group'], oldDiagnostic=old['adamicDiagnostic'], **{key:new[key] for key in ['bodyProof','admission','adamicDiagnostic']}))
print(json.dumps(dict(sourceHashes=baseline['sourceHashes'], selected=result, groups=[dict(group=group,status=status,bodies=count) for (group,status),count in sorted(collections.Counter((row['group'],row['admission']) for row in result).items())]), indent=2))
