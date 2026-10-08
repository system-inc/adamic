"""Group every checker diagnostic and count current cast-frontier obligations.

Affected sites overlap. Clearing one group only releases the recorded diagnosis
barrier if every diagnostic attached to that site's frontier belongs to the group.
This counterfactual is not a proof that rechecking the edited source will be free.
"""
import collections
import gzip
import json
import pathlib
import re
import sys


def grouped(result, any_dependency=False):
    records = {}
    for entry in result['diagnostic_sites']:
        match = re.match(r'^(.*?):(\d+):(\d+): error (TS\d+): (.*)', entry['text'])
        assert match, entry['text']
        absolute, line, column, code, message = match.groups()
        file = 'src/compiler/' + absolute.split('/src/compiler/', 1)[1]
        records[entry['text']] = dict(file=file, code=code, line=int(line), column=int(column), message=message)
    assert set(records) == set(result['diagnostics']), 'raw diagnostic inventory mismatch'
    groups = collections.defaultdict(list)
    for raw, record in records.items():
        groups[(record['code'], record['file'])].append(raw)
    rows = []
    for (code, file), raw in groups.items():
        fixes = set(raw)
        affected = {'own function': set(), 'dependency': set()}
        cleared = set()
        affected_kinds = collections.Counter()
        for site in result['sites']:
            if site['reason'] != 'diagnosed body':
                continue
            identity = (site['file'], site['start'], site['end'])
            obligations = set()
            for cause in site['diagnostic_causes']:
                assert cause['diagnostics'] and set(cause['diagnostics']) <= records.keys()
                obligations.update(cause['diagnostics'])
                if fixes.intersection(cause['diagnostics']):
                    affected[cause['scope']].add(identity)
            assert obligations
            if fixes.intersection(obligations):
                affected_kinds[site['kind']] += 1
                if (bool(fixes.intersection(obligations)) if any_dependency else obligations <= fixes):
                    cleared.add(identity)
        union = affected['own function'] | affected['dependency']
        assert len(union) == sum(affected_kinds.values())
        rows.append(dict(code=code, file=file, diagnostics=len(raw), own_function_sites=len(affected['own function']), dependency_sites=len(affected['dependency']), affected_sites=len(union), affected_kinds=dict(affected_kinds), recorded_barrier_cleared_alone=len(cleared), locations=[records[d] for d in sorted(raw, key=lambda d: (records[d]['line'], records[d]['column']))]))
    rows.sort(key=lambda row: (-row['affected_sites'], row['file'], row['code']))
    assert sum(row['diagnostics'] for row in rows) == len(records)
    return dict(measurement=result['measurement'], diagnostic_count=len(records), diagnosed_cast_sites=sum(site['reason'] == 'diagnosed body' for site in result['sites']), groups=rows, note='Affected sites overlap and count a diagnosed dependency unblocked. Cleared alone counts all recorded diagnosis obligations removed by fixing this entire code/file group; rechecking can expose other frontiers. Neither count certifies free casts.')


def controls():
    texts = ['/tmp/tree/src/compiler/one.ts:1:1: error TS2322: wrong assignment', '/tmp/tree/src/compiler/two.ts:2:1: error TS2345: wrong argument']
    def site(start, diagnostics):
        return dict(file='src/compiler/one.ts', start=start, end=start+1, kind='tagged', reason='diagnosed body', diagnostic_causes=[dict(scope='dependency', diagnostics=diagnostics)])
    fixture = dict(measurement='measured on a checker-rejected program', diagnostics=texts, diagnostic_sites=[dict(text=text) for text in texts], sites=[site(0, texts), site(2, texts[:1])])
    rows = grouped(fixture)['groups']
    assert [(row['affected_sites'], row['recorded_barrier_cleared_alone']) for row in rows] == [(2, 1), (1, 0)]
    mutant = grouped(fixture, any_dependency=True)['groups']
    assert [(row['affected_sites'], row['recorded_barrier_cleared_alone']) for row in mutant] != [(2, 1), (1, 0)]
    print('Overlapping-dependency mutant caught: one fixed producer does not clear every diagnosis barrier.')


if __name__ == '__main__':
    controls()
    source, output, table = map(pathlib.Path, sys.argv[1:4])
    with gzip.open(source, 'rt') if source.suffix == '.gz' else source.open() as stream:
        result = json.load(stream)
    report = grouped(result)
    output.write_text(json.dumps(report, indent=2)+'\n')
    lines = ['Every number below is measured on a checker-rejected program.', '', 'The table covers all 315 diagnostics, grouped by code and file. Affected sites count casts whose own function or a dependency carries that group; these columns overlap across rows. Clearing the group unblocks those recorded dependencies. “Cleared alone” counts casts whose entire recorded diagnostic frontier belongs to this group. It is a counterfactual over current edges: rechecking edited source may expose other dependencies, and host/unsupported/readiness barriers remain. It is not an observed free-cast count.', '', '| Code | File | Diagnostics | Own function sites | Dependency sites | Affected sites | Cleared alone |', '|---|---|---:|---:|---:|---:|---:|']
    for row in report['groups']:
        lines.append(f"| {row['code']} | `{row['file']}` | {row['diagnostics']} | {row['own_function_sites']} | {row['dependency_sites']} | {row['affected_sites']} | {row['recorded_barrier_cleared_alone']} |")
    lines += ['', f"Diagnostic count check: {report['diagnostic_count']}; diagnosed casts: {report['diagnosed_cast_sites']}. Exact adapted line/column locations and messages for every diagnostic are in diagnostic-actions.json. Locations refer to the source hashes at e7f42029, not a newer adaptation tree."]
    table.write_text('\n'.join(lines)+'\n')
    print(json.dumps(dict(diagnostics=report['diagnostic_count'], groups=len(report['groups']), diagnosed_cast_sites=report['diagnosed_cast_sites'], top=[{key:row[key] for key in ('code','file','diagnostics','affected_sites','recorded_barrier_cleared_alone')} for row in report['groups'][:5]]), indent=2))
