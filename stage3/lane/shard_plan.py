"""Generate deterministic gate units from the measured whole-run census."""
from collections import Counter
import gzip
import hashlib
import json
from pathlib import Path
import re

LANE = Path(__file__).resolve().parent
WATCHER = 'src/testRunner/unittests/sys/symlinkWatching.ts'
PUBLIC_API = 'src/testRunner/unittests/publicApi.ts'
RUNTIME = ['run.py', '../oracle/run.py', '../oracle/profile-tasks.cjs', 'artifact.py', 'view.py', 'shard_plan.py', 'shard.py', 'merge_shards.py',
           'prepare_shards.py', 'check.py', 'normalize-api.cjs', 'expected.json',
           'sanctioned-api.json', '../oracle/run-prepared.mjs', '../oracle/run-tasks.cjs', '../oracle/observe-tests.cjs',
           '../oracle/observe-errors.cjs', '../oracle/failure_details.py']


def digest(value):
    return hashlib.sha256(json.dumps(value, sort_keys=True, separators=(',', ':')).encode()).hexdigest()


def runtime_hash():
    return digest({name: hashlib.sha256((LANE / name).read_bytes()).hexdigest() for name in RUNTIME})


def title(name):
    return ' '.join(part for part in name if part)


def names_digest(rows):
    # Keep the original task identity as well as the full describe/it title.
    counts = Counter((row['task']['runner'], row['task']['file'], title(row['name'])) for row in rows)
    return digest(sorted((*key, count) for key, count in counts.items()))


def exact_grep(name):
    return '^' + re.escape(title(name)) + '$'


def build_plan(artifact, api_digest):
    evidence = LANE / 'evidence/shards'
    with gzip.open(evidence / 'candidate-plan.json.gz', 'rt') as stream:
        old = json.load(stream)
    with gzip.open(evidence / 'tests-and-shards.jsonl.gz', 'rt') as stream:
        tests = [json.loads(line) for line in stream]
    units = {}
    for row in old['tasks']:
        if row['file'] == WATCHER:
            continue
        name = 'stage3-lane-byte-comparisons' if row['file'] == PUBLIC_API else f"stage3-lane-file-{row['shard']:03d}"
        unit = units.setdefault(name, dict(name=name, tasks=[], tests=[], measured_seconds=0))
        unit['tasks'].append(row['task'])
        unit['measured_seconds'] += row['seconds']
    watcher_tests = sorted((row for row in tests if row['file'] == WATCHER), key=lambda row: title(row['name']))
    if len({title(row['name']) for row in watcher_tests}) != len(watcher_tests):
        raise ValueError('watcher titles are not uniquely selectable by upstream grep')
    owners = {}
    for index, row in enumerate(watcher_tests):
        name = f'stage3-lane-symlink-case-{index:02d}'
        units[name] = dict(name=name, tasks=[row['task']], tests=[], grep=exact_grep(row['name']))
        owners[row['id']] = name
    for row in tests:
        name = owners[row['id']] if row['file'] == WATCHER else ('stage3-lane-byte-comparisons'
               if row['file'] == PUBLIC_API else f"stage3-lane-file-{row['shard']:03d}")
        units[name]['tests'].append(row)
        row['unit'] = name
    if len({row['id'] for row in tests}) != len(tests):
        raise ValueError('duplicate upstream identity in partition')
    runtime = runtime_hash()
    for unit in units.values():
        records = unit.pop('tests')
        unit['expect'] = dict(pass_=sum(row['status'] == 'pass' for row in records),
                              fail=sum(row['status'] == 'fail' for row in records), skip=0)
        unit['expect']['pass'] = unit['expect'].pop('pass_')
        unit['names_digest'] = names_digest(records)
        unit['baseline_diffs'] = ['api/typescript.d.ts'] if unit['name'] == 'stage3-lane-byte-comparisons' else []
        unit['inputs_hash'] = digest(dict(artifact=artifact, api=api_digest, runtime=runtime, unit=unit))
    return dict(version=1, artifact=artifact, api_digest=api_digest, runtime=runtime,
                counts={'pass': 106366, 'fail': 1, 'skip': 0}, units=sorted(units.values(), key=lambda row: row['name'])), tests


def gate_manifest(plan):
    return dict(version=1, units=[dict(name=unit['name'], command=[
        'python3', 'stage3/lane/shard.py', unit['name'], '$STAGE3_RESULTS',
        '--inputs-hash', unit['inputs_hash']], expect=unit['expect']) for unit in plan['units']])


def load_plan():
    return json.loads((LANE / 'shard-plan.json').read_text())


if __name__ == '__main__':
    import argparse
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('artifact', type=Path)
    args = parser.parse_args()
    ready = json.loads((args.artifact / 'ready.json').read_text())
    api = json.loads((args.artifact / 'shards-ready.json').read_text())
    plan, tests = build_plan(ready['key'], api['api_digest'])
    (LANE / 'shard-plan.json').write_text(json.dumps(plan, indent=2) + '\n')
    (LANE / 'shards.json').write_text(json.dumps(gate_manifest(plan), indent=2) + '\n')
    listing = LANE / 'evidence/case-shards'
    listing.mkdir(parents=True, exist_ok=True)
    with gzip.open(listing / 'tests-and-units.jsonl.gz', 'wt') as stream:
        for row in tests:
            stream.write(json.dumps(row, sort_keys=True) + '\n')
    print(json.dumps(dict(units=len(plan['units']), counts=plan['counts'])))
