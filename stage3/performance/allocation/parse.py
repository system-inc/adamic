#!/usr/bin/env python3
"""Read V8 schemas, preserving exact live counts and unknown allocation totals."""
import argparse
import gzip
import json
from pathlib import Path

CATEGORIES = ('Node', 'Symbol', 'Type', 'Signature', 'Array', 'Map', 'String', 'Array backing', 'Other')
CONSTRUCTORS = {'Node4': 'Node', 'Token': 'Node', 'Identifier2': 'Node', 'Symbol4': 'Symbol', 'Type3': 'Type', 'Signature2': 'Signature'}
NAMES = {**CONSTRUCTORS, 'FixtureNode': 'Node', 'FixtureSymbol': 'Symbol', 'FixtureType': 'Type', 'FixtureSignature': 'Signature'}


def read_json(path):
    if not path.exists() and Path(str(path) + '.gz').exists():
        path = Path(str(path) + '.gz')
    if path.suffix == '.gz':
        with gzip.open(path, 'rt') as stream:
            return json.load(stream)
    return json.loads(path.read_text())


def snapshot(path, drop=None):
    data = read_json(path)
    metadata = data['snapshot']['meta']
    fields = metadata['node_fields']
    width = len(fields)
    positions = {name: fields.index(name) for name in ('type', 'name', 'id', 'self_size')}
    kinds = metadata['node_types'][positions['type']]
    strings = data['strings']
    values = data['nodes']
    # Edges are not needed for shallow self bytes or object-ID intersections.
    del data['edges']
    groups = {category: {} for category in CATEGORIES}
    for offset in range(0, len(values), width):
        kind = kinds[values[offset + positions['type']]]
        name = strings[values[offset + positions['name']]]
        if kind == 'object':
            category = NAMES.get(name, 'Array' if name == 'Array' else 'Map' if name == 'Map' else 'Other')
        elif kind in ('string', 'concatenated string', 'sliced string'):
            category = 'String'
        elif kind == 'array':
            category = 'Array backing'
        else:
            category = 'Other'
        if category == drop:
            continue
        identifier = values[offset + positions['id']]
        groups[category][identifier] = values[offset + positions['self_size']]
    return groups


def sampling(path):
    data = read_json(path)
    nodes = {}
    totals = {}
    def walk(node, parents):
        names = parents + [node['callFrame']['functionName']]
        category = 'unclassified'
        for name in reversed(names):
            if name in CONSTRUCTORS:
                category = CONSTRUCTORS[name] + '-associated stack'
                break
            if name in ('createBaseNode', 'createBaseTokenNode', 'createBaseIdentifier', 'createBaseSourceFileNode'):
                category = 'Node-associated stack'; break
            if name in ('createSymbol', 'createTransientSymbol', 'cloneSymbol'):
                category = 'Symbol-associated stack'; break
            if name in ('createType', 'createObjectType', 'createTypeReference'):
                category = 'Type-associated stack'; break
            if name == 'createSignature':
                category = 'Signature-associated stack'; break
        nodes[node['id']] = category
        totals[category] = totals.get(category, 0) + node['selfSize']
        for child in node.get('children', []):
            walk(child, names)
    walk(data['head'], [])
    samples = {}
    sampled_bytes = {}
    for item in data.get('samples', []):
        # V8 can return records whose allocation stack node has been pruned.
        # Preserve their bytes and counts instead of inventing an attribution.
        name = nodes.get(item['nodeId'], 'unresolved sample nodeId')
        samples[name] = samples.get(name, 0) + 1
        sampled_bytes[name] = sampled_bytes.get(name, 0) + item['size']
    return {'estimated_bytes_by_stack': sampled_bytes, 'sample_records_by_stack': samples,
            'total_estimated_bytes': sum(sampled_bytes.values()), 'call_tree_estimated_bytes': sum(totals.values()),
            'sample_records': len(data.get('samples', []))}


