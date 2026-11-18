"""Pair findings within a run and owner; keep locations as evidence."""
import collections
import json
import re
import sys
from pathlib import Path

root = re.compile(r'(?:/workspace/scratch/y852c3c-adapted|/tmp/y852c3c/adapted)/')
position = re.compile(r':\d+:\d+(?=\b)')
phase_order = dict(prepass=0, registration=1, refusal_scan=2, lowering=3, deferred=4, tail=5)
identity_fields = ('kind', 'phase', 'reason', 'text')
target_reason = 'a dictionary member narrowed to an array without an element-checked view'

def normalized(value):
    return root.sub('', value)

def semantics(finding):
    return tuple(position.sub('', normalized(finding.get(key, ''))) for key in identity_fields)

def location(finding):
    return normalized(finding.get('site_where') or finding.get('where') or '')

def rows(path):
    records = [json.loads(line) for line in path.read_text().splitlines()]
    header = records[0]
    if header['status'] == 'blocked':
        return header, {}, []
    assert records[-1]['status'] == 'relation_report', (path, 'missing trailer')
    files = {normalized(record['file']): record for record in records if 'file' in record}
    found = []
    for record in records:
        for finding in record.get('findings') or []:
            found.append(dict(finding, census_file=normalized(record.get('file', ''))))
    return header, files, found

def owner(finding):
    return finding['census_file'], normalized(finding.get('unit', ''))

def remove_pairs(base, head, key):
    head_index = collections.defaultdict(collections.deque)
    for index, entry in enumerate(head):
        head_index[key(entry)].append(index)
    used = set()
    remaining_base, paired = [], []
    for entry in base:
        candidates = head_index.get(key(entry))
        if candidates:
            index = candidates.popleft()
            used.add(index)
            paired.append((entry, head[index]))
        else:
            remaining_base.append(entry)
    return remaining_base, [entry for index, entry in enumerate(head) if index not in used], paired

def later(new, old):
    if new['phase'] != old['phase']:
        return phase_order.get(new['phase'], 0) > phase_order.get(old['phase'], 0)
    new_place = location(new).rsplit(':', 2)
    old_place = location(old).rsplit(':', 2)
    return (len(new_place) == len(old_place) == 3 and new_place[0] == old_place[0]
            and tuple(map(int, new_place[1:])) > tuple(map(int, old_place[1:])))

def pair(base, head):
    base, head, exact = remove_pairs(base, head, lambda row: (owner(row), semantics(row), location(row)))
    base, head, relocated = remove_pairs(base, head, lambda row: (owner(row), semantics(row)))
    base, head, changed = remove_pairs(base, head, lambda row: (owner(row), row['phase'], location(row)))
    # Location is only a candidate signal for later reach. Review against lowering
    # order, unit boundaries and exact source before assigning the final grade.
    removed_by_owner = collections.defaultdict(list)
    for row in base:
        if row['kind'] in {'Refused', 'NotYet', 'error'}:
            removed_by_owner[owner(row)].append(row)
    deltas = [dict(grade='changed', base=before, head=after) for before, after in changed]
    deltas += [dict(grade='gone', base=row) for row in base]
    for row in head:
        candidates = [old for old in removed_by_owner[owner(row)] if later(row, old)]
        hit = dict(grade='new', head=row)
        if candidates:
            hit['possible_earlier_base_stops'] = candidates
        deltas.append(hit)
    return {'identical': len(exact), 'relocated_same_semantics': [
        dict(base=before, head=after) for before, after in relocated], 'deltas': deltas}

if __name__ == '__main__':
    left, right = (rows(Path(name)) for name in sys.argv[1:3])
    result = pair(left[2], right[2])
    for label, (header, files, found) in zip(('base', 'head'), (left, right)):
        result[label] = dict(status=header['status'], format=header.get('module_format'),
                             diagnostics=len(header.get('diagnostics', [])),
                             files=len(files), units=sum(len(record['units']) for record in files.values()),
                             findings=len(found), target_sites=[row for row in found if row['reason'] == target_reason])
    print(json.dumps(result, indent=2))
