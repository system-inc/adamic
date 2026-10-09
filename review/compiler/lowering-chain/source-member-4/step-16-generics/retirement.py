"""Compare exact root blockers without treating a moved site as retirement."""
from collections import defaultdict
import gzip
import hashlib
import json
from pathlib import Path
import sys
import unittest

MUTANT = '--mutant' in sys.argv

def local(position):
    return position.split('/src/compiler/', 1)[-1]

def blockers(rows):
    return {(local(r['unit']), r['kind'], r['reason'], local(r['where'])) if MUTANT
            else (local(r['unit']), r['kind'], r['reason'])
            for r in rows if r['kind'] in ('NotYet', 'Refused')}

class RetirementTests(unittest.TestCase):
    def test_moved_site_is_not_a_retired_root(self):
        def row(unit, reason, where):
            return dict(unit=unit, kind='NotYet', reason=reason, where=where)
        before = [row('a:1:1', 'optional result', 'a:3:1'),
                  row('a:1:1', 'optional result', 'a:3:1'),
                  row('b:1:1', 'optional result', 'b:3:1')]
        after = [row('a:1:1', 'optional result', 'a:4:1'),
                 row('b:1:1', 'unsupported operation', 'b:4:1')]
        retired = blockers(before) - blockers(after)
        self.assertEqual(len(retired), 1)
        self.assertEqual(next(iter(retired))[:3], ('b:1:1', 'NotYet', 'optional result'))
        self.assertTrue(any(r[0] == 'b:1:1' for r in blockers(after)))


def read(path):
    return [json.loads(line) for line in Path(path).read_text().splitlines()]

def compare(before_path, after_path, parameter_path):
    before, after = read(before_path), read(after_path)
    assert before[0] == after[0], 'checker diagnostics or measurement mode changed'
    def units(records):
        return {local(u['where']): u for record in records[1:] for u in record['units']}
    old_units, new_units = units(before), units(after)
    assert old_units == new_units, 'root eligibility, identity, or body range changed'
    parameters = json.loads(Path(parameter_path).read_text())
    base = json.load(gzip.open(Path(__file__).with_name('baseline.json.gz'), 'rt'))
    assert parameters['hashes'] == base['source_sha256'], 'source corpus changed'
    assert hashlib.sha256(Path(before_path).read_bytes()).hexdigest() == base['census_sha256'], 'baseline input differs'
    old_rows = [r for record in before[1:] for r in record['findings']]
    new_rows = [r for record in after[1:] for r in record['findings']]
    old, new = blockers(old_rows), blockers(new_rows)
    selected = blockers(base['findings'])
    retired = sorted(selected - new)
    changed_roots = {r[0] for r in retired}
    affected = []
    for position in sorted(changed_roots):
        unit = old_units[position]
        affected.append(dict(unit=position, name=unit.get('name'),
            retired=[list(r[1:]) for r in retired if r[0] == position],
            remaining=[list(r[1:]) for r in sorted(new) if r[0] == position],
            newly_exposed=[list(r[1:]) for r in sorted(new-old) if r[0] == position]))
    grouped = defaultdict(set)
    for unit, kind, reason in retired:
        grouped[(kind, reason)].add(unit)
    result = dict(base=base['base'], before_sha256=base['census_sha256'],
        after_sha256=hashlib.sha256(Path(after_path).read_bytes()).hexdigest(),
        generic_go_sha256=hashlib.sha256(Path('internal/lower/generic.go').read_bytes()).hexdigest(),
        checker_diagnostics=len(before[0]['diagnostics']), source_sha256=parameters['hashes'],
        eligibility_identical=True, roots=affected,
        newly_exposed_blockers=[list(r) for r in sorted(new-old)],
        newly_blocked_roots=sorted({r[0] for r in new}-{r[0] for r in old}),
        newly_blocked_details=[dict(unit=position, name=old_units[position].get('name'),
            findings=[list(r[1:]) for r in sorted(new) if r[0] == position])
            for position in sorted({r[0] for r in new}-{r[0] for r in old})],
        retired_blocker_root_pairs=len(retired), affected_roots=len(affected),
        roots_without_remaining_findings=sum(not r['remaining'] for r in affected),
        cleared_roots_that_lost_refusal=sum(not r['remaining'] and any(b[0]=='Refused' for b in r['retired']) for r in affected),
        retired_reasons=[dict(kind=k[0], reason=k[1], roots=len(v)) for k,v in sorted(grouped.items())])
    dest = Path(__file__).parent
    with (dest / 'retirement.json.gz').open('wb') as raw:
        with gzip.GzipFile(filename='', mode='wb', fileobj=raw, mtime=0) as stream:
            stream.write(json.dumps(result, indent=2).encode())
    lines = ['## Measured root retirement', '',
        'The full guarded census was repeated over the identical adapted compiler corpus. Checker diagnostics, every root identity/eligibility/body range and every source hash were unchanged. Compare exact `(unit, kind, reason)` sets; a moved diagnostic does not retire a root. This is measurement-only, not whole-program compilation or runtime acceptance.', '',
        f"{len(retired)} scoped blocker/root pairs disappeared across {len(affected)} distinct roots. {result['roots_without_remaining_findings']} of those roots have no remaining Refused/NotYet finding in the continuing measurement. The rest expose or retain other blockers. Neither number credits the historical hidden-byte estimates as measured native progress.", '',
        'The complete affected root names, retired blockers, remaining blockers and newly exposed reasons are in [retirement.json.gz](step-16-generics/retirement.json.gz). The ledger also retains every newly exposed blocker across the corpus, including roots outside the selected baseline subset.', '',
        f"{len(result['newly_blocked_roots'])} roots gained a failure where the baseline measurement had none. These are recorded below, rather than hidden in a net total. {result['cleared_roots_that_lost_refusal']} roots that lost a Refused blocker became free of all findings; a disappearing refusal in a root that still fails is not acceptance of that program. The baseline and continuing corpus are checker-rejected, so an empty measurement finding set is not a valid-program or backend proof.", '',
        '| Retired exact blocker | Roots |', '| --- | ---: |']
    for row in result['retired_reasons']:
        lines.append('| ' + row['kind'] + ': ' + row['reason'].replace('|', chr(92)+'|') + ' | ' + str(row['roots']) + ' |')
    lines += ['', '| Root with no remaining measurement finding | Name |', '| --- | --- |']
    for row in affected:
        if not row['remaining']:
            lines.append('| ' + row['unit'] + ' | ' + str(row['name']) + ' |')
    lines += ['', '| Newly blocked root | Name | Exact findings |', '| --- | --- | --- |']
    for row in result['newly_blocked_details']:
        findings = '<br>'.join(kind + ': ' + reason for kind, reason in row['findings']).replace('|', chr(92)+'|')
        lines.append('| ' + row['unit'] + ' | ' + str(row['name']) + ' | ' + findings + ' |')
    (dest / 'retirement-section.md').write_text('\n'.join(lines)+'\n')
    print(json.dumps({k:result[k] for k in ['retired_blocker_root_pairs','affected_roots','roots_without_remaining_findings','retired_reasons']},indent=2))

if __name__ == '__main__':
    if '--test' in sys.argv or MUTANT:
        unittest.main(argv=[sys.argv[0]])
    else:
        compare(*sys.argv[1:])
