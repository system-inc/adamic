#!/usr/bin/env python3
"""Independent corruptions must fail the pinned corpus audit."""
import copy, importlib.util, json, sys
from pathlib import Path
spec = importlib.util.spec_from_file_location('corpus_audit', Path(__file__).with_name('audit-corpus.py'))
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)
inventory, grouped, measured = map(module.read, sys.argv[1:4])
tree = Path(sys.argv[4])
module.audit(inventory, grouped, measured, tree)

def missing_body(i, g, m):
    m['predicates'].pop()
def missing_call(i, g, m):
    next(r for r in m["predicates"] if r["calls"])["calls"].pop()
def invent_proof(i, g, m):
    r = next(r for r in m['predicates'] if not r['bodyProof'])
    r['admission'] = 'Proven'
def pending_as_checked(i, g, m):
    c = next(c for r in m['predicates'] for c in r['calls'] if c['status'] == 'Pending')
    c['status'] = 'CheckedSeam'
    c['directions'] = [{'Status': 'checked', 'Direction': 'true', 'Reason': 'invented'}]
def pending_view_as_proven(i, g, m):
    r = next(r for r in m['predicates'] if r['admission'] == 'PendingView')
    r['admission'] = 'Proven'
def change_source_hash(i, g, m):
    name = next(iter(i['hashes']))
    i['hashes'][name] = '0' * 64

mutants = [
    ('drop measured body', missing_body, 'measurement body coverage drift'),
    ('drop measured call', missing_call, 'call coverage drift'),
    ('invent body proof', invent_proof, 'proof admission contradiction'),
    ('label failed call checked', pending_as_checked, 'failed call labelled checked'),
    ('label pending view proven', pending_view_as_proven, 'proof admission contradiction'),
    ('change pinned source hash', change_source_hash, 'source hash drift'),
]
results = []
for name, mutate, catcher in mutants:
    i, g, m = copy.deepcopy((inventory, grouped, measured))
    mutate(i, g, m)
    try:
        module.audit(i, g, m, tree)
    except AssertionError as failure:
        assert catcher in str(failure), (name, str(failure))
        results.append({'mutant': name, 'catcher': str(failure)})
    else:
        raise AssertionError('mutant escaped: ' + name)
print(json.dumps(results, indent=2))
