#!/usr/bin/env python3
"""Join the complete pinned body inventory to independently measured production seams."""
import gzip, hashlib, json, sys
from collections import Counter
from pathlib import Path

def read(p):
    return json.loads(gzip.decompress(Path(p).read_bytes()) if str(p).endswith('.gz') else Path(p).read_bytes())
def key(p):
    return 'src/compiler/' + p.split('/src/compiler/')[-1] if '/src/compiler/' in p else p

def audit(inventory, grouped, measured, tree):
    bodies = {p['location']: p for p in inventory['predicates'] if p['hasBody']}
    assert len(inventory['predicates']) == 651 and len(bodies) == 580, 'pinned body coverage drift'
    assert len(grouped['predicates']) == 580, 'group coverage drift'
    groups = {p['location']: p['group'] for p in grouped['predicates']}
    assert set(groups) == set(bodies), 'group coordinate drift'
    for name, expected in inventory['hashes'].items():
        assert hashlib.sha256((tree / name).read_bytes()).hexdigest() == expected, 'source hash drift: ' + name
    rows = {key(p['location']): p for p in measured['predicates']}
    assert len(measured['predicates']) == len(rows) == 580 and set(rows) == set(bodies), 'measurement body coverage drift'
    assert measured['roots'] == 79, 'root coverage drift'
    calls = [c for p in measured['predicates'] for c in p['calls']]
    ordinals = [c['ordinal'] for c in calls]
    assert len(ordinals) == len(set(ordinals)) == measured['matchedCalls'] and set(ordinals) == set(range(measured['matchedCalls'])), 'call coverage drift'
    assert len({c['where'] for c in calls}) == len(calls), 'call coordinate drift'
    report = []
    for location, source in bodies.items():
        r = rows[location]
        assert r['admission'] in ['Proven', 'Refused', 'PendingView'], 'invalid proof status'
        assert 'ProbeError:' not in r['proofDiagnostic'] + r['adamicDiagnostic'], 'probe failure: ' + location
        assert (r['admission'] == 'Proven') == (r['bodyProof'] and not r['adamicDiagnostic']), 'proof admission contradiction'
        adamic_kind = 'Proven' if not r['adamicDiagnostic'] else 'Refused' if 'Adamic 0.1 refuses ' in r['adamicDiagnostic'] else 'NotYet' if "stage 0 can't lower " in r['adamicDiagnostic'] else 'OtherError'
        assert adamic_kind != 'OtherError', 'unknown .a diagnostic category: ' + location
        calls = r['calls']
        for call in calls:
            call['where'] = key(call['where'])
            assert call['status'] in ['CheckedSeam', 'ProvenSeam', 'Pending'], 'invalid call status'
            assert 'ProbeError:' not in call['diagnostic'], 'call probe failure: ' + call['where']
            if call['status'] == 'CheckedSeam':
                assert any(d['Status'] == 'checked' for d in call['directions']), 'checked call without a constructed check'
                assert not call['diagnostic'], 'failed call labelled checked'
            if call['status'] == 'Pending':
                assert call['diagnostic'], 'pending call without diagnostic'
        checked = [c for c in calls if c['status'] == 'CheckedSeam']
        pending = [c for c in calls if c['status'] == 'Pending']
        view_words = ['checked-view', 'checked field', 'checked view', 'checked object', 'open tag domain', 'view schema', 'view contract', 'getter in a checked', 'checked predicate overload array', 'recursive predicate overload target']
        view_pending = r['admission'] == 'PendingView' or any(any(w in c['diagnostic'] for w in view_words) or c.get('viewTarget', False) and 'checked predicate overload target' in c['diagnostic'] for c in pending)
        # Keep non-reifiable target contracts distinct from a confirmed view stop.
        if r['admission'] == 'Proven':
            ts = 'ProvenBodyPendingCalls' if pending else 'Proven'
        elif checked and not pending:
            ts = 'CheckedSeam'
        elif checked:
            ts = 'PartlyCheckedSeam'
        elif view_pending:
            ts = 'PendingView'
        elif pending:
            ts = 'PendingOther'
        else:
            ts = 'NoResolvedCall'
        report.append({**source, **r, 'location': location, 'group': groups[location], 'ts': ts, 'adamicDiagnosticKind': adamic_kind, 'pendingViews': view_pending, 'checkedCallSites': [c['where'] for c in checked]})
    table = []
    for group in [g['group'] for g in grouped['groups']]:
        members = [r for r in report if r['group'] == group]
        table.append({'group': group, 'bodies': len(members), 'bodyProof': sum(r['bodyProof'] for r in members), 'adamic': dict(Counter(r['admission'] for r in members)), 'adamicDiagnosticKinds': dict(Counter(r['adamicDiagnosticKind'] for r in members)), 'ts': dict(Counter(r['ts'] for r in members)), 'pendingViews': sum(r['pendingViews'] for r in members), 'checkedPredicates': sum(bool(r['checkedCallSites']) for r in members), 'checkedCallSites': sum(len(r['checkedCallSites']) for r in members)})
    return {'table': table, 'predicates': report, 'checkerDiagnostics': measured['checkerDiagnostics'], 'sourceHashes': inventory['hashes'], 'roots': measured['roots']}

if __name__ == '__main__':
    inventory, grouped, measured = map(read, sys.argv[1:4])
    result = audit(inventory, grouped, measured, Path(sys.argv[4]))
    Path(sys.argv[5]).write_text(json.dumps(result, indent=2)+'\n')
    print(json.dumps(result['table'], indent=2))
