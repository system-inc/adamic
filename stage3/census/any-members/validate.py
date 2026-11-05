"""Independent arithmetic and provenance checks over the saved census artifacts."""
import collections
import gzip
import hashlib
import json
import sys
from pathlib import Path

out = Path(__file__).resolve().parent
e = out / 'evidence'
delta = json.loads((out / 'delta.json').read_text())
def observed(label, suffix):
    prefix = delta.get('head_evidence_prefix', 'head') if label == 'head' else label
    return e / f'{prefix}-{suffix}'
if delta.get('status') == 'incomplete':
    streams = {}
    timers = {}
    for label in ('base', 'head'):
        with gzip.open(observed(label, 'latent.jsonl.gz'), 'rt') as f:
            streams[label] = [json.loads(line) for line in f]
        with gzip.open(observed(label, 'times.jsonl.gz'), 'rt') as f:
            timers[label] = [json.loads(line) for line in f]
        assert streams[label][0]['latent_mode'] == 'first-error'
        for key in ('wall_seconds', 'cpu_seconds'):
            assert sum(t[key] for t in timers[label]) == delta['unit_time_sums'][label][key]
        pins = json.loads((e / f'{label}-actual-pins.json').read_text())
        assert pins == {'.': delta[f'{label}_commit'], 'cohere': '06af2e399c6c1fb93fd570b774027942747400b4', 'cohere/TypeScript': 'd92d9bfee114c80be2c375d72edae966176e3a4f'}
        assert not any(json.loads(observed(label, 'post-run-status.json').read_text()).values())
    assert streams['base'][0] == streams['head'][0]
    assert len(streams['base']) == 80 and len(streams['head']) == 1
    assert len(timers['base']) == 4056 and len(timers['head']) == 3153
    assert delta['foreground_run']['source_hashes_matched'] == 746
    assert delta['foreground_run']['compiler_commit'] == delta['head_commit']
    assert 'foreground' in delta['foreground_run']['launch']
    assert 'returncode' not in delta['foreground_run']
    assert delta['interruption']['state'] == 'Z'
    assert delta['prior_interrupted_attempt']['head_completed_timers'] == 3865
    assert [r['where'] for r in timers['head']] == [r['where'] for r in timers['base'][:3153]]
    unique = {tuple(f[k] for k in ('kind', 'where', 'reason', 'text')) for row in streams['base'][1:] for f in row['findings']}
    counter = collections.Counter(key[0] for key in unique)
    assert delta['counts']['base'] == {kind: counter[kind] for kind in ('NotYet', 'Refused', 'panic', 'error')}
    assert delta['counts']['head'] is None
    for key in ('any_member_findings', 'any_member_unit_first_stops', 'regressions', 'new_panics', 'unit_first_stop_delta'):
        assert delta[key] is None
    assert delta['pairing_succeeded'] is False
    assert delta['pairing_error'] == 'unit identities differ'
    with gzip.open(e / 'rerun-head-pair-error.log.gz', 'rt') as f:
        assert 'unit identities differ' in f.read()
    assert not observed('head', 'exit.txt').exists() and not observed('head', 'end.txt').exists()
    assert (e / 'base-exit.txt').read_text().strip() == '0'
    assert (e / 'base-memory-before.txt').read_bytes() == (e / 'base-memory-after.txt').read_bytes()
    stock = json.loads((e / 'stock-units.json').read_text())
    expected = {u['where'] for row in stock for u in row['units'] if not u['deferred_generic']}
    assert expected == {r['where'] for r in timers['base']}
    assert {row['file'] for row in stock} == {row['file'] for row in streams['base'][1:]}
    tree = out / 'scratch/adapted'
    manifest = json.loads((e / 'source-manifest.json').read_text())
    assert manifest == {str(p.relative_to(tree)): hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree / 'src').rglob('*')) if p.is_file()}
    for name, digest in json.loads((out / 'artifacts.sha256.json').read_text()).items():
        assert hashlib.sha256((out / name).read_bytes()).hexdigest() == digest, name
    print('PASS: complete base counts and stock-unit coverage; matching first-error headers; interrupted foreground head timer prefix and partial sums; preserved first attempt; explicit unknown answers; rejected pairing; exit evidence; source hashes; clean compiler/dependency pins; artifact hashes.')
    sys.exit(0)
