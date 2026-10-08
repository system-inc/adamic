"""Audit each snapshot, then join exact cast locations without inferred shares."""
import argparse
import collections
import gc
import gzip
import hashlib
import importlib.util
import json
import pathlib

parser = argparse.ArgumentParser()
parser.add_argument('before', type=pathlib.Path)
parser.add_argument('after', type=pathlib.Path)
parser.add_argument('mapped', type=pathlib.Path)
parser.add_argument('adapted_root', type=pathlib.Path)
parser.add_argument('output', type=pathlib.Path)
parser.add_argument('--attributions', type=pathlib.Path)
args = parser.parse_args()
base = pathlib.Path(__file__).resolve().parents[1]
spec = importlib.util.spec_from_file_location('census_artifacts', base / 'dynamic-keys/artifacts.py')
artifacts = importlib.util.module_from_spec(spec)
spec.loader.exec_module(artifacts)
mapped = json.loads(args.mapped.read_text())


def read(path):
    with gzip.open(path, 'rt') if path.suffix == '.gz' else path.open() as stream:
        value = json.load(stream)
    if value.get('format') == 'interned-census-json-v1':
        decoded = artifacts.decode(value)
        assert artifacts.json_hash(decoded) == value['json_sha256'], 'changed lossless observations'
        return decoded
    return value


def key(row):
    return row['file'], row['start'], row['end']


def compact(row):
    return {
        'outcome': row['outcome'],
        'cause': row['reason'],
        'detail_sha256': artifacts.json_hash(sorted(set(row['detail']))),
        'diagnostic_causes': len(row.get('diagnostic_causes') or []),
    }


def bucket(row):
    if row['outcome'] != 'unknown':
        return row['outcome']
    return 'unknown: ' + row['cause']


def totals(rows):
    result = {}
    for kind in ('tagged', 'untagged', 'all'):
        selected = [r for r in rows if kind == 'all' or r['kind'] == kind]
        counts = collections.Counter(bucket(r) for r in selected)
        free = counts['conforms and ready (free)']
        ready_only = counts['conforms but not proven ready (readiness checks only)']
        result[kind] = {
            'sites': len(selected), 'buckets': dict(counts),
            'free': free, 'free_share': free / len(selected),
            'proven_conforming': free + ready_only,
            'proven_conforming_share': (free + ready_only) / len(selected),
            'conditional_conformance': counts['conforms-if'],
        }
    return result


before = read(args.before)
artifacts.audit(before, mapped, args.adapted_root)
old = {key(r): compact(r) | {'kind': r['kind'], 'text': r['text']} for r in before['sites']}
assert len(old) == 2936
before_totals = totals(list(old.values()))
diagnostics = set(before['diagnostics'])
before_schemas = len(before['allocation_schemas'])
del before
gc.collect()
after = read(args.after)
artifacts.audit(after, mapped, args.adapted_root)
assert set(after['diagnostics']) == diagnostics, 'checker inventory changed'
assert {key(r) for r in after['sites']} == set(old), 'cast identity inventory changed'
attributions = json.loads(args.attributions.read_text()) if args.attributions else {}
rows, newly_free, lost_free, transitions = [], [], [], collections.Counter()
valid_rules = {'object-binding-property-edge', 'nonnull-callable-reference'}
for site in after['sites']:
    previous = old[key(site)]
    assert site['kind'] == previous['kind'] and site['text'] == previous['text']
    current = compact(site)
    became_free = previous['outcome'] != 'conforms and ready (free)' and current['outcome'] == 'conforms and ready (free)'
    ceased_free = previous['outcome'] == 'conforms and ready (free)' and current['outcome'] != 'conforms and ready (free)'
    identifier = '%s:%d:%d' % key(site)
    attribution = attributions.get(identifier)
    if became_free:
        assert attribution, 'producer-path attribution required for newly Free site: ' + identifier
        assert set(attribution['rules']) <= valid_rules and attribution['rules']
        assert attribution['evidence'], 'source producer evidence required'
        newly_free.append(identifier)
    else:
        assert attribution is None, 'attribution supplied for a site that was not newly Free'
    if ceased_free:
        lost_free.append(identifier)
    rows.append({
        **{k: site[k] for k in ('file', 'line', 'column', 'start', 'end', 'kind', 'text')},
        'site_key': identifier, 'before': {k: previous[k] for k in current}, 'after': current,
        'newly_free': became_free, 'ceased_free': ceased_free,
        'erasure_attribution': attribution,
    })
    transitions[(bucket(previous), bucket(current), site['kind'])] += 1
assert set(attributions) == set(newly_free)
args.output.mkdir(parents=True, exist_ok=True)
(args.output / 'per-site-table.json').write_text(json.dumps(rows, indent=2) + '\n')
new_rows = [r for r in rows if r['newly_free']]
(args.output / 'newly-erased-sites.json').write_text(json.dumps(new_rows, indent=2) + '\n')
changes = [r for r in rows if r['before']['outcome'] != r['after']['outcome'] or r['before']['cause'] != r['after']['cause']]
(args.output / 'changed-buckets.json').write_text(json.dumps(changes, indent=2) + '\n')
result = {
    'measurement': after['measurement'], 'status': 'PASS',
    'before': before_totals,
    'after': totals([compact(r) | {'kind': r['kind']} for r in after['sites']]),
    'per_site_rows': len(rows), 'unchanged_diagnostics': len(diagnostics),
    'before_allocation_schemas': before_schemas, 'after_allocation_schemas': len(after['allocation_schemas']),
    'newly_free_sites': newly_free, 'lost_free_sites': lost_free,
    'changed_buckets': len(changes),
    'changed_detail_sets': sum(r['before']['detail_sha256'] != r['after']['detail_sha256'] for r in rows),
    'changed_diagnostic_frontier_counts': sum(r['before']['diagnostic_causes'] != r['after']['diagnostic_causes'] for r in rows),
    'transitions': [{'before': a, 'after': b, 'kind': k, 'sites': n} for (a, b, k), n in sorted(transitions.items())],
    'definition': 'Proven conformance is Free plus readiness-only. Conforms-if remains conditional. Free is analysis eligibility for erasure; rejected source produces no production IR.',
    'rule_scope': {
        'uninitialized-declaration-cause': 'names an existing Unknown; cannot itself erase a cast',
        'object-binding-property-edge': 'connects a bound identifier to its initializer field',
        'iteration-producer-frontier': 'retains Unknown and iterable provenance; cannot itself erase a cast',
        'nonnull-callable-reference': 'preserves an already represented operand identity through reference classification',
    },
}
(args.output / 'summary.json').write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps(result, indent=2), flush=True)
