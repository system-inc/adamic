#!/usr/bin/env python3
"""Audit the pinned 445 condition calls without claiming a full corpus build."""
import gzip, hashlib, json, sys
from pathlib import Path
baseline=json.load(gzip.open(sys.argv[1]))
measured=json.load(open(sys.argv[2]))
root=Path(sys.argv[3]).resolve()
for name,wanted in baseline['sourceHashes'].items():
    assert hashlib.sha256((root/name).read_bytes()).hexdigest()==wanted,name
assert measured['roots']==baseline['roots']==79
assert measured['checkerDiagnostics']==baseline['checkerDiagnostics']
assert measured['matchedCalls']==445
assert len(measured['predicates'])==1
row=measured['predicates'][0]
assert row['location'].removeprefix(str(root)+'/')=='src/compiler/debug.ts:213:142'
assert row['bodyProof'] and row['admission']=='Proven'
assert not row['proofDiagnostic'] and not row['adamicDiagnostic']
old=next(r for r in baseline['predicates'] if r['location']=='src/compiler/debug.ts:213:142')
expected={call['where'] for call in old['calls']}
assert len(expected)==445
actual={call['where'].removeprefix(str(root)+'/') for call in row['calls']}
assert actual==expected and len(row['calls'])==445
assert {call['ordinal'] for call in row['calls']}==set(range(445))
for call in row['calls']:
    assert call['status']=='ProvenSeam' and not call['diagnostic']
    assert len(call['directions'])==1
    direction=call['directions'][0]
    assert direction['Direction']=='asserts' and direction['Status']=='proven'
    assert direction['Reason'].endswith(' -> truthy condition')
print('Debug.assert: 445 proven, 0 checked, 0 pending; all 79 source hashes and original call coordinates match')