def analyze(folder, drop=None):
    observations = read_json(folder / 'observations.json')
    phases = ('parse', 'bind', 'check')
    live = {}
    cohorts = {}
    previous = {category: set() for category in CATEGORIES}
    baseline = folder / 'baseline.heapsnapshot'
    if baseline.exists() or Path(str(baseline) + '.gz').exists():
        previous = {category: set(values) for category, values in snapshot(baseline, drop).items()}
    for phase in phases:
        groups = snapshot(folder / (phase + '.heapsnapshot'), drop)
        live[phase] = {category: {'count': len(values), 'self_bytes': sum(values.values())} for category, values in groups.items()}
        cohorts[phase] = {category: {identifier: size for identifier, size in values.items() if identifier not in previous[category]}
                          for category, values in groups.items()}
        previous = {category: set(values) for category, values in groups.items()}
    at_end = previous
    released = snapshot(folder / 'released.heapsnapshot', drop)
    rows = []
    for phase in phases:
        allocation = {category: 0 for category in ('Node', 'Symbol', 'Type', 'Signature')}
        for constructor, count in observations['counters'].get(phase, {}).items():
            allocation[CONSTRUCTORS[constructor]] += count
        for category in CATEGORIES:
            born = cohorts[phase][category]
            survivors = {identifier: size for identifier, size in born.items() if identifier in at_end[category]}
            after_release = {identifier: size for identifier, size in born.items() if identifier in released[category]}
            allocated = allocation.get(category)
            # Negative mortality would mean the hooks or constructor mapping are wrong.
            if allocated is not None and allocated < len(born):
                raise ValueError(f'{phase}/{category}: live cohort exceeds constructor calls')
            rows.append({'phase': phase, 'constructor': category, 'exact_constructor_calls': allocated,
                         'observed_new_live_count': len(born), 'new_live_self_bytes': sum(born.values()),
                         'survives_check_count': len(survivors), 'survives_check_birth_self_bytes': sum(survivors.values()),
                         'gone_before_phase_snapshot_count': None if allocated is None else allocated - len(born),
                         'gone_by_check_count': len(born) - len(survivors),
                         'survives_release_count': len(after_release),
                         'survives_release_birth_self_bytes': sum(after_release.values())})
    result = {'live_at_boundaries': live, 'released': {category: {'count': len(values), 'self_bytes': sum(values.values())}
                                                    for category, values in released.items()},
              'phase_cohorts': rows, 'sampling_including_collected': sampling(folder / 'collected.heapprofile'),
              'interpretation': 'Counts/self bytes in snapshots are exact live graph observations after snapshot GC. Cumulative allocation calls are exact only for hooked compiler constructors. Cohorts omit objects dead before a snapshot. Array/Map/String cumulative allocations and allocated bytes cannot be recovered by constructor from sampling call stacks.'}
    return result


def verify_fixture(result, observations):
    expected_live = {'parse': {'Node': 6, 'Symbol': 2, 'Type': 0, 'Signature': 0},
                     'bind': {'Node': 7, 'Symbol': 3, 'Type': 3, 'Signature': 1},
                     'check': {'Node': 7, 'Symbol': 4, 'Type': 7, 'Signature': 2}}
    expected_allocated = {'Node': 13, 'Symbol': 9, 'Type': 12, 'Signature': 5}
    for phase, categories in expected_live.items():
        for category, wanted in categories.items():
            actual = result['live_at_boundaries'][phase][category]['count']
            if actual != wanted:
                raise AssertionError(f'fixture {phase}/{category}: expected {wanted} live, got {actual}')
    expected_new = {'parse': {'Node': 6, 'Symbol': 2, 'Type': 0, 'Signature': 0},
                    'bind': {'Node': 1, 'Symbol': 1, 'Type': 3, 'Signature': 1},
                    'check': {'Node': 0, 'Symbol': 1, 'Type': 4, 'Signature': 1}}
    for row in result['phase_cohorts']:
        category = row['constructor']
        if category not in expected_allocated: continue
        wanted = expected_new[row['phase']][category]
        if row['observed_new_live_count'] != wanted or row['survives_check_count'] != wanted:
            raise AssertionError(f'fixture {row["phase"]}/{category}: incorrect cohort or survival IDs')
    for category, wanted in expected_allocated.items():
        actual = sum(row['exact_constructor_calls'] or 0 for row in result['phase_cohorts'] if row['constructor'] == category)
        if actual != wanted:
            raise AssertionError(f'fixture {category}: expected {wanted} allocations, got {actual}')
        if result['released'][category]['count'] != 0:
            raise AssertionError(f'fixture {category}: release left live instances')


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('folder', type=Path)
    parser.add_argument('--fixture', action='store_true')
    parser.add_argument('--drop-constructor', choices=CATEGORIES)
    args = parser.parse_args()
    result = analyze(args.folder, args.drop_constructor)
    if args.fixture:
        verify_fixture(result, read_json(args.folder / 'observations.json'))
    name = 'mutant-summary.json' if args.drop_constructor else 'summary.json'
    (args.folder / name).write_text(json.dumps(result, indent=2) + '\n')
    print('PASS fixture allocation/live counts' if args.fixture else 'PASS parsed snapshots and profiles')


if __name__ == '__main__':
    main()
