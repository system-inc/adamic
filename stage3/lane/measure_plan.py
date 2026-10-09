#!/usr/bin/env python3
"""Inventory upstream observations and reject file plans exceeding the unit budget."""
import argparse
from collections import Counter, defaultdict
import gzip
import hashlib
import json
import math
from pathlib import Path


def make_plan(rows, sources, target_seconds=10):
    mapping = {}
    for source in sources:
        if source['title'] in mapping and mapping[source['title']] != source['file']:
            raise ValueError('ambiguous source file for ' + source['title'])
        mapping[source['title']] = source['file']
    files = defaultdict(list)
    seen = set()
    for row in rows:
        task = row['task']
        key = (task['runner'], task['file'])
        if key in seen:
            raise ValueError('duplicate upstream task: ' + str(key))
        seen.add(key)
        file = mapping[task['file']] if task['runner'] == 'unittest' else task['file']
        files[file].append(row)
    weights = {file: sum(row['duration'] for row in group) / 1000
               for file, group in files.items()}
    count = max(1, math.ceil(sum(weights.values()) / target_seconds))
    loads = [0.] * count
    assignment = {}
    for file in sorted(files, key=lambda name: (-weights[name], name)):
        shard = min(range(count), key=lambda index: (loads[index], index))
        assignment[file] = shard
        loads[shard] += weights[file]
    inventory = []
    tasks = []
    for file in sorted(files):
        for row in sorted(files[file], key=lambda row: (row['task']['runner'], row['task']['file'])):
            task = row['task']
            tasks.append(dict(file=file, shard=assignment[file], task=task,
                              seconds=row['duration'] / 1000))
            occurrences = Counter()
            for status, records in [('pass', row['passes']), ('fail', row['errors'])]:
                for record in records:
                    name = record['name']
                    identity = json.dumps([task['runner'], task['file'], name], separators=(',', ':'))
                    occurrence = occurrences[identity]
                    occurrences[identity] += 1
                    inventory.append(dict(id=hashlib.sha256((identity + ':' + str(occurrence)).encode()).hexdigest(),
                                          file=file, shard=assignment[file], task=task,
                                          name=name, status=status))
    if len({row['id'] for row in inventory}) != len(inventory):
        raise ValueError('duplicate upstream test identity')
    oversized = [dict(file=file, seconds=weights[file], shard=assignment[file])
                 for file in sorted(files) if weights[file] >= 30]
    return dict(shards=count, tasks=tasks, tests=inventory, loads=loads, oversized_files=oversized,
                status='invalid' if oversized else 'requires cold measurement',
                counts=dict(passing=sum(row['status'] == 'pass' for row in inventory),
                            failing=sum(row['status'] == 'fail' for row in inventory), pending=0))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('profile', type=Path)
    parser.add_argument('oracle', type=Path)
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    rows = [json.loads(line) for line in args.profile.read_text().splitlines()]
    sources = [json.loads(line) for line in Path(str(args.profile) + '.sources').read_text().splitlines()]
    plan = make_plan(rows, sources)
    observed = json.loads((args.oracle / 'report.json').read_text())
    if plan['counts'] != observed['counts']:
        raise SystemExit('IPC inventory does not equal upstream reporter counts')
    args.output.mkdir(parents=True, exist_ok=False)
    tests = plan.pop('tests')
    with gzip.open(args.output / 'tests-and-shards.jsonl.gz', 'wt') as listing:
        for row in tests:
            listing.write(json.dumps(row, sort_keys=True) + '\n')
    (args.output / 'plan.json').write_text(json.dumps(plan, indent=2) + '\n')
    print(json.dumps(dict(shards=plan['shards'], counts=plan['counts'],
                          oversized_files=plan['oversized_files'], status=plan['status']), indent=2))
    raise SystemExit(1 if plan['oversized_files'] else 0)
