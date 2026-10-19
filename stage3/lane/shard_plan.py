"""Generate deterministic gate units from the measured whole-run census."""
from collections import Counter, defaultdict
import gzip
import hashlib
import json
import math
from pathlib import Path
import re

LANE = Path(__file__).resolve().parent
WATCHER = 'src/testRunner/unittests/sys/symlinkWatching.ts'
PUBLIC_API = 'src/testRunner/unittests/publicApi.ts'
# These whole suites exceeded nine seconds in the product-path cold probe.
# Keep repeated full titles together: upstream grep selects every occurrence.
CASE_PARTITIONS = {
    'src/testRunner/unittests/tsserver/projectReferencesSourcemap.ts': 'project-reference-cases',
    'src/testRunner/unittests/services/convertToAsyncFunction.ts': 'async-function-cases',
    'src/testRunner/unittests/tsbuild/noEmit.ts': 'build-no-emit-cases',
    'src/testRunner/unittests/tsc/noEmit.ts': 'compiler-no-emit-cases',
}
# A compilation and its clean-build comparison share state within one describe.
PAIRED_CASE_FILES = {
    'src/testRunner/unittests/tsbuild/noEmit.ts',
    'src/testRunner/unittests/tsc/noEmit.ts',
}
CASE_GROUPS = 3
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


def file_assignment(tasks, target_seconds=9.0):
    # The retained cold ledger includes four concurrent commands. Use its worst
    # wall as a conservative calibration, then prove the new walls serially.
    ledger_path = LANE / 'evidence/case-shards/cold-ledger.json'
    ledger = json.loads(ledger_path.read_text())
    maximum = max(row['seconds'] for row in ledger['units'])
    weight_target = 10.0 * target_seconds / maximum
    files = defaultdict(float)
    for row in tasks:
        if row['file'] not in (WATCHER, PUBLIC_API) and row['file'] not in CASE_PARTITIONS:
            files[row['file']] += row['seconds']
    count = math.ceil(sum(files.values()) / weight_target)
    loads = [0.] * count
    owners = {}
    for file in sorted(files, key=lambda file: (-files[file], file)):
        index = min(range(count), key=lambda index: (loads[index], index))
        owners[file] = f'stage3-lane-file-{index:03d}'
        loads[index] += files[file]
    return owners, dict(target_seconds=target_seconds, weight_target_seconds=weight_target,
                        calibration_maximum_seconds=maximum,
                        case_groups={file: CASE_GROUPS for file in CASE_PARTITIONS},
                        calibration_sha256=hashlib.sha256(ledger_path.read_bytes()).hexdigest())


def build_plan(artifact, api_digest):
    evidence = LANE / 'evidence/shards'
    with gzip.open(evidence / 'candidate-plan.json.gz', 'rt') as stream:
        old = json.load(stream)
    with gzip.open(evidence / 'tests-and-shards.jsonl.gz', 'rt') as stream:
        tests = [json.loads(line) for line in stream]
    assignments, planning = file_assignment(old['tasks'])
    units = {}
    for row in old['tasks']:
        if row['file'] == WATCHER or row['file'] in CASE_PARTITIONS:
            continue
        name = 'stage3-lane-byte-comparisons' if row['file'] == PUBLIC_API else assignments[row['file']]
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
    for file, label in CASE_PARTITIONS.items():
        tasks = [row['task'] for row in old['tasks'] if row['file'] == file]
        if len(tasks) != 1:
            raise ValueError('case partition requires exactly one upstream task: ' + file)
        titles = defaultdict(list)
        for row in tests:
            if row['file'] == file:
                family = row['name'][:-1] if file in PAIRED_CASE_FILES else row['name']
                titles[title(family)].append(row)
        if len(titles) < CASE_GROUPS:
            raise ValueError('case partition has too few selectable titles: ' + file)
        ordered = sorted(titles)
        for index in range(CASE_GROUPS):
            selected = ordered[index::CASE_GROUPS]
            name = f'stage3-lane-{label}-{index:02d}'
            names = {title(row['name']): row['name'] for text in selected for row in titles[text]}
            units[name] = dict(name=name, tasks=tasks, tests=[],
                               grep='|'.join(exact_grep(names[text]) for text in sorted(names)))
            for text in selected:
                for row in titles[text]:
                    owners[row['id']] = name
    for row in tests:
        name = owners[row['id']] if row['id'] in owners else ('stage3-lane-byte-comparisons'
               if row['file'] == PUBLIC_API else assignments[row['file']])
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
                planning=planning, counts={'pass': 106366, 'fail': 1, 'skip': 0}, units=sorted(units.values(), key=lambda row: row['name'])), tests


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
