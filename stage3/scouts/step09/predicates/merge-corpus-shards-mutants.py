#!/usr/bin/env python3
"""Corruptions of shard agreement and call coverage must fail independently."""
import copy, importlib.util, json, sys
from pathlib import Path
spec = importlib.util.spec_from_file_location('shard_merge', Path(__file__).with_name('merge-corpus-shards.py'))
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)
reports = [json.loads(Path(p).read_bytes()) for p in sys.argv[1:]]
module.merge(reports)

def missing_shard(r):
    r.pop()
def duplicate_call(r):
    row = next(row for row in r[0]['predicates'] if row['calls'])
    row['calls'].append(copy.deepcopy(row['calls'][0]))
def disagree_on_proof(r):
    r[1]['predicates'][0]['bodyProof'] = not r[1]['predicates'][0]['bodyProof']
def move_call(r):
    next(c for row in r[0]['predicates'] for c in row['calls'])['ordinal'] += 1
def change_diagnostic(r):
    r[1]['checkerDiagnostics'].pop()

mutants = [
    ('drop a shard', missing_shard, 'shard count drift'),
    ('duplicate a call', duplicate_call, 'call coverage drift'),
    ('change one shard body proof', disagree_on_proof, 'body proof drift between shards'),
    ('move a call to another shard', move_call, 'call belongs to another shard'),
    ('drop a checker diagnostic', change_diagnostic, 'checker diagnostic drift between shards'),
]
results = []
for name, mutate, catcher in mutants:
    r = copy.deepcopy(reports)
    mutate(r)
    try:
        module.merge(r)
    except AssertionError as failure:
        assert catcher in str(failure), (name, str(failure))
        results.append({'mutant': name, 'catcher': str(failure)})
    else:
        raise AssertionError('mutant escaped: ' + name)
print(json.dumps(results, indent=2))
