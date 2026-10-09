#!/usr/bin/env python3
"""Audit the exact 115 delegation, 105 mixed-kind, and 138 Q5 coordinates."""
import gzip, hashlib, json, sys
from collections import Counter
from pathlib import Path

GROUPS = ['delegation or composition of predicate calls', 'kind partitions with control flow or extra conditions']
def read(path):
    return json.loads(gzip.decompress(Path(path).read_bytes()) if str(path).endswith('.gz') else Path(path).read_bytes())
def key(where):
    return 'src/compiler/'+where.split('/src/compiler/')[-1]
def measure(raw, baseline, selected):
    assert raw['roots'] == 79 and raw['checkerDiagnostics'] == baseline['checkerDiagnostics'], 'checker/root drift'
    rows = {key(row['location']): row for row in raw['predicates']}
    assert len(rows) == len(raw['predicates']) == len(selected) and set(rows) == set(selected), 'body coverage drift'
    calls = [call for row in rows.values() for call in row['calls']]
    expected = {call['where'] for row in selected.values() for call in row['calls']}
    assert {key(call['where']) for call in calls} == expected and len(calls) == len(expected) == raw['matchedCalls'], 'call coverage drift'
    assert {call['ordinal'] for call in calls} == set(range(len(calls))), 'call ordinal drift'
    enriched=[]
    for location,row in rows.items():
        assert 'ProbeError:' not in row['proofDiagnostic']+row['adamicDiagnostic'], 'body probe failure'
        assert (row['admission']=='Proven') == (row['bodyProof'] and not row['adamicDiagnostic']), 'invented admission'
        if not row['bodyProof']:
            assert row['admission']=='Refused' and row['proofDiagnostic'] and 'Adamic 0.1 refuses' in row['adamicDiagnostic'], 'unproved body admitted'
        checked=[];pending=[]
        metadata={call['where']:call for call in selected[location]['calls']}
        for call in row['calls']:
            where=key(call['where']);call['where']=where
            assert 'ProbeError:' not in call['diagnostic'], 'call probe failure'
            call['viewTarget']=metadata[where]['viewTarget']
            if call['status']=='CheckedSeam':
                assert not call['diagnostic'] and any(direction['Status']=='checked' for direction in call['directions']), 'invented check'
                checked.append(where)
            elif call['status']=='Pending':
                assert call['diagnostic'], 'unnamed pending call'
                pending.append(call)
            else: assert call['status']=='ProvenSeam', 'unknown call status'
        words=['checked-view','checked field','checked view','checked object','open tag domain','view schema','view contract','getter in a checked','checked predicate overload array','recursive predicate overload target']
        views=row['admission']=='PendingView' or any(any(word in call['diagnostic'] for word in words) or call['viewTarget'] and 'checked predicate overload target' in call['diagnostic'] for call in pending)
        status=('ProvenBodyPendingCalls' if pending else 'Proven') if row['admission']=='Proven' else 'PartlyCheckedSeam' if checked and pending else 'CheckedSeam' if checked else 'PendingView' if views else 'PendingOther' if pending else 'NoResolvedCall'
        enriched.append({**row,'location':location,'name':selected[location]['name'],'group':selected[location]['group'],'ts':status,'pendingViews':views,'checkedCallSites':checked})
    table=[]
    for group in GROUPS:
        members=[row for row in enriched if row['group']==group]
        table.append({'group':group,'bodies':len(members),'bodyProof':sum(row['bodyProof'] for row in members),'adamic':dict(Counter(row['admission'] for row in members)),'ts':dict(Counter(row['ts'] for row in members)),'checkedPredicates':sum(bool(row['checkedCallSites']) for row in members),'checkedCalls':sum(len(row['checkedCallSites']) for row in members),'pendingViews':sum(row['pendingViews'] for row in members),'proofReasons':dict(Counter(row['proofDiagnostic'].split(';')[1].strip() for row in members if not row['bodyProof']))})
    return {'table':table,'predicates':enriched,'matchedCalls':len(calls)}
def audit(baseline,before,after,optional,original,tree):
    for name,expected in baseline['sourceHashes'].items():
        assert hashlib.sha256((tree/name).read_bytes()).hexdigest()==expected,'source hash drift'
    selected={row['location']:row for row in baseline['predicates'] if row['group'] in GROUPS}
    assert len(selected)==220 and Counter(row['group'] for row in selected.values())==dict(zip(GROUPS,[115,105])), 'group drift'
    old=measure(before,baseline,selected);new=measure(after,baseline,selected)
    expected={row['location']:row for row in optional['selected'] if 'utilitiesPublic.ts:768:16' in row['adamicDiagnostic']}
    actual={key(row['location']):row for row in original['predicates']}
    assert len(expected)==len(actual)==len(original['predicates'])==138 and set(expected)==set(actual),'Q5 coverage drift'
    assert original['roots']==79 and original['checkerDiagnostics']==baseline['checkerDiagnostics'] and original['matchedCalls']==0,'Q5 probe drift'
    for location,row in actual.items():
        assert row['bodyProof'] and row['admission']=='PendingView' and 'utilitiesPublic.ts:768:16' in row['adamicDiagnostic'],'Q5 stop changed: '+location
        assert "checked field alias requiring an optional, accessor, or representation conversion" in row['adamicDiagnostic'],'Q5 reason changed'
    return {'before':old,'after':new,'optionalOriginal':original['predicates'],'sourceHashes':baseline['sourceHashes'],'checkerDiagnostics':baseline['checkerDiagnostics'],'baseCompiler':'223f233a8d5537d4d9b9343d4fcaedbac5e534fe','sourceEvidence':'codex/step09-double-casts 8a7ab17e'}
if __name__=='__main__':
    baseline,before,after,optional,original=map(read,sys.argv[1:6])
    result=audit(baseline,before,after,optional,original,Path(sys.argv[6]))
    Path(sys.argv[7]).write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps({mode:result[mode]['table'] for mode in ['before','after']},indent=2))
    print('PASS: 220 exact bodies, 1134 resolved calls, 138 unchanged node.original stops; 79 source hashes and 320 checker diagnostics retained')
