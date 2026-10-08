"""Compare only observed lowering NotYet root signatures on one frozen manifest."""
import argparse
import collections
import csv
import gzip
import hashlib
import json
from pathlib import Path

PREFIX = '/tmp/stage3-notyet-adapted/'
SIGNATURE = ('kind', 'where', 'reason', 'text')
CAUSE = ('unit', 'phase', *SIGNATURE)


def normalized(value):
    return value.replace(PREFIX, '')


def signature(finding):
    return tuple(normalized(finding[field]) for field in SIGNATURE)


def dump(path, value):
    path.write_text(json.dumps(value, indent=2, sort_keys=True) + '\n')


def csv_out(path, rows, fields):
    with path.open('w', newline='') as stream:
        writer = csv.DictWriter(stream, fieldnames=fields, lineterminator='\n')
        writer.writeheader()
        writer.writerows(rows)


def read_records(path):
    opener = gzip.open if str(path).endswith('.gz') else open
    with opener(path, 'rt') as stream:
        return [json.loads(line) for line in stream]


def verify_manifest(manifest, source):
    assert len(manifest) == 81 and len({row['file'] for row in manifest}) == 81, '81 unique source files'
    for row in manifest:
        raw = (source / row['file']).read_bytes()
        assert len(raw) == row['bytes'], ('source bytes', row['file'])
        assert hashlib.sha256(raw).hexdigest() == row['sha256'], ('source hash', row['file'])


def census(records, manifest):
    metadata, files = records[0], records[1:]
    assert metadata['latent_mode'] == 'full', 'full latent census only'
    assert {normalized(row['file']) for row in files} == {row['file'] for row in manifest}, 'exact frozen source set'
    assert len(files) == 81, 'one result for each frozen file'
    assert all(row['measurement'] == metadata['measurement'] for row in files), 'measurement labels retained'
    sites = collections.defaultdict(list)
    edges = set()
    other_roots = set()
    units = {}
    for record in files:
        actual = {tuple(f[k] for k in CAUSE): f for f in record['findings']}
        for unit in record['units']:
            key = (normalized(record['file']), normalized(unit['where']), unit['kind'])
            value = (unit['status'], json.dumps(unit['checker_diagnostics'], sort_keys=True))
            assert key not in units or units[key] == value, 'unit status agrees'
            units[key] = value
        for finding in record['findings']:
            if (finding['phase'], finding['kind']) != ('lowering', 'NotYet'):
                continue
            site = signature(finding)
            sites[site].append(finding)
            cause = finding.get('blocked_by')
            if cause:
                assert finding['reason'].startswith('reading '), 'tagged read'
                assert finding.get('blocked_symbol_declaration'), 'declaration symbol provenance'
                root = actual.get(tuple(cause[k] for k in CAUSE))
                assert root is not None, 'actual root exists in the same attempt'
                assert root['phase'] == 'lowering' and not root.get('blocked_by'), 'flattened lowering root'
                root_signature = signature(root)
                assert root_signature != site, 'no self echo'
                edges.add((root_signature, site))
                if root['kind'] != 'NotYet':
                    other_roots.add(root_signature)
    echoes = {site for site, observations in sites.items() if all(f.get('blocked_by') for f in observations)}
    mixed = {site for site, observations in sites.items() if any(f.get('blocked_by') for f in observations) and site not in echoes}
    roots = set(sites) - echoes
    # A read observed without provenance in any attempt remains a root conservatively.
    assert len(roots) + len(echoes) == len(sites)
    summary = dict(notyet_sites=len(sites), notyet_observations=sum(map(len, sites.values())),
                   echo_only_sites=len(echoes), mixed_sites=len(mixed), root_sites=len(roots),
                   root_reasons=len({site[2] for site in roots}),
                   unattributed_read_roots=sum(site[2].startswith('reading ') for site in roots),
                   referenced_other_roots=len(other_roots), echo_edges=len(edges),
                   checker_rejected=metadata['checker_rejected'], measurement=metadata['measurement'],
                   unit_statuses=dict(collections.Counter(value[0] for value in units.values())),
                   diagnostic_sites=len(metadata['diagnostic_sites']))
    return dict(roots=roots, echoes=echoes, mixed=mixed, sites=sites, edges=edges,
                other_roots=other_roots, summary=summary, units=units, metadata=metadata)