streams = {}
for side, label in (('before', 'base'), ('after', 'head')):
    with gzip.open(observed(label, 'latent.jsonl.gz'), 'rt') as f:
        rows = [json.loads(line) for line in f]
    streams[label] = rows
    assert rows[0]['latent_mode'] == 'first-error'
    unique = {tuple(f[k] for k in ('kind', 'where', 'reason', 'text')) for row in rows[1:] for f in row['findings']}
    counts = dict(collections.Counter(k[0] for k in unique))
    assert counts == delta['raw_findings']['totals'][side]
    with gzip.open(observed(label, 'times.jsonl.gz'), 'rt') as f:
        timers = [json.loads(line) for line in f]
    assert len(timers) == delta['unit_time_sample_counts'][label]
    assert len({t['where'] for t in timers}) == len(timers)
    for key in ('wall_seconds', 'cpu_seconds'):
        assert sum(t[key] for t in timers) == delta['unit_time_sums'][label][key]
    assert observed(label, 'exit.txt').read_text().strip() == '0'
    assert observed(label, 'memory-before.txt').read_bytes() == observed(label, 'memory-after.txt').read_bytes()
    assert not (e / f'{label}-status.txt').read_text()
assert streams['base'][0] == streams['head'][0]
assert len(streams['base']) == len(streams['head'])
first_by_label = {}
for label, rows in streams.items():
    stops = {}
    for row in rows[1:]:
        for f in row['findings']:
            if f['phase'] == 'lowering':
                stops.setdefault(f['unit'], (f['where'], f['kind'], f['reason']))
    first_by_label[label] = stops
all_units = {u['where'] for row in streams['base'][1:] for u in row['units']}
changed = {u for u in all_units if first_by_label['base'].get(u) != first_by_label['head'].get(u)}
assert changed == {r['unit']['where'] for r in delta['raw_findings']['unit_first_stop_delta']}
first = {}
for row in streams['head'][1:]:
    for f in row['findings']:
        if f['phase'] == 'lowering':
            first.setdefault(f['unit'], f)
expected = {tuple(f[k] for k in ('kind', 'where', 'reason', 'text', 'unit', 'phase')) for row in streams['head'][1:] for f in row['findings'] if f['reason'].startswith('a member typed any')}
actual = {tuple(r['finding'][k] for k in ('kind', 'where', 'reason', 'text', 'unit', 'phase')) for r in delta['any_member_findings']}
assert expected == actual
for r in delta['any_member_findings']:
    f = r['finding']
    assert r['is_unit_first_stop'] == (f['phase'] == 'lowering' and first.get(f['unit']) == f)
manifest = json.loads((e / 'source-manifest.json').read_text())
tree = out / 'scratch/adapted'
assert manifest == {str(p.relative_to(tree)): hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree / 'src').rglob('*')) if p.is_file()}
for label, commit in (('base', delta['base_commit']), ('head', delta['head_commit'])):
    pins = json.loads((e / f'{label}-actual-pins.json').read_text())
    assert pins == {'.': commit, 'cohere': '06af2e399c6c1fb93fd570b774027942747400b4', 'cohere/TypeScript': 'd92d9bfee114c80be2c375d72edae966176e3a4f'}
stock = json.loads((e / 'stock-units.json').read_text())
stock_attempts = {u['where'] for row in stock for u in row['units'] if not u['deferred_generic']}
if not streams['base'][0]['checker_rejected']:
    for label, rows in streams.items():
        attempted = {u['where'] for row in rows[1:] for u in row['units'] if u['status'] in ('attempted', 'panic')}
        assert attempted == stock_attempts
print('PASS: unique kind counts, timing sums and attempt counts, first-error headers, complete any-member ownership and first-stop flags, exit/status/memory records, dependency pins, and unchanged adapted source hashes.')
