"""Audit the retained deferred-reach ledger without rerunning the compiler."""
import gzip
import hashlib
import importlib.util
import json
from collections import Counter
from pathlib import Path

HERE = Path(__file__).resolve().parent
EVIDENCE = HERE / 'evidence/steps2-3'
rel = lambda s: 'src/' + s.split('/src/', 1)[1] if '/src/' in s else s

def read(name):
    path = EVIDENCE / name
    return gzip.decompress(path.read_bytes()).decode() if name.endswith('.gz') else path.read_text()

def rows(name):
    return [json.loads(line) for line in read(name).splitlines()]

report = json.loads((HERE / 'REPORT.json').read_text())
units = report['units']
head = rows('head.jsonl.gz')
ledger = {rel(u['where']): u for row in head for u in row.get('units', [])
          if u['status'] == 'deferred_generic_instantiation'}
refs = {r['where']: r for r in json.loads(read('classification-references.json.gz'))['units']}
assert len(units) == len(ledger) == len({u['where'] for u in units}) == 504
assert {u['where'] for u in units} == set(ledger)
hidden = "unreached hidden behind a census caller's earlier stop"
sig = 'signature only, no body'
probes = {r['unit']: r for n in ['caller-probe-audit.json.gz', 'caller-probe-extra-audit.json.gz', 'caller-probe-callback-audit.json.gz']
          for r in json.loads(read(n))}
first = {}
for row in head:
    for f in row.get('findings', []):
        if f['phase'] in ('lowering', 'deferred'):
            first.setdefault(rel(f['unit']), f)
requests = {}
for q in rows('head-requests.jsonl.gz'):
    target = rel(q['target'])
    requests.setdefault(target, []).append(q)
    impl = refs.get(target, {}).get('implementation')
    if impl and impl != target and refs[target].get('implementation_generic'):
        requests.setdefault(impl, []).append(q)
trace_events=[json.loads(line) for n in ['caller-probe-stages.jsonl.gz','caller-probe-extra-stages.jsonl.gz','caller-probe-callback-stages.jsonl.gz'] for line in read(n).splitlines()]
bindings={(b['target'], b['passing_call'], b['callee']) for b in json.loads(read('callback-bindings.json.gz'))}
for u in units:
    where = u['where']; raw = ledger[where]; semantic = refs[where]
    assert u['has_body'] == semantic['has_body']
    assert u['raw_instantiated'] == raw['instantiated']
    assert u['instantiated_by'] == raw.get('instantiated_by', [])
    assert u['reference_count'] == len(semantic['references'])
    if not u['has_body']:
        assert u['classification'] == sig
    elif u['classification'] == 'reached':
        assert raw['instantiated'] and u['instantiated_by']
    else:
        assert not raw['instantiated']
    if u['classification'] == 'unreached with no call anywhere':
        assert not semantic['references'] and not semantic['calls']
    if u['classification'] == 'unreached, called only from outside the census':
        assert not u['census_callers'] and u['outside_caller_files']
    if u['classification'] == hidden:
        assert not requests.get(where) and u['census_callers']
        for c in u['census_callers']:
            probe = probes[c['unit']]
            assert probe['first_stop_matches'] and probe['baseline_stop'] == first[c['unit']]
            assert c['diagnostic_replay_matches_first_stop'] and c['stop']
            assert c['stop']['where'] == rel(first[c['unit']]['where'])
            assert c['stop']['reason'] == first[c['unit']]['reason']
            for edge in c['route']:
                assert any(call['call_site'] == edge['call_site'] for call in refs[edge['target']]['calls'])
                if edge.get('callback_callee') and edge.get('external_callback') is False:
                    assert (edge['target'], edge['call_site'], edge['callback_callee']) in bindings
    if u['classification'] is None:
        assert requests.get(where)
        expected=[{k:rel(v) if k in ('unit','declaration') else v for k,v in e.items()} for e in trace_events if rel(e['declaration'])==where and any(rel(e['unit'])==q['unit'] for q in u['requests'])]
        assert u['request_traces']==expected and expected
        if where=='src/compiler/utilities.ts:9031:1':
            assert any(e['stage']=='lowerBody_exit' and not e['error'] for e in expected)
        elif where=='src/compiler/executeCommandLine.ts:981:1':
            assert any(e['stage']=='lowerBody_entry' for e in expected)
            assert any(e['stage']=='lowerBody_exit' and e['error'] for e in expected)
        else:
            assert not any(e['stage']=='lowerBody_entry' for e in expected)
            assert any(e['stage']=='latentMainSignatureRaw_exit' and e['error'] for e in expected)
all_counts = Counter(u['classification'] for u in units)
core = [u for u in units if u['where'].startswith('src/compiler/core.ts:')]
assert all_counts == Counter({'reached':111, sig:137, hidden:223,
    'unreached with no call anywhere':2, 'unreached, called only from outside the census':28, None:3})
assert Counter(u['classification'] for u in core) == Counter({'reached':61,sig:84,hidden:44,
    'unreached, called only from outside the census':18})
assert sum(u['has_body'] for u in core) == 123
assert sum(u['has_body'] for u in units) == 367
assert sum(u['raw_instantiated'] for u in units) == 115
assert len(report['caller_stop_ledger']) == 112
# Recheck field-only byte identity using retained baseline files from Step 1.
spec = importlib.util.spec_from_file_location('witness_identity', HERE.parent/'latent/audit_instantiated_by.py')
identity = importlib.util.module_from_spec(spec); spec.loader.exec_module(identity)
for scope, filename in [('directory', 'head.jsonl.gz'), ('entry', 'head-entry.jsonl.gz')]:
    stripped, count = identity.strip_witnesses(read(filename).encode())
    saved = json.loads(read('corpus-identity.json'))[scope]
    assert count == 115
    assert hashlib.sha256(stripped).hexdigest() == saved['base_sha256'] == saved['stripped_sha256']
if (EVIDENCE/'full-core-callers-result.json').exists():
    attempt=json.loads(read('full-core-callers-result.json'))
    assert report['bounded_full_caller_attempt'] == attempt
    assert attempt['child_exit'] == int(read('full-core-callers.exit'))
    assert len(attempt['selected_files']) == 32
for name, digest in json.loads(read('SHA256.json')).items():
    assert hashlib.sha256((EVIDENCE/name).read_bytes()).hexdigest() == digest, name
print('PASS: 504 rows, 137 signatures, 367 bodies, 223 hidden bodies, 112 replayed caller stops; 3 explicit unclassified requests')