def compare(before, after):
    left, right = before['roots'], after['roots']
    retired, exposed, surviving = left - right, right - left, left & right
    rows = []
    grouped = collections.defaultdict(lambda: dict(retired=0, newly_exposed=0, surviving=0))
    for site in sorted(left | right):
        values = dict(retired=int(site in retired), newly_exposed=int(site in exposed), surviving=int(site in surviving))
        assert sum(values.values()) == 1, 'one comparison category per signature'
        rows.append(dict(zip(SIGNATURE, site), **values))
        for field, count in values.items():
            grouped[site[2]][field] += count
    reasons = [dict(reason=reason, **counts, net_retirement=counts['retired'] - counts['newly_exposed']) for reason, counts in grouped.items()]
    reasons.sort(key=lambda row: (-row['newly_exposed'], -row['retired'], row['reason']))
    same_site = {(site[0], site[1]) for site in retired} & {(site[0], site[1]) for site in exposed}
    unit_keys = before['units'].keys() | after['units'].keys()
    changed_units = [dict(attempt_file=key[0], where=key[1], kind=key[2], before=before['units'].get(key), after=after['units'].get(key))
                     for key in sorted(unit_keys) if before['units'].get(key) != after['units'].get(key)]
    totals = dict(before=len(left), after=len(right), retired=len(retired), newly_exposed=len(exposed),
                  surviving=len(surviving), net_retirement=len(retired) - len(exposed),
                  newly_exposed_at_sites_with_retired_signature=sum((site[0], site[1]) in same_site for site in exposed),
                  unit_eligibility_changes=len(changed_units))
    assert totals['before'] == totals['retired'] + totals['surviving'], 'before partition'
    assert totals['after'] == totals['newly_exposed'] + totals['surviving'], 'after partition'
    assert totals['net_retirement'] == totals['before'] - totals['after'], 'net retirement'
    assert all(sum(row[field] for row in rows) == totals[field] for field in ('retired', 'newly_exposed', 'surviving')), 'signature totals'
    assert all(sum(row[field] for row in reasons) == totals[field] for field in ('retired', 'newly_exposed', 'surviving')), 'reason totals'
    return totals, rows, reasons, changed_units


def save_run(output, name, result):
    directory = output / name
    directory.mkdir(exist_ok=True)
    dump(directory / 'summary.json', result['summary'])
    csv_out(directory / 'roots.csv', [dict(zip(SIGNATURE, site)) for site in sorted(result['roots'])], SIGNATURE)
    edge_rows = []
    for root, echo in sorted(result['edges']):
        edge_rows.append(dict(root_kind=root[0], root_where=root[1], root_reason=root[2], root_text=root[3],
                              echo_kind=echo[0], echo_where=echo[1], echo_reason=echo[2], echo_text=echo[3],
                              echo_only=echo in result['echoes'], mixed=echo in result['mixed']))
    csv_out(directory / 'echoes.csv', edge_rows, ['root_kind', 'root_where', 'root_reason', 'root_text', 'echo_kind', 'echo_where', 'echo_reason', 'echo_text', 'echo_only', 'mixed'])
    csv_out(directory / 'referenced-other-roots.csv', [dict(zip(SIGNATURE, site)) for site in sorted(result['other_roots'])], SIGNATURE)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--manifest', type=Path, required=True)
    parser.add_argument('--source', type=Path, required=True)
    parser.add_argument('--before', type=Path, required=True)
    parser.add_argument('--after', type=Path, required=True)
    parser.add_argument('--all', type=Path)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    manifest = json.loads(args.manifest.read_text())
    verify_manifest(manifest, args.source)
    args.output.mkdir(exist_ok=True)
    runs = dict(before=census(read_records(args.before), manifest), after=census(read_records(args.after), manifest))
    if args.all:
        runs['all'] = census(read_records(args.all), manifest)
    for name, result in runs.items():
        save_run(args.output, name, result)
    results = {}
    pairs = [('before', 'after')]
    if args.all:
        pairs += [('before', 'all'), ('after', 'all')]
    for first, second in pairs:
        name = first + '-to-' + second
        totals, rows, reasons, changed = compare(runs[first], runs[second])
        csv_out(args.output / (name + '-signatures.csv'), rows, [*SIGNATURE, 'retired', 'newly_exposed', 'surviving'])
        csv_out(args.output / (name + '-reasons.csv'), reasons, ['reason', 'retired', 'newly_exposed', 'surviving', 'net_retirement'])
        dump(args.output / (name + '-unit-changes.json'), changed)
        results[name] = dict(totals=totals, top_newly_exposed=[row for row in reasons if row['newly_exposed']][:10])
    dump(args.output / 'comparison.json', results)
    print(json.dumps(results, indent=2))


if __name__ == '__main__':
    main()
