#!/usr/bin/env python3
"""Record the pinned refusal groups and their required contract support."""
import collections
import json
from pathlib import Path

root = Path(__file__).resolve().parent
baseline = json.loads((root / 'flow-container-containers-census.json').read_text())
current = json.loads((root / 'census.json').read_text())
class_d = [row for row in current['records'] if row['class'] == 'd']
assert len(class_d) == 515
assert all(row['status'] in {'checked relation', 'proven relation', 'stopped'} for row in class_d)
by_id = {row['id']: row for row in class_d}
previous_refusals = [row for row in baseline['records'] if row['class'] == 'd' and row['status'] == 'stopped']
previous_unmatched = [row for row in baseline['records'] if row['class'] == 'd' and row['status'] == 'unmatched']
assert len(previous_refusals) == 78 and len(previous_unmatched) == 11

capabilities = {
    'EmitNode.autoGenerate': 'Delivered: represented plain intersections and actual allocation contracts, including compatible callback parameters.',
    'compatible callback fields': 'Delivered: reversed parameter field relations and forward result field relations for one non-generic signature.',
    'boxed union fields or elements': 'Further step 10 work: boxed-union slot contracts and validated representation conversion. The existing boxed representation alone is insufficient.',
    'uninstantiated writable generic relation': 'Further generic relation work: instantiate both receiving and producing signatures, then prove the concrete relation or retain its allocation contract. Identically printed T names are not identity proofs.',
    'tuple slots including callable elements': 'Further step 10 work: allocation-time positional tuple contracts and callable parameter/result proofs. Each position must retain its declared contract.',
    'optional boolean field': 'Further step 10 work: boolean-or-undefined contracts, checking presence and any literal domain in the existing packed representation.',
    'inferred diagnostic branch loses narrow relation': 'Further step 10 work: retain and check each conditional branch against the receiving diagnostic slot before its inferred wider type hides the narrow relation.',
    'inferred bottom-array branch loses narrow relation': 'Further step 10 work: retain each conditional branch allocation declaration; the inferred array type does not identify the never-element allocation.',
    'recursive structural reference': 'Further step 10 work: cycle-safe recursive reference contracts and directional lifetime proofs for nested mutable references.',
    'array versus Map field representation': 'A write check alone cannot lift this view. It needs a proven copy or representation-safe reads and writes; retain the refusal without that proof.',
    'constrained generic array union': 'Further step 10 and generic relation work: instantiate the element union and record the actual array declaration rather than treating a constraint as the allocated element type.',
    'Set element': 'Further step 10 work: Set element contracts at allocation and before add; Map value contracts do not cover Set keys.',
    'overloaded callback result': 'Further callback relation work: prove every applicable overload parameter/result relation and preserve allocation contracts; the current rule accepts one non-generic signature.',
}

def group(row):
    rid, reason = row['id'], row['first_reason']
    if rid == 482:
        return 'overloaded callback result'
    if 'EmitNode & { autoGenerate: AutoGenerateInfo; }' in reason:
        return 'EmitNode.autoGenerate'
    if rid in {414, 521}:
        return 'compatible callback fields'
    if 'a type parameter whose constraint' in reason:
        return 'uninstantiated writable generic relation'
    if rid in {207, 208, 313, 373}:
        return 'tuple slots including callable elements'
    if rid in {461, 463, 464}:
        return 'optional boolean field'
    if rid in {179, 411, 548}:
        return 'inferred diagnostic branch loses narrow relation'
    if rid in {110, 218}:
        return 'inferred bottom-array branch loses narrow relation'
    if rid in {151, 152}:
        return 'recursive structural reference'
    if rid in {112, 233}:
        return 'array versus Map field representation'
    if rid == 53:
        return 'constrained generic array union'
    if rid == 89:
        return 'Set element'
    if rid in {165, 166, 478, 479, 480, 481, 483, 484, 485, 486, 492, 493, 494, 537}:
        return 'boxed union fields or elements'
    raise ValueError(f'unclassified pinned refusal {rid}')

def grouped(rows):
    buckets = collections.defaultdict(list)
    for row in rows:
        buckets[group(row)].append(row)
    return [dict(reason=name, count=len(items), required_support=capabilities[name], records=items)
            for name, items in sorted(buckets.items(), key=lambda item: (-len(item[1]), item[0]))]

old_groups = grouped(previous_refusals)
for item in old_groups:
    item['now_checked'] = sum(by_id[row['id']]['status'] == 'checked relation' for row in item['records'])
    item['now_refused'] = sum(by_id[row['id']]['status'] == 'stopped' for row in item['records'])
remaining = [row for row in class_d if row['status'] == 'stopped']
report = {
    'measurement': current['measurement'],
    'source_pin': '9b8ebd77',
    'merged_main': '54cbc125',
    'baseline_pin': 'f84d5e45',
    'class_d_counts': current['class_d_counts'],
    'coverage': dict(checked=sum(r['status'] == 'checked relation' for r in class_d), proven=sum(r['status'] == 'proven relation' for r in class_d), refused=len(remaining), unmatched=0, total=515),
    'prior_78_refusal_groups': old_groups,
    'resolved_11': [by_id[row['id']] for row in previous_unmatched],
    'remaining_groups': grouped(remaining),
    'roadmap_scope': 'Step 10, #63x2441. Required capabilities are named; no later numbered roadmap dependency for these gaps is pinned in the repository.',
}
(root / 'emit-results.json').write_text(json.dumps(report, indent=2) + '\n')
print(current['class_d_counts'])
for item in report['remaining_groups']:
    print(item['count'], item['reason'])
